package config

import (
	"errors"
	"os"
)

type Config struct {
	ListenAddr     string
	DatabaseURL    string
	DaggerSockPath string
	ReposRoot      string
	CookieSecure   bool
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:     getEnv("WORKEND_LISTEN_ADDR", ":8080"),
		DatabaseURL:    os.Getenv("WORKEND_DATABASE_URL"),
		DaggerSockPath: getEnv("WORKEND_DAGGER_SOCK", "/run/dagger/buildkitd.sock"),
		ReposRoot:      getEnv("WORKEND_REPOS_ROOT", "/repos"),
		CookieSecure:   getEnv("WORKEND_COOKIE_SECURE", "false") == "true",
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("WORKEND_DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
