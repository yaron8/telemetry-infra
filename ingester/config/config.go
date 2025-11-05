package config

import "time"

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
	Interval time.Duration
}

func NewConfig() *Config {
	return &Config{
		Port: 8080,
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			TTL:  10 * time.Second,
		},
		ETL: ETLConfig{
			Interval: 10 * time.Second,
		},
	}
}
