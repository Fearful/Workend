package config

import (
	"crypto/tls"
	"fmt"
)

// TLSConfig returns a *tls.Config if TLS cert and key files are configured,
// nil otherwise. Supports self-signed certificates.
func (c *Config) TLSConfig() (*tls.Config, error) {
	if c.TLSCertFile == "" || c.TLSKeyFile == "" {
		return nil, nil
	}
	cert, err := tls.LoadX509KeyPair(c.TLSCertFile, c.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load TLS keypair: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}
