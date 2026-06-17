// Package config ...
package config

import (
	_ "embed"
	"fmt"
	"runtime"
	"sync"

	"gopkg.in/yaml.v3"
)

// defaultConfig holds the embedded contents of config_defaults.yaml, used as
// the source of default values when no user config file exists yet.
//
//go:embed config_defaults.yaml
var defaultConfig []byte

const (
	// configFilename is the name of the config file stored in the user config directory.
	configFilename = "config.yaml"
	// osUserConfigGameDir is the name of the application's subdirectory within the OS user config directory.
	osUserConfigGameDir = "goplease"
)

// manager abstracts platform-specific loading and saving of the raw config bytes.
type manager interface {
	// load returns the raw config bytes, e.g. read from the user config directory.
	load() (data []byte, err error)
	// save persists the given raw config bytes.
	save(data []byte) error
}

// Config holds the application's runtime configuration, combining values
// loaded from YAML with values derived at load time (e.g. window dimensions).
type Config struct {
	Resolution string `yaml:"resolution"`
	WindowW    int    `yaml:"-"`
	WindowH    int    `yaml:"-"`

	// ServerAddr is the address of the game server to connect to.
	ServerAddr string `yaml:"server_addr"`

	// Volume is the audio volume level.
	Volume float64 `yaml:"volume"`

	DevMode struct {
		Enabled     bool `yaml:"enabled"`
		MockClient  bool `yaml:"mock_client"`
		LogProtocol bool `yaml:"log_protocol"`
	} `yaml:"dev_mode"`
}

// loadConfigOnce ensures the config is loaded and parsed only once.
var loadConfigOnce sync.Once

// loadedConfig caches the config loaded by Get; nil until the first call.
var loadedConfig *Config

// confM is the platform-specific manager used to load and save the config.
// It is nil until the first call to Get, which initializes it via newConfigManager.
var confM manager

// Get returns the application config, loading and parsing it from the user
// config directory (falling back to config_defaults.yaml) on the first call
// and returning the cached value on subsequent calls. It panics if loading,
// unmarshaling, or resolution parsing fails.
func Get() *Config {
	loadConfigOnce.Do(func() {
		var err error

		confM = newConfigManager()
		data, err := confM.load()
		if err != nil {
			panic(err)
		}

		conf := new(Config)
		err = yaml.Unmarshal(data, conf)
		if err != nil {
			panic(err)
		}

		w, h, err := parseResolution(conf.Resolution)
		if err != nil {
			panic(err)
		}

		conf.WindowW = w
		conf.WindowH = h

		loadedConfig = conf
	})

	return loadedConfig
}

// Save marshals the current config to YAML and persists it via the
// platform-specific manager. It calls Get to obtain the config, which also
// guarantees the manager has been initialized.
func Save() error {
	data, err := yaml.Marshal(Get())
	if err != nil {
		return fmt.Errorf("save config: marshal: %w", err)
	}

	return confM.save(data)
}

// parseResolution parses a resolution string in "WIDTHxHEIGHT" format (e.g.
// "800x600") and returns the corresponding width and height.
func parseResolution(res string) (width, height int, err error) {
	_, err = fmt.Sscanf(res, "%dx%d", &width, &height)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid resolution format: %w (expecting '800x600', got '%s')", err, res)
	}
	return width, height, nil
}

// IsWASM reports whether the binary is running in a WebAssembly (GOARCH=wasm) build.
func IsWASM() bool {
	return runtime.GOARCH == "wasm"
}
