// Package config ...
package config

import (
	_ "embed"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"gopkg.in/yaml.v3"
)

// configDefaultYaml holds the embedded contents of config_defaults.yaml, used as
// the source of default values when no user config file exists yet.
//
//go:embed config_defaults.yaml
var configDefaultYaml []byte

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
	Fullscreen bool   `yaml:"fullscreen"`

	// Window width & height will be parsed from resolution
	WindowW int `yaml:"-"`
	WindowH int `yaml:"-"`

	// ServerAddr is the address of the game server.
	ServerAddr string `yaml:"server_addr"`
	Secure     bool   `yaml:"secure"`

	// Volume is the audio volume level.
	Volume float64 `yaml:"volume"`

	ShowGameLog       bool `yaml:"show_game_log"`
	AutoShowInfoPanel bool `yaml:"auto_show_info_panel"`

	// SkipTutorial disables tutorial overlays for new practice/scenario games.
	SkipTutorial bool `yaml:"skip_tutorial"`

	// TutorialsCompleted stores completed tutorial chapter names.
	TutorialsCompleted []string `yaml:"tutorials_completed"`

	Keybindings *Keybindings `yaml:"keybindings"`

	DevMode struct {
		Enabled     bool `yaml:"enabled"`
		MockClient  bool `yaml:"mock_client"`
		LogProtocol bool `yaml:"log_protocol"`
	} `yaml:"dev_mode"`
}

// Keybindings defines keyboard shortcuts for game actions.
type Keybindings struct {
	Move            *ebiten.Key `yaml:"move,omitempty"`
	Ability1        *ebiten.Key `yaml:"ability_1,omitempty"`
	Ability2        *ebiten.Key `yaml:"ability_2,omitempty"`
	Ability3        *ebiten.Key `yaml:"ability_3,omitempty"`
	Ability4        *ebiten.Key `yaml:"ability_4,omitempty"`
	ShowGameLog     *ebiten.Key `yaml:"show_game_log,omitempty"`
	ShowCoordinates *ebiten.Key `yaml:"show_coordinates,omitempty"`
	EndTurn         *ebiten.Key `yaml:"end_turn,omitempty"`
}

// loadConfigOnce ensures the config is loaded and parsed only once.
var loadConfigOnce sync.Once
var loadDefaultConfigOnce sync.Once

// loadedConfig caches the config loaded by Get; nil until the first call.
var loadedConfig *Config
var loadedDefaultConfig *Config

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

		normalizeConfig(conf)

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

func defaultConfig() *Config {
	loadDefaultConfigOnce.Do(func() {
		conf := new(Config)
		err := yaml.Unmarshal(configDefaultYaml, conf)
		if err != nil {
			panic(err)
		}
		loadedDefaultConfig = conf
	})

	return loadedDefaultConfig
}

func normalizeConfig(conf *Config) {
	defConf := defaultConfig()
	if isOldFormat(conf.ServerAddr) {
		conf.ServerAddr = defConf.ServerAddr
		conf.Secure = defConf.Secure

		newData, err := yaml.Marshal(conf)
		if err != nil {
			panic(err)
		}

		err = confM.save(newData)
		if err != nil {
			log.Printf("unable to save fixed config: %v", err)
		}
	}

	if conf.Keybindings == nil {
		defConf := defaultConfig()
		conf.Keybindings = defConf.Keybindings
	}

	w, h, err := parseResolution(conf.Resolution)
	if err != nil {
		panic(err)
	}

	conf.WindowW = w
	conf.WindowH = h

	conf.ServerAddr = defConf.ServerAddr
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

// isOldFormat detects a config saved before the server_addr format change.
// It used to be a full URL with scheme and path (e.g. "ws://localhost:8090/play/"),
// now it's just "host:port" (e.g. "localhost:8090"), with the scheme
// (ws/wss, http/https) determined separately by the Secure field.
// The presence of "://" is a reliable marker of the old format.
func isOldFormat(addr string) bool {
	return strings.Contains(addr, "://")
}
