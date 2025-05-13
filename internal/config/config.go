package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port int
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %w", err)
	}

	return &config, nil
}

func New() *Config {
	return &Config{
		Port: 8080,
	}
}
