package config

import (
	"os"
	"encoding/json"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DB_URL        string `json:"db_url"`
	Curr_Username string `json:"current_user_name"`
}

func Read() (Config, error) {
	cfg_path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(cfg_path)
	if err != nil {
		return Config{}, err
	}

	var res_cfg Config
	if err := json.Unmarshal(data, &res_cfg); err != nil {
		return Config{}, err
	}

	return res_cfg, nil
}

func (c *Config) SetUser(new_name string) {
	c.Curr_Username = new_name

	write(*c)
}

func getConfigFilePath() (string, error) {
	sub_path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	res_path := sub_path + "/" + configFileName

	return res_path, nil
}

func write(cfg Config) error {
	cfg_path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	new_data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(cfg_path, new_data, 0644); err != nil {
		return err
	}

	return nil
} 
