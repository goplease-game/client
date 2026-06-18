//go:build js && wasm

package game

import "syscall/js"

// OpenURL opens url in a new browser tab via the JavaScript window.open API.
func OpenURL(url string) error {
	js.Global().Call("open", url, "_blank")
	return nil
}
