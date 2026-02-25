package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type SavedPlace struct {
	Name string  `toml:"name"`
	Type string  `toml:"type"` // "StopArea" or "Address"
	ID   string  `toml:"id"`   // stop area ID (for StopArea type)
	City string  `toml:"city"`
	Lat  float64 `toml:"lat,omitempty"`
	Lon  float64 `toml:"lon,omitempty"`
}

type Config struct {
	DefaultPlace string                `toml:"default_place"`
	Places       map[string]SavedPlace `toml:"places"`
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
	f, err := os.OpenFile(Path(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}
