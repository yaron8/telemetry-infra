package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port  int // Port
	Redis RedisConfig
	ETL   ETLConfig
}

type RedisConfig struct {
	Host string
	Port int
	TTL  time.Duration
}

type ETLConfig struct {
	Interval     time.Duration
	GeneratorURL string
}

func NewConfig() *Config {
	// Read Redis host from environment variable, default to localhost
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	// Read Redis port from environment variable, default to 6379
	redisPort := 6379
	if redisPortStr := os.Getenv("REDIS_PORT"); redisPortStr != "" {
		if port, err := strconv.Atoi(redisPortStr); err == nil {
			redisPort = port
		}
	}

	// Read generator URL from environment variable, default to localhost
	generatorURL := os.Getenv("GENERATOR_URL")
	if generatorURL == "" {
		generatorURL = "http://localhost:9001"
	}

	return &Config{
		Port: 8080,
		Redis: RedisConfig{
			Host: redisHost,
			Port: redisPort,
			TTL:  30 * time.Second,
		},
		ETL: ETLConfig{
			Interval:     10 * time.Second,
			GeneratorURL: generatorURL,
		},
	}
}
