package config

import (
	"time"
)

type Config struct {
	Port     int           `yaml:"port"`     // Default port
	CacheTTL time.Duration `yaml:"cacheTTL"` // Default cache TTL
}

func NewConfig() *Config {
	return &Config{
		Port:     9001,
		CacheTTL: 10 * time.Second,
	}
}
