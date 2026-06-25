//go:build js && wasm

// Package clipboard ...
package clipboard

import "syscall/js"

// Write copies text to the browser clipboard using the Clipboard API.
func Write(text string) {
	js.Global().Get("navigator").Get("clipboard").Call("writeText", text)
}

// Read asynchronously reads from the browser clipboard and calls cb with the result.
func Read(cb func(string)) {
	js.Global().Get("navigator").Get("clipboard").Call("readText").Call("then",
		js.FuncOf(func(_ js.Value, args []js.Value) any {
			cb(args[0].String())
			return nil
		}),
	)
}
