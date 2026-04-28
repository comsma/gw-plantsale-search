package config

import "os"

type Config struct {
	DSN            string
	ReportPassword string
}

func Load() Config {
	return Config{
		DSN:            os.Getenv("DATABASE_URL"),
		ReportPassword: os.Getenv("REPORT_PASSWORD"),
	}
}
