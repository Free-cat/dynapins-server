package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"
)

func TestGenerateKeyID(t *testing.T) {
	// Generate a test ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Generate key ID
	keyID := GenerateKeyID(publicKey)

	// Verify key ID is 8 characters
	if len(keyID) != 8 {
		t.Errorf("Expected key ID length 8, got %d", len(keyID))
	}

	// Verify it's deterministic
	keyID2 := GenerateKeyID(publicKey)
	if keyID != keyID2 {
		t.Error("Key ID generation should be deterministic")
	}

	// Verify different keys produce different IDs
	privateKey2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	publicKey2 := &privateKey2.PublicKey
	keyID3 := GenerateKeyID(publicKey2)
	if keyID == keyID3 {
		t.Error("Different keys should produce different key IDs")
	}
}

func TestGenerateSPKIHashes(t *testing.T) {
	// Create test certificates
	cert1 := createTestCertificate(t)
	cert2 := createTestCertificate(t)

	certs := []*x509.Certificate{cert1, cert2}

	// Generate hashes
	hashes := GenerateSPKIHashes(certs)

	if len(hashes) != 2 {
		t.Errorf("Expected 2 hashes, got %d", len(hashes))
	}

	// Verify each hash is a valid base64 string
	for i, hash := range hashes {
		// Decode to verify it's valid base64
		_, err := base64.StdEncoding.DecodeString(hash)
		if err != nil {
			t.Errorf("Hash %d is not valid base64: %v", i, err)
		}
		// Base64-encoded SHA-256 hash should be 44 characters
		if len(hash) != 44 {
			t.Errorf("Hash %d: expected length 44, got %d", i, len(hash))
		}
	}
}

func TestGenerateSPKIHashes_SingleCert(t *testing.T) {
	cert := createTestCertificate(t)
	certs := []*x509.Certificate{cert}

	hashes := GenerateSPKIHashes(certs)

	if len(hashes) != 1 {
		t.Errorf("Expected 1 hash, got %d", len(hashes))
	}

	// Verify hash is deterministic
	hashes2 := GenerateSPKIHashes(certs)
	if hashes[0] != hashes2[0] {
		t.Error("SPKI hash generation should be deterministic")
	}
}

func TestGenerateSPKIHashes_EmptyInput(t *testing.T) {
	hashes := GenerateSPKIHashes([]*x509.Certificate{})

	if len(hashes) != 0 {
		t.Errorf("Expected 0 hashes for empty input, got %d", len(hashes))
	}
}

func TestCreateJWS(t *testing.T) {
	// Generate a test ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Test parameters
	keyID := GenerateKeyID(publicKey)
	domain := "example.com"
	pins := []string{"abc123", "def456"}
	ttl := time.Hour

	// Create JWS token
	jwsToken, err := CreateJWS(privateKey, keyID, domain, pins, ttl)
	if err != nil {
		t.Fatalf("Failed to create JWS: %v", err)
	}

	// Verify token is not empty
	if jwsToken == "" {
		t.Fatal("JWS token is empty")
	}

	// Verify token is in compact serialization format (header.payload.signature)
	parts := splitJWS(jwsToken)
	if len(parts) != 3 {
		t.Errorf("Expected 3 parts in JWS token, got %d", len(parts))
	}

	// Decode and verify header
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("Failed to decode JWS header: %v", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		t.Fatalf("Failed to parse JWS header: %v", err)
	}

	if header["alg"] != "ES256" {
		t.Errorf("Expected alg 'ES256', got '%v'", header["alg"])
	}
	if header["kid"] != keyID {
		t.Errorf("Expected kid '%s', got '%v'", keyID, header["kid"])
	}

	// Decode and verify payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("Failed to decode JWS payload: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		t.Fatalf("Failed to parse JWS payload: %v", err)
	}

	if payload["domain"] != domain {
		t.Errorf("Expected domain '%s', got '%v'", domain, payload["domain"])
	}

	// Verify pins array
	pinsArray, ok := payload["pins"].([]interface{})
	if !ok {
		t.Fatal("Pins is not an array")
	}
	if len(pinsArray) != len(pins) {
		t.Errorf("Expected %d pins, got %d", len(pins), len(pinsArray))
	}

	// Verify timestamps
	if _, ok := payload["iat"]; !ok {
		t.Error("Missing iat claim")
	}
	if _, ok := payload["exp"]; !ok {
		t.Error("Missing exp claim")
	}
	if _, ok := payload["ttl_seconds"]; !ok {
		t.Error("Missing ttl_seconds claim")
	}
}

func TestCreateJWS_WithDifferentInputs(t *testing.T) {
	// Generate a test ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	keyID := GenerateKeyID(publicKey)
	domain := "example.com"
	pins := []string{"abc123"}
	ttl := time.Hour

	// Create JWS with first domain
	jws1, err := CreateJWS(privateKey, keyID, domain, pins, ttl)
	if err != nil {
		t.Fatalf("Failed to create JWS: %v", err)
	}

	// Create JWS with different domain
	jws2, err := CreateJWS(privateKey, keyID, "different.com", pins, ttl)
	if err != nil {
		t.Fatalf("Failed to create JWS: %v", err)
	}

	// JWS tokens should be different due to different domains
	if jws1 == jws2 {
		t.Error("JWS tokens with different domains should be different")
	}
}

func TestCreateJWS_KidHeader(t *testing.T) {
	// Generate a test ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Test with specific key ID
	expectedKeyID := GenerateKeyID(publicKey)
	domain := "example.com"
	pins := []string{"abc123"}
	ttl := time.Hour

	// Create JWS token
	jwsToken, err := CreateJWS(privateKey, expectedKeyID, domain, pins, ttl)
	if err != nil {
		t.Fatalf("Failed to create JWS: %v", err)
	}

	// Split and decode header
	parts := splitJWS(jwsToken)
	if len(parts) != 3 {
		t.Fatalf("Expected 3 parts in JWS token, got %d", len(parts))
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("Failed to decode JWS header: %v", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		t.Fatalf("Failed to parse JWS header: %v", err)
	}

	// Verify kid matches expected value
	actualKid, ok := header["kid"].(string)
	if !ok {
		t.Fatal("kid header is missing or not a string")
	}

	if actualKid != expectedKeyID {
		t.Errorf("Expected kid '%s', got '%s'", expectedKeyID, actualKid)
	}
}

func TestGetPublicKeyFromPrivate(t *testing.T) {
	// Generate a test ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	publicKey := GetPublicKeyFromPrivate(privateKey)

	// Verify public key is not nil
	if publicKey == nil {
		t.Fatal("Public key is nil")
	}

	// Verify public key matches the one from private key
	if publicKey != &privateKey.PublicKey {
		t.Error("Public key doesn't match private key's public key")
	}
}

// Helper function to split JWS compact serialization
func splitJWS(jws string) []string {
	parts := make([]string, 0, 3)
	start := 0
	for i := 0; i < len(jws); i++ {
		if jws[i] == '.' {
			parts = append(parts, jws[start:i])
			start = i + 1
		}
	}
	if start < len(jws) {
		parts = append(parts, jws[start:])
	}
	return parts
}

// Helper function to create a test certificate
func createTestCertificate(t *testing.T) *x509.Certificate {
	t.Helper()

	// Generate RSA key pair for the certificate
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			CommonName:   "test.example.com",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	return cert
}
