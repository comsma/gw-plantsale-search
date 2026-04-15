package config

import "os"

type Config struct {
	DSN string
}

func Load() Config {
	return Config{
		DSN: os.Getenv("DATABASE_URL"),
	}
}
