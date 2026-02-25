package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Token          string `toml:"token"`
	DefaultStation string `toml:"default_station"`
}

func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".metro.toml"
	}
	return filepath.Join(home, ".metro.toml")
}

func Load() (*Config, error) {
	cfg := &Config{}
	path := Path()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	_, err := toml.DecodeFile(path, cfg)
	return cfg, err
}

func Save(cfg *Config) error {
	f, err := os.Create(Path())
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}
