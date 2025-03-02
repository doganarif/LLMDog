package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/doganarif/llmdog/internal/bookmarks"
)

// BookmarkItem represents a bookmark in the UI list
type BookmarkItem struct {
	Name     string
	DescText string
}

// Implement list.Item interface
func (b BookmarkItem) Title() string       { return b.Name }
func (b BookmarkItem) Description() string { return b.DescText }
func (b BookmarkItem) FilterValue() string { return b.Name }

// BookmarksMenu is the UI component for bookmark management
type BookmarksMenu struct {
	list   list.Model
	width  int
	height int
}

// NewBookmarksMenu creates a new bookmarks menu
func NewBookmarksMenu(bookmarks []bookmarks.Bookmark, width, height int) BookmarksMenu {
	var items []list.Item
	for _, b := range bookmarks {
		items = append(items, BookmarkItem{
			Name:     b.Name,
			DescText: b.Description,
		})
	}

	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = " Bookmarks  |  Enter:Apply  •  n:New  •  d:Delete  •  r:Rename  •  Esc:Close "

	return BookmarksMenu{
		list:   l,
		width:  width,
		height: height,
	}
}

// Update handles input for the bookmarks menu
func (b *BookmarksMenu) Update(msg tea.Msg) (BookmarksMenu, tea.Cmd) {
	var cmd tea.Cmd
	b.list, cmd = b.list.Update(msg)
	return *b, cmd
}

// View renders the bookmarks menu
func (b *BookmarksMenu) View() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(b.width).
		Render(b.list.View())
}

// SelectedBookmark returns the currently selected bookmark
func (b *BookmarksMenu) SelectedBookmark() (string, bool) {
	if len(b.list.Items()) == 0 {
		return "", false
	}

	selected, ok := b.list.SelectedItem().(BookmarkItem)
	if !ok {
		return "", false
	}

	return selected.Name, true
}

// TextInputModal is a modal for text input
type TextInputModal struct {
	textInput textinput.Model
	title     string
	width     int
}

// NewTextInputModal creates a new text input modal
func NewTextInputModal(title string, placeholder string, width int) TextInputModal {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()

	return TextInputModal{
		textInput: ti,
		title:     title,
		width:     width,
	}
}

// Update handles input for the text input
func (t *TextInputModal) Update(msg tea.Msg) (TextInputModal, tea.Cmd) {
	var cmd tea.Cmd
	t.textInput, cmd = t.textInput.Update(msg)
	return *t, cmd
}

// View renders the text input modal
func (t *TextInputModal) View() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(t.width).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				EmphasisStyle.Render(t.title),
				"",
				t.textInput.View(),
				"",
				"Enter: Confirm • Esc: Cancel",
			),
		)
}

// Value returns the current input value
func (t *TextInputModal) Value() string {
	return t.textInput.Value()
}
