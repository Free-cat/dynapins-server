package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pinning-server/internal/config"
	"pinning-server/internal/models"
)

func TestHandleGetPins_MethodNotAllowed(t *testing.T) {
	server := createTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/pins?domain=example.com", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleGetPins_MissingDomain(t *testing.T) {
	server := createTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/pins", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var errorResp models.Error
	if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResp.Code != http.StatusBadRequest {
		t.Errorf("Expected error code %d, got %d", http.StatusBadRequest, errorResp.Code)
	}
}

func TestHandleGetPins_DomainNotInWhitelist(t *testing.T) {
	server := createTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/pins?domain=notallowed.com", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var errorResp models.Error
	if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResp.Code != http.StatusForbidden {
		t.Errorf("Expected error code %d, got %d", http.StatusForbidden, errorResp.Code)
	}
}

func TestHandleGetPins_Success(t *testing.T) {
	server := createTestServer(t)

	// Use a real domain that should be accessible
	req := httptest.NewRequest(http.MethodGet, "/v1/pins?domain=google.com", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var envelope models.PinEnvelope
	if err := json.NewDecoder(w.Body).Decode(&envelope); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Validate response structure
	if envelope.Domain != "google.com" {
		t.Errorf("Expected domain 'google.com', got '%s'", envelope.Domain)
	}

	if len(envelope.Pins) == 0 {
		t.Error("Expected at least one pin")
	}

	if envelope.KeyID == "" {
		t.Error("KeyID should not be empty")
	}

	if envelope.Alg != "Ed25519" {
		t.Errorf("Expected algorithm 'Ed25519', got '%s'", envelope.Alg)
	}

	if envelope.Signature == "" {
		t.Error("Signature should not be empty")
	}

	if envelope.TTLSeconds <= 0 {
		t.Errorf("Expected positive TTL, got %d", envelope.TTLSeconds)
	}

	// Validate timestamps
	created, err := time.Parse(time.RFC3339, envelope.Created)
	if err != nil {
		t.Errorf("Invalid created timestamp: %v", err)
	}

	expires, err := time.Parse(time.RFC3339, envelope.Expires)
	if err != nil {
		t.Errorf("Invalid expires timestamp: %v", err)
	}

	if !expires.After(created) {
		t.Error("Expires should be after created")
	}
}

func TestHandleGetPins_WildcardDomain(t *testing.T) {
	server := createTestServerWithDomains(t, []string{"*.google.com"})

	// Test subdomain that matches wildcard
	req := httptest.NewRequest(http.MethodGet, "/v1/pins?domain=www.google.com", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for wildcard match, got %d", http.StatusOK, w.Code)
	}

	// Test domain that doesn't match (too many levels)
	req = httptest.NewRequest(http.MethodGet, "/v1/pins?domain=api.v2.google.com", nil)
	w = httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for non-matching wildcard, got %d", http.StatusForbidden, w.Code)
	}
}

// Helper function to create a test server
func createTestServer(t *testing.T) *Server {
	return createTestServerWithDomains(t, []string{"google.com", "example.com"})
}

func createTestServerWithDomains(t *testing.T, domains []string) *Server {
	t.Helper()

	// Generate a test key pair
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    domains,
		SignatureLifetime: 1 * time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         privateKey.Public().(ed25519.PublicKey),
		CertDialTimeout:   10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		LogLevel:          "info",
	}

	return New(cfg)
}
