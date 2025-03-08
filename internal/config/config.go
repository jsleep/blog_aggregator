package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Exports

type Config struct {
	DBUrl string `json:"db_url"`
	User  string `json:"user"`
}

func (c *Config) SetUser(user string) error {
	c.User = user
	err := write(*c)
	return err
}

func Read() (Config, error) {
	var cfg Config
	fp, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	file, _ := os.ReadFile(fp)
	// fmt.Printf("Raw file content: %s\n", file)
	_ = json.Unmarshal([]byte(file), &cfg)
	return cfg, nil
}

// Helpers

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	fp := filepath.Join(homeDir, configFileName)
	fmt.Println("Config file path:", fp) // Debugging line
	return fp, nil
}

func write(cfg Config) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	fp, err := getConfigFilePath()
	if err != nil {
		return err
	}
	return os.WriteFile(fp, b, 0644)
}
