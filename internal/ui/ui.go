package ui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(1, 0)

	PreviewStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)

	// Item styles
	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	GitIgnoredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Faint(true)

	FolderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("110"))

	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	ContentMatchStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")).
				Bold(true)

	CursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255"))

	SelectedCursorStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("87")).
				Bold(true)

	EmphasisStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
)

// FileItem represents a file or directory in the file system
type FileItem struct {
	Path           string
	Name           string
	IsDir          bool
	Selected       bool
	Depth          int
	Expanded       bool
	GitIgnored     bool
	ChildrenLoaded bool
	MatchesContent bool
}

func (f FileItem) Title() string {
	var builder strings.Builder

	// Add directory expansion indicators
	if f.IsDir {
		if f.Expanded {
			builder.WriteString("▼ ")
		} else {
			builder.WriteString("▶ ")
		}
	} else {
		builder.WriteString("  ")
	}

	// Add selection checkbox
	if f.Selected {
		builder.WriteString("[✓] ")
	} else {
		builder.WriteString("[ ] ")
	}

	// Add icon based on file type
	icon := getFileIcon(f.Name, f.IsDir)
	builder.WriteString(icon)
	builder.WriteString(" ")

	// Add filename
	builder.WriteString(f.Name)

	return builder.String()
}

func (f FileItem) Description() string {
	if f.MatchesContent {
		return "content match"
	}
	if f.GitIgnored {
		return "gitignored"
	}
	info := getFileInfo(f)
	if info != "" {
		return info
	}
	if f.IsDir {
		return "directory"
	}
	return "file"
}

func (f FileItem) FilterValue() string {
	return f.Name
}

// ItemDelegate handles the rendering of list items
type ItemDelegate struct{}

func (d ItemDelegate) Height() int                               { return 1 }
func (d ItemDelegate) Spacing() int                              { return 0 }
func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(FileItem)
	if !ok {
		return
	}

	// Base style with indentation
	style := lipgloss.NewStyle().PaddingLeft(i.Depth * 2)

	// Build the display string
	var builder strings.Builder

	// Add cursor for selected item
	if index == m.Index() {
		builder.WriteString("→ ")
	} else {
		builder.WriteString("  ")
	}

	// Add expansion indicator for directories
	if i.IsDir {
		if i.Expanded {
			builder.WriteString("▼ ")
		} else {
			builder.WriteString("▶ ")
		}
	} else {
		builder.WriteString("  ")
	}

	// Add selection indicator with more visible checkboxes
	if i.Selected {
		builder.WriteString("✅ ")
	} else {
		builder.WriteString("☐  ")
	}

	// Add appropriate icon and name
	icon := getFileIcon(i.Name, i.IsDir)
	builder.WriteString(icon)
	builder.WriteString(" ")
	builder.WriteString(i.Name)

	// Add content match indicator
	if i.MatchesContent {
		builder.WriteString(" 🔍")
	}

	// Add size/count info
	info := getFileInfo(i)
	if info != "" {
		builder.WriteString(" ")
		builder.WriteString(info)
	}

	// Apply appropriate style based on item state
	if i.GitIgnored {
		style = style.Inherit(GitIgnoredStyle)
	} else if i.Selected && index == m.Index() {
		style = style.Inherit(SelectedCursorStyle)
	} else if i.Selected {
		style = style.Inherit(SelectedStyle)
	} else if i.MatchesContent {
		style = style.Inherit(ContentMatchStyle)
	} else if i.IsDir {
		style = style.Inherit(FolderStyle)
	} else if index == m.Index() {
		style = style.Inherit(CursorStyle)
	} else {
		style = style.Inherit(NormalStyle)
	}

	fmt.Fprint(w, style.Render(builder.String()))
}

func getFileIcon(name string, isDir bool) string {
	if isDir {
		return "📁" // Directory icon
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return "🔹"
	case ".py":
		return "🐍"
	case ".js", ".ts":
		return "🟡"
	case ".md":
		return "📝"
	case ".json":
		return "📦"
	case ".yml", ".yaml":
		return "⚙️"
	case ".gitignore":
		return "🐕"
	case ".txt":
		return "📄"
	case ".sh", ".bash":
		return "⚡"
	case ".css":
		return "🎨"
	case ".html":
		return "🌐"
	case ".sql":
		return "💾"
	case ".log":
		return "📊"
	case ".env":
		return "🔒"
	case ".toml":
		return "⚙️"
	case ".xml":
		return "📰"
	case ".csv":
		return "📊"
	case ".pdf":
		return "📕"
	case ".zip", ".tar", ".gz":
		return "📦"
	case ".mp3", ".wav", ".ogg":
		return "🎵"
	case ".mp4", ".mov", ".avi":
		return "🎬"
	case ".jpg", ".jpeg", ".png", ".gif":
		return "🖼️"
	case ".svg":
		return "📐"
	default:
		return "📄"
	}
}

func getFileInfo(item FileItem) string {
	info, err := os.Stat(item.Path)
	if err != nil {
		return ""
	}

	if item.IsDir {
		entries, err := os.ReadDir(item.Path)
		if err != nil {
			return ""
		}
		count := len(entries)
		if count == 1 {
			return "(1 item)"
		}
		return fmt.Sprintf("(%d items)", count)
	}

	size := info.Size()
	switch {
	case size == 0:
		return "(empty)"
	case size < 1024:
		return fmt.Sprintf("(%dB)", size)
	case size < 1024*1024:
		return fmt.Sprintf("(%.1fKB)", float64(size)/1024)
	case size < 1024*1024*1024:
		return fmt.Sprintf("(%.1fMB)", float64(size)/(1024*1024))
	default:
		return fmt.Sprintf("(%.1fGB)", float64(size)/(1024*1024*1024))
	}
}

// LoadFiles walks through the directory tree and returns a slice of FileItems
func LoadFiles(root string, gitRegex *regexp.Regexp, showHidden bool) []FileItem {
	var items []FileItem

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == root {
			return nil
		}

		// Skip hidden files if not enabled
		if !showHidden && isHiddenFile(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Calculate relative path and depth
		rel, _ := filepath.Rel(root, path)
		depth := len(strings.Split(rel, string(os.PathSeparator))) - 1

		// Check if item is gitignored
		isGitIgnored := gitRegex != nil && gitRegex.MatchString(path)

		item := FileItem{
			Path:           path,
			Name:           info.Name(),
			IsDir:          info.IsDir(),
			Selected:       false,
			Depth:          depth,
			Expanded:       false,
			GitIgnored:     isGitIgnored,
			ChildrenLoaded: false,
		}

		items = append(items, item)

		// Skip gitignored directories
		if isGitIgnored && info.IsDir() {
			return filepath.SkipDir
		}

		return nil
	})

	return items
}

// isHiddenFile checks if a file is hidden
func isHiddenFile(name string) bool {
	return strings.HasPrefix(name, ".")
}

// LoadDirectoryChildren loads only the direct children of a directory
func LoadDirectoryChildren(dirPath string, gitRegex *regexp.Regexp, showHidden bool) ([]FileItem, error) {
	var items []FileItem

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", dirPath, err)
	}

	// Calculate depth
	rel, err := filepath.Rel(dirPath, dirPath)
	if err != nil {
		rel = dirPath
	}
	baseDepth := len(strings.Split(rel, string(os.PathSeparator)))
	if baseDepth == 0 {
		baseDepth = 1
	}

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(dirPath, name)

		// Skip hidden files if not enabled
		if !showHidden && isHiddenFile(name) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			// Skip entries with errors instead of failing
			continue
		}

		// Check if item is gitignored
		isGitIgnored := gitRegex != nil && gitRegex.MatchString(path)

		item := FileItem{
			Path:           path,
			Name:           name,
			IsDir:          info.IsDir(),
			Selected:       false,
			Depth:          baseDepth,
			Expanded:       false,
			GitIgnored:     isGitIgnored,
			ChildrenLoaded: false,
		}

		items = append(items, item)
	}

	return items, nil
}

// LoadPreview generates a preview of the file or directory content
func LoadPreview(path string, isDir bool, maxSize int) string {
	if isDir {
		return loadDirectoryPreview(path)
	}
	return loadFilePreview(path, maxSize)
}

func loadDirectoryPreview(path string) string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Sprintf("Error reading directory: %v", err)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Directory: %s\n\n", path))

	// Count files and subdirectories
	var files, dirs int
	var fileTypes = make(map[string]int)

	for _, entry := range entries {
		if entry.IsDir() {
			dirs++
		} else {
			files++
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext != "" {
				fileTypes[ext]++
			} else {
				fileTypes["(no extension)"]++
			}
		}
	}
	builder.WriteString(fmt.Sprintf("Contains: %d files, %d directories\n\n", files, dirs))

	// Show file type breakdown
	if len(fileTypes) > 0 {
		builder.WriteString("File types:\n")
		for ext, count := range fileTypes {
			builder.WriteString(fmt.Sprintf("- %s: %d files\n", ext, count))
		}
		builder.WriteString("\n")
	}

	// List contents
	builder.WriteString("Contents:\n")
	for _, entry := range entries {
		if entry.IsDir() {
			builder.WriteString(fmt.Sprintf("📁 %s/\n", entry.Name()))
		} else {
			icon := getFileIcon(entry.Name(), false)
			builder.WriteString(fmt.Sprintf("%s %s\n", icon, entry.Name()))
		}
	}

	return builder.String()
}

var previewCache = struct {
	sync.RWMutex
	cache map[string]string
}{cache: make(map[string]string)}

func loadFilePreview(path string, maxSize int) string {
	// Check cache first
	previewCache.RLock()
	if preview, ok := previewCache.cache[path]; ok {
		previewCache.RUnlock()
		return preview
	}
	previewCache.RUnlock()

	file, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return fmt.Sprintf("Error getting file info: %v", err)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("File: %s\n", path))
	builder.WriteString(fmt.Sprintf("Size: %s\n", formatSize(info.Size())))
	builder.WriteString(fmt.Sprintf("Modified: %s\n\n", info.ModTime().Format("2006-01-02 15:04:05")))

	// Determine how to preview based on file type
	ext := strings.ToLower(filepath.Ext(path))
	isText := isTextFile(ext)

	if !isText {
		builder.WriteString(fmt.Sprintf("Binary file detected (%s format)\n", ext))
		return builder.String()
	}

	// Read file content
	if maxSize <= 0 {
		maxSize = 10000 // Default
	}
	data := make([]byte, maxSize)
	n, err := file.Read(data)
	if err != nil && err != io.EOF {
		return fmt.Sprintf("Error reading file content: %v", err)
	}

	// Process content
	content := string(data[:n])
	lines := strings.Split(content, "\n")

	// Truncate if too many lines
	maxLines := 50
	if len(lines) > maxLines {
		lines = append(lines[:maxLines], "... (content truncated)")
	}

	// Add syntax highlighting clues
	builder.WriteString("Content Preview:")

	// Simple syntax highlighting for common file types
	switch ext {
	case ".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".cs":
		builder.WriteString(" (code)")
	case ".md", ".txt", ".rst":
		builder.WriteString(" (text)")
	case ".json", ".yaml", ".yml", ".toml":
		builder.WriteString(" (config)")
	case ".html", ".xml", ".svg":
		builder.WriteString(" (markup)")
	case ".css", ".scss":
		builder.WriteString(" (style)")
	}

	builder.WriteString("\n")
	builder.WriteString(strings.Join(lines, "\n"))

	result := builder.String()

	// Cache the result
	previewCache.Lock()
	previewCache.cache[path] = result
	previewCache.Unlock()

	return result
}

// isTextFile checks if a file is likely a text file based on extension
func isTextFile(ext string) bool {
	textExtensions := []string{
		".txt", ".md", ".go", ".py", ".js", ".ts", ".html", ".css", ".json",
		".yaml", ".yml", ".xml", ".csv", ".sh", ".bash", ".toml", ".c", ".cpp",
		".h", ".hpp", ".java", ".properties", ".log", ".svg", ".sql",
		".gitignore", ".env", ".rs", ".rb", ".php",
	}

	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}

	return false
}

func formatSize(size int64) string {
	switch {
	case size < 1024:
		return fmt.Sprintf("%d B", size)
	case size < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	case size < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	default:
		return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
	}
}

// RenderHeader renders the application header
func RenderHeader(title string) string {
	return HeaderStyle.Render(fmt.Sprintf("🐕 %s", title))
}

// RenderLoading renders a loading indicator
func RenderLoading(message string) string {
	return EmphasisStyle.Render(fmt.Sprintf("Loading: %s", message))
}

func TruncatePreview(preview string, maxLines int) string {
	if maxLines <= 0 {
		return preview
	}

	lines := strings.Split(preview, "\n")
	if len(lines) <= maxLines {
		return preview
	}

	// Keep the first few lines with file info and truncate the content
	headerLines := 5                           // Keep file info, size, etc.
	contentLines := maxLines - headerLines - 1 // Reserve one line for truncation message

	result := append(lines[:headerLines], "...")
	result = append(result, lines[len(lines)-contentLines:]...)
	result = append(result, fmt.Sprintf("(Showing %d of %d lines)", maxLines, len(lines)))

	return strings.Join(result, "\n")
}
