package model

import (
	"encoding/json"
	"fmt"
	"github.com/doganarif/llmdog/internal/bookmarks"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/doganarif/llmdog/internal/git"
	"github.com/doganarif/llmdog/internal/ui"
)

// Config holds user configuration
type Config struct {
	ShowHiddenFiles   bool    `json:"showHiddenFiles"`
	FuzzyThreshold    float64 `json:"fuzzyThreshold"`
	MaxPreviewSize    int     `json:"maxPreviewSize"`
	ColorTheme        string  `json:"colorTheme"`
	ContentSearchMode bool    `json:"contentSearchMode"`
}

// LoadConfig loads configuration from file or creates default
func LoadConfig() (Config, error) {
	config := Config{
		ShowHiddenFiles:   false,
		FuzzyThreshold:    0.6,
		MaxPreviewSize:    10000,
		ColorTheme:        "default",
		ContentSearchMode: false,
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "llmdog")
	configPath := filepath.Join(configDir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// If file doesn't exist, create default config
		if os.IsNotExist(err) {
			os.MkdirAll(configDir, 0755)
			saveConfig(config, configPath)
			return config, nil
		}
		return config, err
	}

	err = json.Unmarshal(data, &config)
	return config, err
}

// saveConfig saves configuration to file
func saveConfig(config Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Custom messages
type errMsg struct{ err error }
type successMsg struct{ message string }
type loadingMsg struct{ done bool }
type childrenLoadedMsg struct {
	parentPath string
	children   []ui.FileItem
}
type customSearchMsg struct {
	query string
}
type resetViewMsg struct{}

// Model represents the application state
type Model struct {
	list                list.Model
	preview             string
	items               []ui.FileItem
	cwd                 string
	gitignoreRegexp     *regexp.Regexp
	termWidth           int
	termHeight          int
	showPreview         bool
	isLoading           bool
	loadingMessage      string
	spinner             spinner.Model
	errors              []string
	showErrors          bool
	searchHistory       []string
	searchHistoryIndex  int
	fuzzyThreshold      float64
	contentSearchMode   bool
	selectedCount       int
	selectedSize        int64
	estimatedTokens     int
	config              Config
	statusMessage       string
	statusMessageExpiry time.Time
	lock                sync.Mutex
	isInSearchResults   bool
	bookmarkStore       bookmarks.BookmarkStore
	showBookmarksMenu   bool
	bookmarksMenu       ui.BookmarksMenu
	textInputModal      ui.TextInputModal
	showTextInputModal  bool
	textInputPurpose    string
	tempBookmarkName    string
}

// New creates a new model
func New() *Model {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Load config
	config, err := LoadConfig()
	if err != nil {
		log.Printf("Warning: Could not load config: %v", err)
	}

	gitRegex, _ := git.ParseGitignore(filepath.Join(cwd, ".gitignore"))
	items := ui.LoadFiles(cwd, gitRegex, config.ShowHiddenFiles)

	// Only include top-level items initially since folders are collapsed
	var listItems []list.Item
	for _, item := range items {
		if item.Depth == 0 { // Only include root level items
			listItems = append(listItems, item)
		}
	}

	l := list.New(listItems, ui.ItemDelegate{}, 30, 14)
	l.Title = " Files  |  ↑↓:navigate  •  Space:expand/collapse folder •  Tab:select  •  /:filter  •  Enter:confirm  •  q:quit "
	l.SetFilteringEnabled(true)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	bookmarkStore, err := bookmarks.LoadBookmarks()
	if err != nil {
		log.Printf("Warning: Could not load bookmarks: %v", err)
	}

	return &Model{
		list:               l,
		items:              items,
		cwd:                cwd,
		gitignoreRegexp:    gitRegex,
		showPreview:        true,
		spinner:            s,
		fuzzyThreshold:     config.FuzzyThreshold,
		contentSearchMode:  config.ContentSearchMode,
		config:             config,
		bookmarkStore:      bookmarkStore,
		showBookmarksMenu:  false,
		showTextInputModal: false,
	}
}

// addError adds an error to the error list
func (m *Model) addError(err error) {
	if err != nil {
		m.errors = append(m.errors, err.Error())
		m.showErrors = true
	}
}

// setStatusMessage sets a temporary status message
func (m *Model) setStatusMessage(message string, durationSecs int) {
	m.statusMessage = message
	m.statusMessageExpiry = time.Now().Add(time.Duration(durationSecs) * time.Second)
}

// isGitIgnored checks if a path is git ignored
func (m *Model) isGitIgnored(path string) bool {
	return m.gitignoreRegexp != nil && m.gitignoreRegexp.MatchString(path)
}

// getDirectChildren returns the direct children of a path
func (m *Model) getDirectChildren(parentPath string) []ui.FileItem {
	var children []ui.FileItem
	for i := range m.items {
		if filepath.Dir(m.items[i].Path) == parentPath {
			children = append(children, m.items[i])
		}
	}
	return children
}

// isVisible determines if an item should be visible
func (m *Model) isVisible(item ui.FileItem) bool {
	if item.Depth == 0 {
		return true
	}

	parentPath := filepath.Dir(item.Path)
	for parentPath != m.cwd && parentPath != "." {
		found := false
		for i := range m.items {
			if m.items[i].Path == parentPath && m.items[i].IsDir {
				if !m.items[i].Expanded {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false
		}
		parentPath = filepath.Dir(parentPath)
	}
	return true
}

// refreshVisibleItems updates the list of visible items
func (m *Model) refreshVisibleItems() {
	m.lock.Lock()
	defer m.lock.Unlock()

	visible := make([]list.Item, 0)
	selectedIndex := m.list.Index()
	var selectedPath string
	if sel, ok := m.list.SelectedItem().(ui.FileItem); ok {
		selectedPath = sel.Path
	}

	// Track all selected items before refresh
	selectedItems := make(map[string]bool)
	for _, item := range m.items {
		if item.Selected {
			selectedItems[item.Path] = true
		}
	}

	for i := range m.items {
		if m.isVisible(m.items[i]) {
			// Ensure selection state is preserved
			if _, ok := selectedItems[m.items[i].Path]; ok {
				m.items[i].Selected = true
			}
			visible = append(visible, m.items[i])
		}
	}

	m.list.SetItems(visible)

	// Restore cursor position
	if selectedPath != "" {
		for i, item := range visible {
			if fileItem, ok := item.(ui.FileItem); ok && fileItem.Path == selectedPath {
				m.list.Select(i)
				break
			}
		}
	} else if selectedIndex < len(visible) {
		m.list.Select(selectedIndex)
	}

	// Refresh selection statistics
	m.refreshSelectionStats()
}

// refreshSelectionStats updates statistics about selected items
func (m *Model) refreshSelectionStats() {
	m.selectedCount = 0
	m.selectedSize = 0
	m.estimatedTokens = 0

	for _, item := range m.items {
		if item.Selected && !item.IsDir && !m.isGitIgnored(item.Path) {
			m.selectedCount++

			// Get file size
			info, err := os.Stat(item.Path)
			if err == nil {
				m.selectedSize += info.Size()

				// Estimate tokens (very rough approximation)
				// Assuming 4 characters per token on average
				m.estimatedTokens += int(info.Size()) / 4
			}
		}
	}
}

// getAllDescendants returns all descendants of a path
func (m *Model) getAllDescendants(parentPath string) []ui.FileItem {
	var descendants []ui.FileItem
	parentWithSep := parentPath + string(os.PathSeparator)
	for i := range m.items {
		if strings.HasPrefix(m.items[i].Path, parentWithSep) {
			descendants = append(descendants, m.items[i])
		}
	}
	return descendants
}

// areAllDescendantsSelected checks if all non-gitignored descendants are selected
func (m *Model) areAllDescendantsSelected(parentPath string) bool {
	descendants := m.getAllDescendants(parentPath)
	if len(descendants) == 0 {
		return false
	}

	for _, desc := range descendants {
		if m.isGitIgnored(desc.Path) {
			continue // Skip gitignored items
		}
		for i := range m.items {
			if m.items[i].Path == desc.Path {
				if !m.items[i].Selected {
					return false
				}
				break
			}
		}
	}
	return true
}

// setSelectionStateForDescendants sets selection state for all descendants
func (m *Model) setSelectionStateForDescendants(parentPath string, selected bool) {
	// Update all descendants
	for i := range m.items {
		if strings.HasPrefix(m.items[i].Path, parentPath+string(os.PathSeparator)) {
			if !m.isGitIgnored(m.items[i].Path) {
				m.items[i].Selected = selected
			}
		}
	}
}

// updateParentSelectionState updates a parent's selection state based on children
func (m *Model) updateParentSelectionState(childPath string) {
	parentPath := filepath.Dir(childPath)
	if parentPath == m.cwd {
		return
	}

	for i := range m.items {
		if m.items[i].Path == parentPath && m.items[i].IsDir {
			m.items[i].Selected = m.areAllDescendantsSelected(parentPath)
			// Recursively update parent directories
			m.updateParentSelectionState(parentPath)
			break
		}
	}
}

// toggleExpansion expands or collapses a directory
func (m *Model) toggleExpansion(path string) tea.Cmd {
	var currentItem *ui.FileItem
	var cmds []tea.Cmd

	for i := range m.items {
		if m.items[i].Path == path {
			if m.items[i].IsDir {
				m.items[i].Expanded = !m.items[i].Expanded
				currentItem = &m.items[i]

				// If expanding and no children loaded yet, load them
				if m.items[i].Expanded && !m.items[i].ChildrenLoaded {
					m.isLoading = true
					m.loadingMessage = "Loading directory..."

					// Return a command instead of using a goroutine directly
					cmds = append(cmds, func() tea.Msg {
						children, err := ui.LoadDirectoryChildren(path, m.gitignoreRegexp, m.config.ShowHiddenFiles)
						if err != nil {
							return errMsg{err}
						}
						return childrenLoadedMsg{
							parentPath: path,
							children:   children,
						}
					})
				}
			}
			break
		}
	}

	if currentItem != nil {
		m.refreshVisibleItems()
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

// toggleSelection toggles selection state for an item
func (m *Model) toggleSelection(path string, forceSelect ...bool) {
	// Find the item
	var currentItem *ui.FileItem
	for i := range m.items {
		if m.items[i].Path == path {
			currentItem = &m.items[i]
			break
		}
	}

	if currentItem == nil || m.isGitIgnored(currentItem.Path) {
		return
	}

	// Handle force selection if provided
	force := false
	forceValue := false
	if len(forceSelect) > 0 {
		force = true
		forceValue = forceSelect[0]
	}

	if currentItem.IsDir {
		if (currentItem.Selected && !force) || (force && !forceValue) {
			// Unselect directory and all descendants
			currentItem.Selected = false
			m.setSelectionStateForDescendants(currentItem.Path, false)
		} else {
			// Select directory and all non-gitignored descendants
			currentItem.Selected = true
			m.setSelectionStateForDescendants(currentItem.Path, true)
		}
	} else {
		// Toggle file selection
		if force {
			currentItem.Selected = forceValue
		} else {
			currentItem.Selected = !currentItem.Selected
		}
	}

	// Update parent directory selection states
	m.updateParentSelectionState(path)
	m.refreshVisibleItems()
}

// performSearch executes a search based on current search mode
func (m *Model) performSearch(query string) {
	// If no query, show all visible items
	if query == "" {
		m.refreshVisibleItems()
		return
	}

	results := make([]list.Item, 0)
	foundPaths := make(map[string]bool)

	// Reset content match flags
	for i := range m.items {
		m.items[i].MatchesContent = false
	}

	// Process query to lowercase for case-insensitive matching
	queryLower := strings.ToLower(query)

	// Search through ALL files, regardless of their visibility state
	for i := range m.items {
		matched := false

		// Filename search - simple contains for now
		if strings.Contains(strings.ToLower(m.items[i].Name), queryLower) {
			matched = true
		}

		// Content search if enabled and not a directory
		if !matched && m.contentSearchMode && !m.items[i].IsDir {
			// Only attempt to read small files to avoid performance issues
			info, err := os.Stat(m.items[i].Path)
			if err == nil && info.Size() < 1024*1024 { // Skip files larger than 1MB
				content, err := os.ReadFile(m.items[i].Path)
				if err == nil && strings.Contains(strings.ToLower(string(content)), queryLower) {
					matched = true
					m.items[i].MatchesContent = true // Flag for UI highlight
				}
			}
		}

		if matched {
			// Add the item itself to results
			foundPaths[m.items[i].Path] = true

			// Make sure all parent directories are expanded to make this item visible
			m.ensureParentPathsExpanded(m.items[i].Path)

			// Add this item to results
			results = append(results, m.items[i])
		}
	}

	// Now add all necessary parent directories to make the hierarchy visible
	for i := range m.items {
		if foundPaths[m.items[i].Path] {
			// This item is already in the results
			continue
		}

		// Check if this is a parent directory of any matched item
		if m.items[i].IsDir {
			for path := range foundPaths {
				// Check if this directory is a parent of any matched file
				if strings.HasPrefix(path, m.items[i].Path+string(os.PathSeparator)) {
					// This is a parent directory, add it to results if not already there
					if !foundPaths[m.items[i].Path] {
						foundPaths[m.items[i].Path] = true
						results = append(results, m.items[i])
					}
					break
				}
			}
		}
	}

	// Sort results to maintain hierarchy
	sort.Slice(results, func(i, j int) bool {
		itemI, _ := results[i].(ui.FileItem)
		itemJ, _ := results[j].(ui.FileItem)
		return itemI.Path < itemJ.Path
	})

	// If we have results, show them
	if len(results) > 0 {
		m.list.SetItems(results)
		// Set status message with count
		m.setStatusMessage(fmt.Sprintf("Found %d matches", len(foundPaths)), 2)
	} else if m.contentSearchMode {
		// If no results with content search, show a message
		m.setStatusMessage("No matches found. Try different search terms.", 2)
		// Maintain current view
	} else {
		// If no results with filename search only, show a message
		m.setStatusMessage("No filename matches. Try content search (Ctrl+S).", 2)
		// Maintain current view
	}
}

// ensureParentPathsExpanded makes sure all parent directories of a path are expanded
func (m *Model) ensureParentPathsExpanded(path string) {
	dir := filepath.Dir(path)

	// If we've reached the root, stop
	if dir == m.cwd || dir == "." {
		return
	}

	// Recursively process parent directories first
	m.ensureParentPathsExpanded(dir)

	// Then expand this directory
	for i := range m.items {
		if m.items[i].Path == dir && m.items[i].IsDir {
			// Ensure this directory is expanded
			m.items[i].Expanded = true

			// If children aren't loaded yet, load them synchronously
			if !m.items[i].ChildrenLoaded {
				children, err := ui.LoadDirectoryChildren(dir, m.gitignoreRegexp, m.config.ShowHiddenFiles)
				if err == nil {
					// Check for duplicates before adding
					existingPaths := make(map[string]bool)
					for _, item := range m.items {
						existingPaths[item.Path] = true
					}

					for _, child := range children {
						if !existingPaths[child.Path] {
							m.items = append(m.items, child)
						}
					}

					m.items[i].ChildrenLoaded = true
				}
			}
			break
		}
	}
}

// ensureParentDirsExpanded ensures all parent directories are expanded
// and adds them to results for display
func (m *Model) ensureParentDirsExpanded(path string, results *[]list.Item, foundPaths *map[string]bool) {
	parentPath := filepath.Dir(path)
	if parentPath == m.cwd || parentPath == "." {
		return
	}

	// Recursively process parents first
	m.ensureParentDirsExpanded(parentPath, results, foundPaths)

	// Then add this parent if not already included
	if !(*foundPaths)[parentPath] {
		for i := range m.items {
			if m.items[i].Path == parentPath && m.items[i].IsDir {
				// Mark directory as expanded
				m.items[i].Expanded = true

				// Add it to results if not already there
				(*foundPaths)[parentPath] = true
				*results = append(*results, m.items[i])
				break
			}
		}
	}
}

// selectAll selects all visible items
func (m *Model) selectAll() {
	for _, item := range m.list.Items() {
		if fileItem, ok := item.(ui.FileItem); ok && !m.isGitIgnored(fileItem.Path) {
			m.toggleSelection(fileItem.Path, true)
		}
	}
}

// deselectAll deselects all items
func (m *Model) deselectAll() {
	for i := range m.items {
		m.items[i].Selected = false
	}
	m.refreshVisibleItems()
}

// selectByExtension selects all items with given extension
func (m *Model) selectByExtension(ext string) {
	// Ensure extension has a dot prefix
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	for i := range m.items {
		if !m.items[i].IsDir && strings.HasSuffix(strings.ToLower(m.items[i].Path), strings.ToLower(ext)) {
			m.toggleSelection(m.items[i].Path, true)
		}
	}
}

// toggleContentSearchMode toggles content search mode
func (m *Model) toggleContentSearchMode() {
	m.contentSearchMode = !m.contentSearchMode
	m.config.ContentSearchMode = m.contentSearchMode
	saveConfig(m.config, filepath.Join(os.Getenv("HOME"), ".config", "llmdog", "config.json"))

	if m.contentSearchMode {
		m.setStatusMessage("Content search enabled", 2)
	} else {
		m.setStatusMessage("Content search disabled", 2)
	}
}

// Init initializes the bubbletea model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
	)
}

// BuildOutput creates the markdown output from selected items
func BuildOutput(items []ui.FileItem, cwd string) string {
	var sb strings.Builder

	// File structure section
	sb.WriteString("# Directory Structure\n```\n")
	for _, item := range items {
		rel, err := filepath.Rel(cwd, item.Path)
		if err != nil {
			rel = item.Path
		}
		if item.IsDir {
			sb.WriteString(fmt.Sprintf("%s/\n", rel))
			sb.WriteString(buildTree(item.Path, 0))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", rel))
		}
	}
	sb.WriteString("```\n")

	// File contents section
	sb.WriteString("\n# File Contents\n")
	for _, item := range items {
		if !item.IsDir {
			rel, err := filepath.Rel(cwd, item.Path)
			if err != nil {
				rel = item.Path
			}

			content, err := os.ReadFile(item.Path)
			if err == nil {
				ext := filepath.Ext(item.Path)
				if ext == "" {
					ext = "txt"
				} else {
					ext = ext[1:]
				}

				sb.WriteString(fmt.Sprintf("\n## File: %s\n", rel))
				sb.WriteString("```" + ext + "\n")
				sb.WriteString(string(content))
				if !strings.HasSuffix(string(content), "\n") {
					sb.WriteString("\n")
				}
				sb.WriteString("```\n")
			}
		}
	}
	return sb.String()
}

func buildTree(root string, level int) string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Sprintf("Error reading directory: %v", err)
	}

	var sb strings.Builder
	indent := strings.Repeat(" ", level*2)

	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.IsDir() {
			sb.WriteString(fmt.Sprintf("%s|- %s/\n", indent, entry.Name()))
			sb.WriteString(buildTree(path, level+1))
		} else {
			sb.WriteString(fmt.Sprintf("%s|- %s\n", indent, entry.Name()))
		}
	}
	return sb.String()
}

// Update updates the application state
// Update updates the application state
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case errMsg:
		m.addError(msg.err)
		m.isLoading = false
		return m, nil

	case successMsg:
		m.setStatusMessage(msg.message, 2)
		return m, nil

	case childrenLoadedMsg:
		// First mark the parent directory as having loaded children
		for i := range m.items {
			if m.items[i].Path == msg.parentPath {
				m.items[i].ChildrenLoaded = true
				break
			}
		}

		// Check for duplicates before adding children
		existingPaths := make(map[string]bool)
		for _, item := range m.items {
			existingPaths[item.Path] = true
		}

		// Only add children that don't already exist
		var newChildren []ui.FileItem
		for _, child := range msg.children {
			if !existingPaths[child.Path] {
				newChildren = append(newChildren, child)
			}
		}

		// Add the new children to the items slice
		if len(newChildren) > 0 {
			m.items = append(m.items, newChildren...)
		}

		m.isLoading = false
		m.refreshVisibleItems()
		return m, nil

	case tea.KeyMsg:
		// Handle text input modal if active
		if m.showTextInputModal {
			switch msg.String() {
			case "esc":
				m.showTextInputModal = false
				return m, nil

			case "enter":
				// Process based on purpose
				inputValue := m.textInputModal.Value()
				if inputValue == "" {
					m.setStatusMessage("Bookmark name cannot be empty", 2)
					m.showTextInputModal = false
					return m, nil
				}

				switch m.textInputPurpose {
				case "new_bookmark":
					err := m.saveCurrentSelectionAsBookmark(inputValue, "")
					if err != nil {
						m.addError(err)
					} else {
						m.setStatusMessage(fmt.Sprintf("Saved bookmark: %s", inputValue), 2)
					}

				case "rename_bookmark":
					err := m.renameBookmark(m.tempBookmarkName, inputValue)
					if err != nil {
						m.addError(err)
					} else {
						m.setStatusMessage(fmt.Sprintf("Renamed bookmark to: %s", inputValue), 2)
					}

				case "bookmark_description":
					// Get the bookmark and update its description
					bookmark, found := m.bookmarkStore.GetBookmark(m.tempBookmarkName)
					if found {
						bookmark.Description = inputValue
						bookmark.Modified = time.Now()
						err := m.bookmarkStore.SaveBookmark(bookmark)
						if err != nil {
							m.addError(err)
						} else {
							m.setStatusMessage("Updated bookmark description", 2)
						}
					}
				}

				m.showTextInputModal = false

				// Refresh bookmarks menu if it's open
				if m.showBookmarksMenu {
					m.bookmarksMenu = ui.NewBookmarksMenu(
						m.bookmarkStore.Bookmarks,
						m.termWidth/2,
						m.termHeight/2,
					)
				}

				return m, nil

			default:
				// Pass other keys to text input
				modal, cmd := m.textInputModal.Update(msg)
				m.textInputModal = modal
				return m, cmd
			}
		}

		// Handle bookmarks menu if active
		if m.showBookmarksMenu {
			switch msg.String() {
			case "esc":
				m.showBookmarksMenu = false
				return m, nil

			case "enter":
				// Apply selected bookmark
				if name, ok := m.bookmarksMenu.SelectedBookmark(); ok {
					err := m.applyBookmark(name)
					if err != nil {
						m.addError(err)
					}
					m.showBookmarksMenu = false
				}
				return m, nil

			case "n":
				// Create new bookmark
				m.showBookmarksMenu = false
				m.showNewBookmarkDialog()
				return m, nil

			case "d":
				// Delete selected bookmark
				if name, ok := m.bookmarksMenu.SelectedBookmark(); ok {
					err := m.deleteBookmark(name)
					if err != nil {
						m.addError(err)
					}

					// Refresh bookmarks menu
					m.bookmarksMenu = ui.NewBookmarksMenu(
						m.bookmarkStore.Bookmarks,
						m.termWidth/2,
						m.termHeight/2,
					)
				}
				return m, nil

			case "r":
				// Rename selected bookmark
				m.showRenameBookmarkDialog()
				return m, nil

			case "i":
				// Add/edit description for the bookmark
				if name, ok := m.bookmarksMenu.SelectedBookmark(); ok {
					bookmark, found := m.bookmarkStore.GetBookmark(name)
					if found {
						m.tempBookmarkName = name
						m.textInputModal = ui.NewTextInputModal(
							"Enter Bookmark Description",
							bookmark.Description,
							m.termWidth/2,
						)
						m.showTextInputModal = true
						m.textInputPurpose = "bookmark_description"
					}
				}
				return m, nil

			default:
				// Pass other keys to bookmarks menu
				bmMenu, cmd := m.bookmarksMenu.Update(msg)
				m.bookmarksMenu = bmMenu
				return m, cmd
			}
		}

		// Handle filtering state separately
		if m.list.FilterState() == list.Filtering {
			switch msg.String() {
			case "up":
				if m.searchHistoryIndex > 0 {
					m.searchHistoryIndex--
					// Since we can't set the filter directly, we'll apply our custom search
					// on the current search history item
					if len(m.searchHistory) > 0 {
						m.performSearch(m.searchHistory[m.searchHistoryIndex])
					}
				}
				return m, nil

			case "down":
				if m.searchHistoryIndex < len(m.searchHistory)-1 {
					m.searchHistoryIndex++
					// Apply search with history item
					if len(m.searchHistory) > 0 {
						m.performSearch(m.searchHistory[m.searchHistoryIndex])
					}
				}
				return m, nil

			case "enter", "esc":
				query := m.list.FilterValue()
				if query != "" && (len(m.searchHistory) == 0 || m.searchHistory[len(m.searchHistory)-1] != query) {
					m.searchHistory = append(m.searchHistory, query)
					m.searchHistoryIndex = len(m.searchHistory)
				}

				// Perform search instead of default behavior
				if msg.String() == "enter" {
					m.performSearch(query)
					return m, nil
				}
			}
		} else {
			// Override default list behavior for filter input changes
			if m.list.FilterState() == list.Filtering {
				m.list, cmd = m.list.Update(msg)

				// After update, check if the filter changed and perform our custom search
				query := m.list.FilterValue()
				m.performSearch(query)

				return m, cmd
			}

			// Regular key handling
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit

			case " ": // Space key for expansion/collapse
				selectedItem, ok := m.list.SelectedItem().(ui.FileItem)
				if !ok {
					return m, nil
				}
				cmd := m.toggleExpansion(selectedItem.Path)
				return m, cmd

			case "tab": // Tab key for selection
				selectedItem, ok := m.list.SelectedItem().(ui.FileItem)
				if !ok {
					return m, nil
				}
				m.toggleSelection(selectedItem.Path)
				return m, nil

			case "ctrl+/":
				m.showPreview = !m.showPreview
				return m, nil

			case "ctrl+s":
				m.toggleContentSearchMode()
				return m, nil

			case "ctrl+a": // Select all visible
				m.selectAll()
				return m, nil

			case "ctrl+d": // Deselect all
				m.deselectAll()
				return m, nil

			case "ctrl+b": // Toggle bookmarks menu
				if !m.showBookmarksMenu {
					m.bookmarksMenu = ui.NewBookmarksMenu(
						m.bookmarkStore.Bookmarks,
						m.termWidth/2,
						m.termHeight/2,
					)
					m.showBookmarksMenu = true
				} else {
					m.showBookmarksMenu = false
				}
				return m, nil

			case "ctrl+shift+b": // Save bookmark shortcut
				m.showNewBookmarkDialog()
				return m, nil

			case "esc":
				if m.showErrors {
					m.showErrors = false
					m.errors = []string{}
					return m, nil
				}

			case "enter":
				var selected []ui.FileItem
				for _, item := range m.items {
					if item.Selected && !m.isGitIgnored(item.Path) {
						selected = append(selected, item)
					}
				}
				if len(selected) == 0 {
					if sel, ok := m.list.SelectedItem().(ui.FileItem); ok && !m.isGitIgnored(sel.Path) {
						selected = append(selected, sel)
					}
				}

				if len(selected) == 0 {
					m.setStatusMessage("No files selected!", 2)
					return m, nil
				}

				output := BuildOutput(selected, m.cwd)
				err := clipboard.WriteAll(output)
				if err != nil {
					m.addError(fmt.Errorf("Failed to copy to clipboard: %v", err))
					return m, nil
				}

				fmt.Printf("\nFetched %d items! 🐕 Woof!\n", len(selected))
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.list.SetWidth(msg.Width / 2)
		m.list.SetHeight(msg.Height - 5)
	}

	m.list, cmd = m.list.Update(msg)
	if sel, ok := m.list.SelectedItem().(ui.FileItem); ok {
		m.preview = ui.LoadPreview(sel.Path, sel.IsDir, m.config.MaxPreviewSize)
	}
	return m, cmd
}

// View renders the UI
// View renders the UI
func (m *Model) View() string {
	if m.isLoading {
		return fmt.Sprintf("%s %s", m.spinner.View(), m.loadingMessage)
	}

	// Base view creation
	var mainView string
	if !m.showPreview {
		mainView = ui.RenderHeader("llmdog") + "\n" +
			m.list.View() + "\n" +
			m.renderStatusBar()
	} else {
		// Calculate appropriate widths
		listWidth := m.termWidth * 2 / 3            // File list gets 2/3 of width
		previewWidth := m.termWidth - listWidth - 4 // Preview gets remaining space

		m.list.SetWidth(listWidth)
		previewStyle := ui.PreviewStyle.MaxWidth(previewWidth).MaxHeight(m.termHeight - 6)

		leftPanel := m.list.View()
		rightPanel := previewStyle.Render(ui.TruncatePreview(m.preview, m.termHeight-8))

		mainView = ui.RenderHeader("llmdog") + "\n" +
			lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	}

	// Show text input modal if active
	if m.showTextInputModal {
		mainView = lipgloss.Place(
			m.termWidth,
			m.termHeight-2, // Account for status bar
			lipgloss.Center,
			lipgloss.Center,
			m.textInputModal.View(),
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("240")),
		)
	}

	// Show bookmarks menu if active
	if m.showBookmarksMenu {
		mainView = lipgloss.Place(
			m.termWidth,
			m.termHeight-2, // Account for status bar
			lipgloss.Center,
			lipgloss.Center,
			m.bookmarksMenu.View(),
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("240")),
		)
	}

	// Show error messages
	if m.showErrors && len(m.errors) > 0 {
		errorText := strings.Join(m.errors, "\n")
		errorBox := lipgloss.NewStyle().
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Render(errorText)

		mainView += "\n" + errorBox
	}

	return mainView + "\n" + m.renderStatusBar()
}

func (m *Model) renderStatusBar() string {
	// Show status message if it's active
	if m.statusMessage != "" && time.Now().Before(m.statusMessageExpiry) {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("205")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			Width(m.termWidth).
			Render(m.statusMessage)
	}

	// Stats part
	statsText := fmt.Sprintf("Selected: %d files (%.1f KB) • Est. Tokens: ~%d",
		m.selectedCount, float64(m.selectedSize)/1024, m.estimatedTokens)

	// Add bookmark count to stats text if bookmarks exist
	if len(m.bookmarkStore.Bookmarks) > 0 {
		statsText = fmt.Sprintf("%s • Bookmarks: %d", statsText, len(m.bookmarkStore.Bookmarks))
	}

	// Help part
	var helpText string
	if m.showBookmarksMenu {
		helpText = "Enter:Apply • n:New • d:Delete • r:Rename • Esc:Close"
	} else {
		helpText = "Tab:Select • Ctrl+B:Bookmarks • Ctrl+S:Search Mode"
	}

	// Show content search mode
	modeText := "Mode: "
	if m.contentSearchMode {
		modeText += "Content Search"
	} else {
		modeText += "Filename Search"
	}

	// Combine everything
	statusBar := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(m.termWidth/3).Render(statsText),
		lipgloss.NewStyle().Width(m.termWidth/3).Align(lipgloss.Center).Render(modeText),
		lipgloss.NewStyle().Width(m.termWidth/3).Align(lipgloss.Right).Render(helpText),
	)

	return lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("255")).
		Width(m.termWidth).
		Render(statusBar)
}

// executeCustomSearch performs a custom search operation
func (m *Model) executeCustomSearch(query string) {
	// If no query, show all visible items
	if query == "" {
		m.isInSearchResults = false
		m.refreshVisibleItems()
		return
	}

	m.isInSearchResults = true

	// Process query to lowercase for case-insensitive matching
	queryLower := strings.ToLower(query)

	// Reset content match flags first
	for i := range m.items {
		m.items[i].MatchesContent = false
	}

	// First pass: Find exact matches to the search term
	var exactMatches []ui.FileItem
	for i := range m.items {
		if !m.items[i].IsDir {
			// Check if the filename exactly matches the search term
			// This is for files like "main.go" when searching for "main.go"
			if strings.ToLower(m.items[i].Name) == queryLower {
				// Create a copy to avoid modifying the original
				fileItem := m.items[i]
				exactMatches = append(exactMatches, fileItem)
			}
		}
	}

	// If we have exact matches, just use those
	if len(exactMatches) > 0 {
		// Now we need to make sure all parent directories are shown
		var results []list.Item
		resultPaths := make(map[string]bool)

		// Add all exact matches and their parent directories
		for _, item := range exactMatches {
			// Skip if already processed
			if resultPaths[item.Path] {
				continue
			}

			// Add the item to results
			resultPaths[item.Path] = true
			results = append(results, item)

			// Make sure all parent directories are expanded and visible
			addParentDirs(item.Path, m.cwd, &results, &resultPaths, m.items)
		}

		// Sort results by path to maintain hierarchy
		sort.Slice(results, func(i, j int) bool {
			itemI, _ := results[i].(ui.FileItem)
			itemJ, _ := results[j].(ui.FileItem)
			return itemI.Path < itemJ.Path
		})

		// Show results
		m.list.SetItems(results)
		m.setStatusMessage(fmt.Sprintf("Found %d exact matches", len(exactMatches)), 2)
		return
	}

	// If no exact matches, do a fuzzy search
	var results []list.Item
	resultPaths := make(map[string]bool)
	var matchCount int

	// Look for partial matches in filenames
	for i := range m.items {
		// Skip directories in partial matching to avoid matching parent folder names
		if !m.items[i].IsDir && strings.Contains(strings.ToLower(m.items[i].Name), queryLower) {
			if !resultPaths[m.items[i].Path] {
				resultPaths[m.items[i].Path] = true
				results = append(results, m.items[i])
				matchCount++

				// Make sure all parent directories are expanded and visible
				addParentDirs(m.items[i].Path, m.cwd, &results, &resultPaths, m.items)
			}
		}
	}

	// If no filename matches and content search is enabled, search in content
	if matchCount == 0 && m.contentSearchMode {
		// Search file contents for the query
		for i := range m.items {
			if !m.items[i].IsDir && !resultPaths[m.items[i].Path] {
				// Only check smaller files to avoid performance issues
				info, err := os.Stat(m.items[i].Path)
				if err == nil && info.Size() < 1024*1024 { // Skip files larger than 1MB
					content, err := os.ReadFile(m.items[i].Path)
					if err == nil && strings.Contains(strings.ToLower(string(content)), queryLower) {
						// Mark as content match for UI highlighting
						fileItem := m.items[i]
						fileItem.MatchesContent = true

						// Add to results
						resultPaths[fileItem.Path] = true
						results = append(results, fileItem)
						matchCount++

						// Make sure all parent directories are expanded and visible
						addParentDirs(fileItem.Path, m.cwd, &results, &resultPaths, m.items)
					}
				}
			}
		}
	}

	// Sort results by path to maintain hierarchy
	sort.Slice(results, func(i, j int) bool {
		itemI, _ := results[i].(ui.FileItem)
		itemJ, _ := results[j].(ui.FileItem)
		return itemI.Path < itemJ.Path
	})

	// Show results or message
	if len(results) > 0 {
		m.list.SetItems(results)
		m.setStatusMessage(fmt.Sprintf("Found %d matches", matchCount), 2)
	} else if m.contentSearchMode {
		m.setStatusMessage("No matches found. Try different search terms.", 2)
		m.isInSearchResults = false
		m.refreshVisibleItems()
	} else {
		m.setStatusMessage("No matches found. Try content search (Ctrl+S).", 2)
		m.isInSearchResults = false
		m.refreshVisibleItems()
	}
}

// addParentDirs adds all parent directories of a path to the results
func addParentDirs(path, rootPath string, results *[]list.Item, resultPaths *map[string]bool, allItems []ui.FileItem) {
	// Get the parent directory path
	parentPath := filepath.Dir(path)

	// If we've reached the root or above, stop
	if parentPath == rootPath || parentPath == "." {
		return
	}

	// Recursively process parent directories first
	addParentDirs(parentPath, rootPath, results, resultPaths, allItems)

	// Then add this parent if not already added
	if !(*resultPaths)[parentPath] {
		// Find the parent directory in allItems
		for _, item := range allItems {
			if item.Path == parentPath && item.IsDir {
				// Add it to results
				(*resultPaths)[parentPath] = true
				*results = append(*results, item)
				break
			}
		}
	}
}

func (m *Model) saveCurrentSelectionAsBookmark(name, description string) error {
	var selectedPaths []string

	for _, item := range m.items {
		if item.Selected && !m.isGitIgnored(item.Path) {
			// Store paths relative to the current working directory
			relPath, err := filepath.Rel(m.cwd, item.Path)
			if err == nil {
				selectedPaths = append(selectedPaths, relPath)
			} else {
				selectedPaths = append(selectedPaths, item.Path)
			}
		}
	}

	if len(selectedPaths) == 0 {
		return fmt.Errorf("no files selected")
	}

	bookmark := bookmarks.Bookmark{
		Name:        name,
		Description: description,
		FilePaths:   selectedPaths,
		RootPath:    m.cwd,
		Created:     time.Now(),
		Modified:    time.Now(),
	}

	return m.bookmarkStore.SaveBookmark(bookmark)
}

// applyBookmark applies a saved bookmark selection
func (m *Model) applyBookmark(name string) error {
	bookmark, found := m.bookmarkStore.GetBookmark(name)
	if !found {
		return fmt.Errorf("bookmark not found: %s", name)
	}

	// Reset current selection
	m.deselectAll()

	// Apply bookmark selection
	for _, relPath := range bookmark.FilePaths {
		// Convert relative path to absolute based on current directory
		absPath := filepath.Join(m.cwd, relPath)

		// Find item and select it
		for i := range m.items {
			if m.items[i].Path == absPath {
				m.toggleSelection(absPath, true)

				// Ensure parent directories are expanded to make the item visible
				m.ensureParentPathsExpanded(absPath)
				break
			}
		}
	}

	m.refreshVisibleItems()
	m.setStatusMessage(fmt.Sprintf("Applied bookmark: %s", name), 2)
	return nil
}

// deleteBookmark deletes a bookmark
func (m *Model) deleteBookmark(name string) error {
	err := m.bookmarkStore.DeleteBookmark(name)
	if err != nil {
		return err
	}

	m.setStatusMessage(fmt.Sprintf("Deleted bookmark: %s", name), 2)
	return nil
}

// renameBookmark renames a bookmark
func (m *Model) renameBookmark(oldName, newName string) error {
	// Get the bookmark
	bookmark, found := m.bookmarkStore.GetBookmark(oldName)
	if !found {
		return fmt.Errorf("bookmark not found: %s", oldName)
	}

	// Delete the old bookmark
	err := m.bookmarkStore.DeleteBookmark(oldName)
	if err != nil {
		return err
	}

	// Save with new name
	bookmark.Name = newName
	bookmark.Modified = time.Now()
	return m.bookmarkStore.SaveBookmark(bookmark)
}

// showNewBookmarkDialog shows the dialog for creating a new bookmark
func (m *Model) showNewBookmarkDialog() {
	m.textInputModal = ui.NewTextInputModal(
		"Enter Bookmark Name",
		"My Bookmark",
		m.termWidth/2,
	)
	m.showTextInputModal = true
	m.textInputPurpose = "new_bookmark"
}

// showRenameBookmarkDialog shows the dialog for renaming a bookmark
func (m *Model) showRenameBookmarkDialog() {
	if name, ok := m.bookmarksMenu.SelectedBookmark(); ok {
		m.tempBookmarkName = name
		m.textInputModal = ui.NewTextInputModal(
			"Enter New Bookmark Name",
			name,
			m.termWidth/2,
		)
		m.showTextInputModal = true
		m.textInputPurpose = "rename_bookmark"
	}
}
