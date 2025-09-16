package config

import (
	"os"
)

type Config struct {
	APIBaseURL string
	DBPath     string
	AuthToken  string
}

func Load() Config {
	cfg := Config{
		APIBaseURL: getenv("HABITS_API_BASE", "http://localhost:8080"),
		DBPath:     getenv("HABITS_DB_PATH", "habits.db"),
		AuthToken:  getenv("HABITS_AUTH_TOKEN", ""),
	}
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
