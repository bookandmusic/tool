package common

import (
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

type PluginConfig struct {
	Install   map[string]any `yaml:"install,omitempty"`
	Uninstall map[string]any `yaml:"uninstall,omitempty"`
}

type Executor struct {
	Shell  string `yaml:"shell"`
	Python string `yaml:"python"`
}

type Config struct {
	PluginDirs     []string                `yaml:"plugin_dirs"`
	EnabledPlugins []string                `yaml:"enabled_plugins"`
	Plugins        map[string]PluginConfig `yaml:"plugins"`
	Executor       *Executor               `yaml:"executor"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	path = filepath.Clean(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func GenerateDefault() *Config {
	defaultCfg := &Config{
		PluginDirs:     []string{"./plugins"},
		EnabledPlugins: []string{},
		Plugins:        map[string]PluginConfig{},
		Executor: &Executor{
			Python: "/usr/bin/python3",
			Shell:  "/bin/bash",
		},
	}
	return defaultCfg
}
