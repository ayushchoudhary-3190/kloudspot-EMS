package config

import (
	"os"
	"strconv"
)

// Config holds runtime configuration sourced from environment variables.
type Config struct {
	Port          string
	DBPath        string
	JWTSecret     string
	JWTExpiryHrs  int
	AdminUsername string
	AdminPassword string
	CORSOrigins   string
}

// Load reads configuration from the environment, applying sensible defaults.
func Load() Config {
	return Config{
		Port:          getenv("PORT", "8080"),
		DBPath:        getenv("DB_PATH", "ems.db"),
		JWTSecret:     getenv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiryHrs:  getenvInt("JWT_EXPIRY_HOURS", 24),
		AdminUsername: getenv("ADMIN_USERNAME", "admin"),
		AdminPassword: getenv("ADMIN_PASSWORD", "admin123"),
		CORSOrigins:   getenv("CORS_ORIGINS", "http://localhost:5173,http://localhost:4173,http://localhost:3000"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
