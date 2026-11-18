// Package todo holds the to-do items and the rules for reading them.
package todo

import (
	_ "embed"
	"encoding/json"
)

// Data is the built-in list of to-do items, as JSON.
//
//go:embed items.json
var Data []byte

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
