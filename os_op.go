//go:build !js || !wasm

package game

import (
	"github.com/pkg/browser"
)

// OpenURL opens url in a new browser tab.
func OpenURL(url string) error {
	_ = browser.OpenURL(url)
	return nil
}
