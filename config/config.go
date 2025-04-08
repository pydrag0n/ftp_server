package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	RootPath = "./files"
	TemplatePath = "templates/*.html"
	IconPath = "static/icon"
	ServerPort = ":1212"
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

	if config.FTP.RootPath, err = filepath.Abs(config.FTP.RootPath); err != nil {
		return nil, fmt.Errorf("[ERROR] invalid FTP path: %w", err)
	}

	if config.Template.Path, err = filepath.Abs(config.Template.Path); err != nil {
		return nil, fmt.Errorf("[ERROR] invalid template path: %w", err)
	}

	return &config, nil
}
