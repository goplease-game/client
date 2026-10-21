//go:build js

package user

import "syscall/js"

const localStorageKeyToken = "goplease_auth_token"
const localStorageKeyUsername = "goplease_auth_username"

func loadPersistedSession() (token, username string, ok bool) {
	ls := js.Global().Get("localStorage")
	t := ls.Call("getItem", localStorageKeyToken)
	u := ls.Call("getItem", localStorageKeyUsername)

	if t.IsNull() || t.String() == "" {
		return "", "", false
	}

	return t.String(), u.String(), true
}

func persistSession(token, username string) {
	ls := js.Global().Get("localStorage")
	ls.Call("setItem", localStorageKeyToken, token)
	ls.Call("setItem", localStorageKeyUsername, username)
}

func clearPersistedSession() {
	ls := js.Global().Get("localStorage")
	ls.Call("removeItem", localStorageKeyToken)
	ls.Call("removeItem", localStorageKeyUsername)
}
