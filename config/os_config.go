//go:build !js || !wasm

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// osConfig implements manager by reading and writing the config file in the
// OS user config directory; used on non-WASM builds.
type osConfig struct{}

// newConfigManager returns a manager that loads and saves the config file
// from the OS user config directory.
func newConfigManager() manager {
	return &osConfig{}
}

// load reads the config file from the OS user config directory. If the file
// does not exist, it returns the embedded default config instead.
func (*osConfig) load() ([]byte, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(dir, osUserConfigGameDir)
	configPath := filepath.Join(configDir, configFilename)
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		return configDefaultYaml, nil
	}

	data, err := os.ReadFile(configPath) //nolint:gosec
	if err != nil {
		err = fmt.Errorf("read file '%s': %w", configPath, err)
		return nil, err
	}

	return data, nil
}

// save writes data to the config file in the OS user config directory,
// creating the directory if it does not already exist.
func (*osConfig) save(data []byte) error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	configDir := filepath.Join(dir, osUserConfigGameDir)
	err = os.MkdirAll(configDir, 0o750)
	if err != nil {
		return fmt.Errorf("save config: mkdir %s: %w", configDir, err)
	}

	configPath := filepath.Join(configDir, configFilename)
	err = os.WriteFile(configPath, data, 0o600)
	if err != nil {
		return fmt.Errorf("save config: write %s: %w", configPath, err)
	}

	return nil
}
