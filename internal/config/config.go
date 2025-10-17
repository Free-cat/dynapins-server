package config

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	// Domain and security configuration
	AllowedDomains    []string
	SignatureLifetime time.Duration
	PrivateKey        ed25519.PrivateKey
	PublicKey         ed25519.PublicKey

	// Certificate retrieval configuration
	CertDialTimeout time.Duration

	// Logging configuration
	LogLevel string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Server configuration
	var err error
	cfg.Port, err = getEnvInt("PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	cfg.ReadTimeout, err = getEnvDuration("READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid READ_TIMEOUT: %w", err)
	}

	cfg.WriteTimeout, err = getEnvDuration("WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid WRITE_TIMEOUT: %w", err)
	}

	cfg.IdleTimeout, err = getEnvDuration("IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid IDLE_TIMEOUT: %w", err)
	}

	cfg.ShutdownTimeout, err = getEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
	}

	// Domain and security configuration
	allowedDomainsStr := os.Getenv("ALLOWED_DOMAINS")
	if allowedDomainsStr == "" {
		return nil, errors.New("ALLOWED_DOMAINS environment variable is required")
	}
	cfg.AllowedDomains = strings.Split(allowedDomainsStr, ",")
	for i := range cfg.AllowedDomains {
		cfg.AllowedDomains[i] = strings.TrimSpace(cfg.AllowedDomains[i])
	}

	cfg.SignatureLifetime, err = getEnvDuration("SIGNATURE_LIFETIME", 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("invalid SIGNATURE_LIFETIME: %w", err)
	}

	privateKeyPEM := os.Getenv("PRIVATE_KEY_PEM")
	if privateKeyPEM == "" {
		return nil, errors.New("PRIVATE_KEY_PEM environment variable is required")
	}

	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PRIVATE_KEY_PEM: %w", err)
	}
	cfg.PrivateKey = privateKey
	cfg.PublicKey = privateKey.Public().(ed25519.PublicKey)

	// Certificate retrieval configuration
	cfg.CertDialTimeout, err = getEnvDuration("CERT_DIAL_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid CERT_DIAL_TIMEOUT: %w", err)
	}

	// Logging configuration
	cfg.LogLevel = getEnvString("LOG_LEVEL", "info")

	return cfg, nil
}

// getEnvString retrieves a string environment variable with a default value
func getEnvString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvInt retrieves an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// getEnvDuration retrieves a duration environment variable with a default value
func getEnvDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// parsePrivateKey parses an Ed25519 private key from PEM format
func parsePrivateKey(pemData string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// Try parsing as PKCS8
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if ed25519Key, ok := key.(ed25519.PrivateKey); ok {
			return ed25519Key, nil
		}
		return nil, errors.New("private key is not Ed25519")
	}

	// If PKCS8 fails, try as raw Ed25519 (some tools export this way)
	if len(block.Bytes) == ed25519.PrivateKeySize {
		return ed25519.PrivateKey(block.Bytes), nil
	}

	return nil, fmt.Errorf("unsupported private key format: %w", err)
}
