package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	Env         string // "development" or "production"
	LogLevel    string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "3000"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/usersdb?sslmode=disable"),
		Env:         getEnv("APP_ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
