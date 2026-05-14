package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	PocketBaseUrl          string
	ServiceAccountEmail    string
	ServiceAccountPassword string
	SeekConfigFile         string
	JoraConfigFile         string
	ErrorLogFile           string
	SmtpDomain             string
	SmtpPort               int
	SenderEmail            string
	SenderEmailPassword    string
	EmailRecipient         string
}

// Load reads runtask.json from the directory containing the executable.
func Load() (Config, error) {
	exe, err := os.Executable()
	if err != nil {
		return Config{}, fmt.Errorf("resolve executable path: %w", err)
	}
	configPath := filepath.Join(filepath.Dir(exe), "runtask.json")
	f, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("open %s: %w", configPath, err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}
