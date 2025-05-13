package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	var data Config
	sourceLoc, _ := getConfigPath()
	jsonData, err := os.ReadFile(sourceLoc)
	if err != nil {
		return data, fmt.Errorf("error reading source file: %w", err)
	}
	if err2 := json.Unmarshal(jsonData, &data); err2 != nil {
		return data, fmt.Errorf("error unmarshalling JSON: %w", err2)
	}
	return data, nil
}

func (c *Config) SetUser(name string) error {
	c.CurrentUserName = name
	return write(*c)
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to open home directory: %w", err)
	}
	return homeDir + "/" + configFileName, nil
}

func write(cfg Config) error {
	var jsonData []byte
	destLoc, _ := getConfigPath()
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}
	err = os.WriteFile(destLoc, jsonData, 0666)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}
