package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"
)

func TestGenerateKeyID(t *testing.T) {
	// Generate a test key pair
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

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
	publicKey2, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	keyID3 := GenerateKeyID(publicKey2)
	if keyID == keyID3 {
		t.Error("Different keys should produce different key IDs")
	}
}

func TestGenerateSPKIHash(t *testing.T) {
	// Create a test certificate
	cert := createTestCertificate(t)

	// Generate hash
	hash := GenerateSPKIHash(cert)

	// Verify hash is a hex string
	if len(hash) != 64 { // SHA-256 produces 32 bytes = 64 hex characters
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}

	// Verify it's deterministic
	hash2 := GenerateSPKIHash(cert)
	if hash != hash2 {
		t.Error("Hash generation should be deterministic")
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

	// Verify each hash is valid
	for i, hash := range hashes {
		if len(hash) != 64 {
			t.Errorf("Hash %d: expected length 64, got %d", i, len(hash))
		}
	}
}

func TestSignPayload(t *testing.T) {
	// Generate a test key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create a test payload
	payload := SignablePayload{
		Domain:     "example.com",
		Pins:       []string{"abc123", "def456"},
		Created:    "2025-10-17T08:00:00Z",
		Expires:    "2025-10-17T09:00:00Z",
		TTLSeconds: 3600,
		KeyID:      "testkey1",
		Alg:        "Ed25519",
	}

	// Sign the payload
	signature, err := SignPayload(payload, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign payload: %v", err)
	}

	// Verify signature is base64 encoded
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		t.Fatalf("Signature is not valid base64: %v", err)
	}

	// Verify signature with public key
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	valid := ed25519.Verify(publicKey, payloadJSON, signatureBytes)
	if !valid {
		t.Error("Signature verification failed")
	}
}

func TestSignPayload_Deterministic(t *testing.T) {
	// Generate a test key pair
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	payload := SignablePayload{
		Domain:     "example.com",
		Pins:       []string{"abc123"},
		Created:    "2025-10-17T08:00:00Z",
		Expires:    "2025-10-17T09:00:00Z",
		TTLSeconds: 3600,
		KeyID:      "testkey1",
		Alg:        "Ed25519",
	}

	// Sign twice
	sig1, err := SignPayload(payload, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign payload: %v", err)
	}

	sig2, err := SignPayload(payload, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign payload: %v", err)
	}

	// Ed25519 signatures are deterministic with the same key and payload
	if sig1 != sig2 {
		t.Error("Signatures should be deterministic")
	}
}

// Helper function to create a test certificate
func createTestCertificate(t *testing.T) *x509.Certificate {
	t.Helper()

	// Generate a key pair for the certificate
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
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
