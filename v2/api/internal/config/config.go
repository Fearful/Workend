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
	LogsRoot       string
	CookieSecure   bool

	TokenKey string // base64 32-byte key for at-rest encryption

	// OAuth provider configuration is loaded directly by the oauth package
	// from env vars (per-provider) or a JSON file (multi-instance). See
	// oauth.LoadConfig for the env-var names.

	WebPublicURL string
	SMTPHost     string
	SMTPFrom     string
	SMTPUsername string
	SMTPPassword string
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:     getEnv("WORKEND_LISTEN_ADDR", ":8080"),
		DatabaseURL:    os.Getenv("WORKEND_DATABASE_URL"),
		DaggerSockPath: getEnv("WORKEND_DAGGER_SOCK", "/run/dagger/buildkitd.sock"),
		ReposRoot:      getEnv("WORKEND_REPOS_ROOT", "/repos"),
		LogsRoot:       getEnv("WORKEND_LOGS_ROOT", "/var/lib/workend/logs"),
		CookieSecure:   getEnv("WORKEND_COOKIE_SECURE", "false") == "true",

		TokenKey: os.Getenv("WORKEND_TOKEN_KEY"),

		WebPublicURL: getEnv("WORKEND_WEB_PUBLIC_URL", "http://localhost:3000"),
		SMTPHost:     os.Getenv("WORKEND_SMTP_HOST"),
		SMTPFrom:     os.Getenv("WORKEND_SMTP_FROM"),
		SMTPUsername: os.Getenv("WORKEND_SMTP_USERNAME"),
		SMTPPassword: os.Getenv("WORKEND_SMTP_PASSWORD"),
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
