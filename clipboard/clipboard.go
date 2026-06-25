//go:build !js || !wasm

// Package clipboard ...
package clipboard

import (
	"log"

	"golang.design/x/clipboard"
)

func init() {
	err := clipboard.Init()
	if err != nil {
		log.Printf("[clipboard] init failed: %v", err)
	}
}

// Write copies text to the system clipboard.
func Write(text string) {
	clipboard.Write(clipboard.FmtText, []byte(text))
}

// Read reads text from the system clipboard and passes it to cb.
func Read(cb func(string)) {
	cb(string(clipboard.Read(clipboard.FmtText)))
}
