package model

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/doganarif/llmdog/internal/git"
	"github.com/doganarif/llmdog/internal/ui"
)

type Model struct {
	list            list.Model
	preview         string
	items           []ui.FileItem
	cwd             string
	gitignoreRegexp *regexp.Regexp
	termWidth       int
	termHeight      int
	showPreview     bool
}

func New() *Model {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	gitRegex, _ := git.ParseGitignore(filepath.Join(cwd, ".gitignore"))
	items := ui.LoadFiles(cwd, gitRegex)

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

	return &Model{
		list:            l,
		items:           items,
		cwd:             cwd,
		gitignoreRegexp: gitRegex,
		showPreview:     true,
	}
}

func (m *Model) isGitIgnored(path string) bool {
	return m.gitignoreRegexp != nil && m.gitignoreRegexp.MatchString(path)
}

func (m *Model) getDirectChildren(parentPath string) []ui.FileItem {
	var children []ui.FileItem
	for i := range m.items {
		if filepath.Dir(m.items[i].Path) == parentPath {
			children = append(children, m.items[i])
		}
	}
	return children
}

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

func (m *Model) refreshVisibleItems() {
	visible := make([]list.Item, 0)
	selectedIndex := m.list.Index()
	var selectedPath string
	if sel, ok := m.list.SelectedItem().(ui.FileItem); ok {
		selectedPath = sel.Path
	}

	for i := range m.items {
		if m.isVisible(m.items[i]) {
			visible = append(visible, m.items[i])
		}
	}

	m.list.SetItems(visible)

	// Restore selection
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
}

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

func (m *Model) toggleExpansion(path string) {
	var currentItem *ui.FileItem
	for i := range m.items {
		if m.items[i].Path == path {
			if m.items[i].IsDir {
				m.items[i].Expanded = !m.items[i].Expanded
				currentItem = &m.items[i]
			}
			break
		}
	}

	if currentItem != nil {
		m.refreshVisibleItems()
	}
}

func (m *Model) toggleSelection(path string) {
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

	if currentItem.IsDir {
		// If directory is already selected, unselect it and all descendants
		if currentItem.Selected {
			currentItem.Selected = false
			m.setSelectionStateForDescendants(currentItem.Path, false)
		} else {
			// If directory is not selected, select it and all non-gitignored descendants
			currentItem.Selected = true
			m.setSelectionStateForDescendants(currentItem.Path, true)
		}
	} else {
		// Toggle file selection
		currentItem.Selected = !currentItem.Selected
	}

	// Update parent directory selection states
	m.updateParentSelectionState(path)
	m.refreshVisibleItems()
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case " ": // Space key for expansion/collapse
			selectedItem, ok := m.list.SelectedItem().(ui.FileItem)
			if !ok {
				return m, nil
			}
			m.toggleExpansion(selectedItem.Path)
			return m, nil

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
			output := BuildOutput(selected, m.cwd)
			clipboard.WriteAll(output)
			fmt.Printf("\nFetched %d items! 🐕 Woof!\n", len(selected))
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.list.SetWidth(msg.Width / 2)
		m.list.SetHeight(msg.Height - 5)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if sel, ok := m.list.SelectedItem().(ui.FileItem); ok {
		m.preview = ui.LoadPreview(sel.Path, sel.IsDir)
	}
	return m, cmd
}

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

func (m *Model) View() string {
	if !m.showPreview {
		return ui.RenderHeader("llmdog") + "\n" +
			m.list.View() + "\n" +
			"Press ctrl+/ to show preview"
	}

	previewWidth := m.termWidth/2 - 4
	previewStyle := ui.PreviewStyle.MaxWidth(previewWidth)

	leftPanel := m.list.View()
	rightPanel := previewStyle.Render(m.preview)

	return ui.RenderHeader("llmdog") + "\n" +
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel) + "\n" +
		"Press ctrl+/ to toggle preview"
}
