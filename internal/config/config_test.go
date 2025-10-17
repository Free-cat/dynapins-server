package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"
)

func TestLoad_Success(t *testing.T) {
	// Generate a test ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Marshal the private key to PKCS8 format
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Set up environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("ALLOWED_DOMAINS", "example.com, *.example.org")
	os.Setenv("SIGNATURE_LIFETIME", "2h")
	os.Setenv("PRIVATE_KEY_PEM", string(privateKeyPEM))
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ALLOWED_DOMAINS")
		os.Unsetenv("SIGNATURE_LIFETIME")
		os.Unsetenv("PRIVATE_KEY_PEM")
	}()

	// Load configuration
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if cfg.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Port)
	}

	if len(cfg.AllowedDomains) != 2 {
		t.Errorf("Expected 2 allowed domains, got %d", len(cfg.AllowedDomains))
	}

	if cfg.AllowedDomains[0] != "example.com" {
		t.Errorf("Expected first domain 'example.com', got '%s'", cfg.AllowedDomains[0])
	}

	if cfg.AllowedDomains[1] != "*.example.org" {
		t.Errorf("Expected second domain '*.example.org', got '%s'", cfg.AllowedDomains[1])
	}

	if cfg.SignatureLifetime != 2*time.Hour {
		t.Errorf("Expected signature lifetime 2h, got %v", cfg.SignatureLifetime)
	}

	if cfg.PrivateKey == nil {
		t.Error("Private key should not be nil")
	}

	if cfg.PublicKey == nil {
		t.Error("Public key should not be nil")
	}

	// Validate default timeouts
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("Expected default read timeout 10s, got %v", cfg.ReadTimeout)
	}

	if cfg.CertDialTimeout != 10*time.Second {
		t.Errorf("Expected default cert dial timeout 10s, got %v", cfg.CertDialTimeout)
	}
}

func TestLoad_MissingAllowedDomains(t *testing.T) {
	os.Setenv("PORT", "8080")
	os.Setenv("SIGNATURE_LIFETIME", "1h")
	os.Setenv("PRIVATE_KEY_PEM", "dummy")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("SIGNATURE_LIFETIME")
		os.Unsetenv("PRIVATE_KEY_PEM")
	}()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing ALLOWED_DOMAINS")
	}
}

func TestLoad_MissingPrivateKey(t *testing.T) {
	os.Setenv("PORT", "8080")
	os.Setenv("ALLOWED_DOMAINS", "example.com")
	os.Setenv("SIGNATURE_LIFETIME", "1h")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ALLOWED_DOMAINS")
		os.Unsetenv("SIGNATURE_LIFETIME")
	}()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing PRIVATE_KEY_PEM")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	os.Setenv("PORT", "invalid")
	os.Setenv("ALLOWED_DOMAINS", "example.com")
	os.Setenv("SIGNATURE_LIFETIME", "1h")
	os.Setenv("PRIVATE_KEY_PEM", "dummy")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ALLOWED_DOMAINS")
		os.Unsetenv("SIGNATURE_LIFETIME")
		os.Unsetenv("PRIVATE_KEY_PEM")
	}()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for invalid PORT")
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Generate a test ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	os.Setenv("ALLOWED_DOMAINS", "example.com")
	os.Setenv("PRIVATE_KEY_PEM", string(privateKeyPEM))
	defer func() {
		os.Unsetenv("ALLOWED_DOMAINS")
		os.Unsetenv("PRIVATE_KEY_PEM")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Port)
	}

	if cfg.SignatureLifetime != 1*time.Hour {
		t.Errorf("Expected default signature lifetime 1h, got %v", cfg.SignatureLifetime)
	}
}
