package app

import (
	"os"
	"skilltrace/internal/platform"
)

type SourceInventoryItem struct {
	Harness, Status string
	PathPresent     bool
}

func SourceInventory() []SourceInventoryItem {
	items := []SourceInventoryItem{}
	for _, path := range platform.ClaudeRoots() {
		_, err := os.Stat(path)
		status := "available"
		if err != nil {
			status = "unavailable"
		}
		items = append(items, SourceInventoryItem{Harness: "claude", Status: status, PathPresent: err == nil})
	}
	return items
}
