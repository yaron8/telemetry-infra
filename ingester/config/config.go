package config

type Config struct {
	Port int // Port
}

func NewConfig() *Config {
	return &Config{
		Port: 8080,
	}
}
