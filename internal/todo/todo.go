// Package todo holds the to-do items and the rules for reading and storing them.
package todo

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Data is the built-in list of sample to-do items, as JSON. It is the list a
// user starts with until the first item is saved.
//
//go:embed items.json
var Data []byte

// FileEnv names the environment variable that overrides where items are kept.
const FileEnv = "GO_TODO_FILE"

// Item is one entry on the to-do list.
type Item struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Parse decodes a JSON array of items.
func Parse(data []byte) ([]Item, error) {
	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// Incomplete returns the items that are not done, in their original order.
func Incomplete(items []Item) []Item {
	var open []Item
	for _, item := range items {
		if !item.Done {
			open = append(open, item)
		}
	}
	return open
}

// Path returns the file the items are kept in: the file named by GO_TODO_FILE
// when it is set and not empty, otherwise ~/.go-todo/items.json.
func Path() (string, error) {
	if path := os.Getenv(FileEnv); path != "" {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".go-todo", "items.json"), nil
}

// Load reads the items kept at path. A file that does not exist yet holds the
// built-in sample items.
func Load(path string) ([]Item, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Parse(Data)
	}
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Save writes items to path, creating its folder when needed. It writes
// path + ".tmp" first and renames it into place, so a failed write never
// leaves a half-written list behind.
func Save(path string, items []Item) error {
	if items == nil {
		items = []Item{}
	}
	// An Item holds only a string and a bool, which encoding/json always
	// encodes, so MarshalIndent cannot fail here.
	data, _ := json.MarshalIndent(items, "", "  ")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
