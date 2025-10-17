package config

import (
	"crypto/ecdsa"
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
	Port              int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	ReadHeaderTimeout time.Duration
	MaxHeaderBytes    int

	// Domain and security configuration
	AllowedDomains    []string
	SignatureLifetime time.Duration
	PrivateKey        *ecdsa.PrivateKey
	PublicKey         *ecdsa.PublicKey
	AllowIPLiterals   bool

	// Certificate retrieval configuration
	CertDialTimeout time.Duration
	CertCacheTTL    time.Duration

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

	cfg.ReadHeaderTimeout, err = getEnvDuration("READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid READ_HEADER_TIMEOUT: %w", err)
	}

	cfg.MaxHeaderBytes, err = getEnvInt("MAX_HEADER_BYTES", 1<<20) // 1MB default
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_HEADER_BYTES: %w", err)
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
	cfg.PublicKey = &privateKey.PublicKey

	cfg.AllowIPLiterals = getEnvBool("ALLOW_IP_LITERALS", false)

	// Certificate retrieval configuration
	cfg.CertDialTimeout, err = getEnvDuration("CERT_DIAL_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid CERT_DIAL_TIMEOUT: %w", err)
	}

	cfg.CertCacheTTL, err = getEnvDuration("CERT_CACHE_TTL", 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid CERT_CACHE_TTL: %w", err)
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

// getEnvBool retrieves a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	// Accept: true/false, 1/0, yes/no, on/off (case insensitive)
	valueStr = strings.ToLower(strings.TrimSpace(valueStr))
	switch valueStr {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// parsePrivateKey parses an ECDSA P-256 private key from PEM format
func parsePrivateKey(pemData string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// Try parsing as PKCS8 (preferred format)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if ecdsaKey, ok := key.(*ecdsa.PrivateKey); ok {
			return ecdsaKey, nil
		}
		return nil, errors.New("private key is not ECDSA")
	}

	// Try parsing as SEC1 EC private key
	ecKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err == nil {
		return ecKey, nil
	}

	return nil, fmt.Errorf("unsupported private key format (expected ECDSA P-256): %w", err)
}
