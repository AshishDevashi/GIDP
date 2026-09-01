package config

import (
	"os"
	"time"
)

// Config holds all application configuration loaded from the environment.
type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
	GitHubToken string
}

// Load reads configuration from environment variables, applying sane defaults.
func Load() *Config {
	return &Config{
		Env:         getEnv("APP_ENV", "development"),
		Port:        getEnv("APP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/wolf_platform?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:      getEnvDuration("JWT_TTL", 24*time.Hour),
		GitHubToken: getEnv("GITHUB_TOKEN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
