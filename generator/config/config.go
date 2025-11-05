package config

import (
	"time"
)

type Config struct {
	Port        int           // Port
	SnapshotTTL time.Duration // Snapshot TTL
}

func NewConfig() *Config {
	return &Config{
		Port:        9001,
		SnapshotTTL: 10 * time.Second,
	}
}
