package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	ListenAddr     string
	DatabaseURL    string
	DaggerSockPath string
	ReposRoot      string
	LogsRoot       string
	ArtifactsRoot  string
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

	// DefaultTaskTimeoutSec applies to runs whose task has no per-task
	// override. Overrideable via WORKEND_DEFAULT_TASK_TIMEOUT_SECONDS.
	DefaultTaskTimeoutSec int

	// VAPID keys for Web Push. All three must be set for push delivery to
	// activate; the frontend handles the "not configured" path gracefully.
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string // e.g. mailto:admin@example.com

	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCProviderName string

	TLSCertFile string
	TLSKeyFile  string
}

func (c *Config) OIDCConfigured() bool {
	return c.OIDCIssuer != "" && c.OIDCClientID != "" && c.OIDCClientSecret != ""
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:     getEnv("WORKEND_LISTEN_ADDR", ":8080"),
		DatabaseURL:    os.Getenv("WORKEND_DATABASE_URL"),
		DaggerSockPath: getEnv("WORKEND_DAGGER_SOCK", "/run/dagger/buildkitd.sock"),
		ReposRoot:      getEnv("WORKEND_REPOS_ROOT", "/repos"),
		LogsRoot:       getEnv("WORKEND_LOGS_ROOT", "/var/lib/workend/logs"),
		ArtifactsRoot:  getEnv("WORKEND_ARTIFACTS_ROOT", "/var/lib/workend/artifacts"),
		CookieSecure:   getEnv("WORKEND_COOKIE_SECURE", "false") == "true",

		TokenKey: os.Getenv("WORKEND_TOKEN_KEY"),

		WebPublicURL: getEnv("WORKEND_WEB_PUBLIC_URL", "http://localhost:3000"),
		SMTPHost:     os.Getenv("WORKEND_SMTP_HOST"),
		SMTPFrom:     os.Getenv("WORKEND_SMTP_FROM"),
		SMTPUsername: os.Getenv("WORKEND_SMTP_USERNAME"),
		SMTPPassword: os.Getenv("WORKEND_SMTP_PASSWORD"),

		DefaultTaskTimeoutSec: getEnvInt("WORKEND_DEFAULT_TASK_TIMEOUT_SECONDS", 1800),

		VAPIDPublicKey:  os.Getenv("WORKEND_VAPID_PUBLIC"),
		VAPIDPrivateKey: os.Getenv("WORKEND_VAPID_PRIVATE"),
		VAPIDSubject:    getEnv("WORKEND_VAPID_SUBJECT", "mailto:admin@workend.local"),

		OIDCIssuer:       os.Getenv("WORKEND_OIDC_ISSUER"),
		OIDCClientID:     os.Getenv("WORKEND_OIDC_CLIENT_ID"),
		OIDCClientSecret: os.Getenv("WORKEND_OIDC_CLIENT_SECRET"),
		OIDCProviderName: getEnv("WORKEND_OIDC_PROVIDER_NAME", "SSO"),

		TLSCertFile: os.Getenv("WORKEND_TLS_CERT_FILE"),
		TLSKeyFile:  os.Getenv("WORKEND_TLS_KEY_FILE"),
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

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
