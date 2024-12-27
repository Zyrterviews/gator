package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName string = ".gatorconfig.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (cfg Config) SetUser(userName string) {
	if userName == "" {
		panic("No user name set")
	}

	info, err := os.Stat(configFilePath())
	if err != nil {
		panic(err)
	}

	currentConfig := Read()
	currentConfig.CurrentUserName = userName

	data, err := json.Marshal(currentConfig)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(configFilePath(), data, info.Mode()); err != nil {
		panic(err)
	}
}

func Read() Config {
	rawContent, err := os.ReadFile(configFilePath())
	if err != nil {
		panic(err)
	}

	var cfg Config

	if err := json.Unmarshal(rawContent, &cfg); err != nil {
		panic(err)
	}

	return cfg
}

func configFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(home, configFileName)
}
