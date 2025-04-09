package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Server ServerConfig `json:"server"`
	FTP FTPConfig `json:"FTP"`
	Template TemplateConfig `json:"template"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int `json:"port"`
	Timeout int `json:"timeout"`
	Debug bool `json:"debug"`
}

type FTPConfig struct {
	RootPath string `json:"root_path"`
}

type TemplateConfig struct {
	Path string `json:"path"`
}

func Load(path string) (*Config, error) {
	file, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("[ERROR] read file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("[ERROR] parse JSON: %w", err)
	}

	return &config, nil
}
