package config

import "os"

// Config holds all application configuration loaded from the environment.
type Config struct {
	Env         string
	Port        string
	DatabaseURL string
}

// Load reads configuration from environment variables, applying sane defaults.
func Load() *Config {
	return &Config{
		Env:         getEnv("APP_ENV", "development"),
		Port:        getEnv("APP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/wolf_platform?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
