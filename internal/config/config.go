package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

const FileName = "config.yaml"

type Config struct {
	Monitor MonitorConfig `yaml:"monitor"`
	Target  TargetConfig  `yaml:"target"`
	Logging LoggingConfig `yaml:"logging"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

type MonitorConfig struct {
	Folder       string `yaml:"folder"`
	NotifyLength int    `yaml:"notify_length"`
}

type TargetConfig struct {
	Folder     string `yaml:"folder"`
	Password   string `yaml:"password"`
	Encryption int    `yaml:"encryption"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func Save(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
