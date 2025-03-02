package bookmarks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Bookmark represents a saved selection pattern
type Bookmark struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	FilePaths   []string  `json:"filePaths"`
	RootPath    string    `json:"rootPath"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

// BookmarkStore manages all bookmarks
type BookmarkStore struct {
	Bookmarks []Bookmark `json:"bookmarks"`
}

// LoadBookmarks loads bookmarks from disk
func LoadBookmarks() (BookmarkStore, error) {
	store := BookmarkStore{
		Bookmarks: []Bookmark{},
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "llmdog")
	bookmarksPath := filepath.Join(configDir, "bookmarks.json")

	data, err := os.ReadFile(bookmarksPath)
	if err != nil {
		// If file doesn't exist, create default store
		if os.IsNotExist(err) {
			os.MkdirAll(configDir, 0755)
			saveBookmarks(store, bookmarksPath)
			return store, nil
		}
		return store, err
	}

	err = json.Unmarshal(data, &store)
	return store, err
}

// SaveBookmark adds or updates a bookmark and persists to disk
func (store *BookmarkStore) SaveBookmark(bookmark Bookmark) error {
	// Look for existing bookmark with same name
	found := false
	for i, b := range store.Bookmarks {
		if b.Name == bookmark.Name {
			// Update existing
			store.Bookmarks[i] = bookmark
			found = true
			break
		}
	}

	if !found {
		// Add new
		store.Bookmarks = append(store.Bookmarks, bookmark)
	}

	// Save to disk
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "llmdog")
	bookmarksPath := filepath.Join(configDir, "bookmarks.json")
	return saveBookmarks(*store, bookmarksPath)
}

// saveBookmarks saves bookmarks to disk
func saveBookmarks(store BookmarkStore, path string) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// DeleteBookmark removes a bookmark
func (store *BookmarkStore) DeleteBookmark(name string) error {
	for i, b := range store.Bookmarks {
		if b.Name == name {
			// Remove by index
			store.Bookmarks = append(store.Bookmarks[:i], store.Bookmarks[i+1:]...)

			// Save to disk
			configDir := filepath.Join(os.Getenv("HOME"), ".config", "llmdog")
			bookmarksPath := filepath.Join(configDir, "bookmarks.json")
			return saveBookmarks(*store, bookmarksPath)
		}
	}

	return nil // Bookmark not found - no error
}

// GetBookmark retrieves a bookmark by name
func (store *BookmarkStore) GetBookmark(name string) (Bookmark, bool) {
	for _, b := range store.Bookmarks {
		if b.Name == name {
			return b, true
		}
	}

	return Bookmark{}, false
}
