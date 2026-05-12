package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed config_defaults.yaml
var defaultConfig []byte

const (
	configFilename      = "config.yaml"
	osUserConfigGameDir = "goplease"
)

type ConfigT struct {
	Resolution string `yaml:"resolution"`
	WindowW    int    `yaml:"-"`
	WindowH    int    `yaml:"-"`

	ServerAddr string `yaml:"server_addr"`

	MockClient  bool `yaml:"mock_client"`
	LogProtocol bool `yaml:"log_protocol"`
}

var loadConfigOnce sync.Once
var loadedConfig *ConfigT

// Get returns config from user config dir or from config_defaults.yaml
func Get() *ConfigT {
	loadConfigOnce.Do(func() {
		var err error
		conf, err := loadConfig()
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

func resolveConfig() ([]byte, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(dir, osUserConfigGameDir)
	configPath := filepath.Join(configDir, configFilename)
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		err = fmt.Errorf("read file '%s': %w", configPath, err)
		return nil, err
	}

	return data, nil
}

// ConfigFromFile returns new config from given YAML file.
func loadConfig() (*ConfigT, error) {
	data, err := resolveConfig()
	if err != nil {
		return nil, fmt.Errorf("loadConfig: %w", err)
	}

	conf := new(ConfigT)
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return nil, err
	}

	return conf, err
}

func parseResolution(res string) (width, height int, err error) {
	_, err = fmt.Sscanf(res, "%dx%d", &width, &height)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid resolution format: %w (expecting '800x600', got '%s')", err, res)
	}
	return width, height, nil
}
