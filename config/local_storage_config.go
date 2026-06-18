//go:build js && wasm

package config

import (
	"syscall/js"
)

// configKey is the localStorage key under which the config is stored.
const configKey = "goplease.config"

// localStorageConfig implements manager by reading and writing the config
// to the browser's localStorage; used on WASM builds.
type localStorageConfig struct{}

// newConfigManager returns a manager that loads and saves the config using
// the browser's localStorage.
func newConfigManager() manager {
	return &localStorageConfig{}
}

// load reads the config from localStorage. If no value is stored yet, it
// returns the embedded default config instead.
func (*localStorageConfig) load() ([]byte, error) {
	ls := js.Global().Get("localStorage")

	jsValue := ls.Call("getItem", configKey)
	if jsValue.IsNull() {
		return defaultConfig, nil
	}

	return []byte(jsValue.String()), nil
}

// save writes data to localStorage under configKey.
func (*localStorageConfig) save(data []byte) error {
	localStorage := js.Global().Get("localStorage")
	localStorage.Call("setItem", configKey, string(data))

	return nil
}
