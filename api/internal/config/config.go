package config

import (
	"os"
)

type Config struct {
	Port        string
	DBUrl       string
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DBUrl:       getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/fleetcontrol?sslmode=disable"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTIssuer:   getEnv("JWT_ISSUER", "fleetcontrol"),
		JWTAudience: getEnv("JWT_AUDIENCE", "fleetcontrol-api"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
