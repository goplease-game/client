package mock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

var savedSnapshots = map[string]ds.GameSnapshot{}

// SaveState serializes the current GameState, stores it in memory,
// and writes it to ./mock/data/<name>.json on disk.
func SaveState(name string, snap ds.GameSnapshot) error {
	if !strings.HasSuffix(name, ".json") {
		name += ".json"
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	filePath := filepath.Join("mock", "data", name)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Keep in memory for immediate access.
	savedSnapshots[name] = snap

	return nil
}

func LoadState(name string) (ds.GameSnapshot, error) {
	if !strings.HasSuffix(name, ".json") {
		name += ".json"
	}

	// 1. In-memory (runtime saves).
	if snap, ok := savedSnapshots[name]; ok {
		return snap, nil
	}

	// 2. Disk (includes both original embedded files and saved files).
	filePath := filepath.Join("mock", "data", name)
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return ds.GameSnapshot{}, fmt.Errorf("read file %q: %w", name, err)
	}

	var snap ds.GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		return ds.GameSnapshot{}, fmt.Errorf("parse %q: %w", name, err)
	}

	return snap, nil
}

// ListStates returns all available state names: embedded first, then runtime saves.
func ListStates() []string {
	var names []string
	seen := map[string]bool{}

	// Read from disk (includes both original and saved files).
	entries, _ := os.ReadDir(filepath.Join("mock", "data"))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
			seen[e.Name()] = true
		}
	}

	// In-memory saves not yet written (shouldn't happen now, but as fallback).
	for name := range savedSnapshots {
		if !seen[name] {
			names = append(names, name)
		}
	}

	return names
}
