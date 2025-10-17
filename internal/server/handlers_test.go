package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pinning-server/internal/cert"
	"pinning-server/internal/config"
	"pinning-server/internal/models"
)

func TestHandleGetPins_MethodNotAllowed(t *testing.T) {
	server, _ := createTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/pins?domain=example.com", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleGetPins_MissingDomain(t *testing.T) {
	server, _ := createTestServer(t)

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
	server, _ := createTestServer(t)

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
	server, retriever := createTestServerWithFakeRetriever(t, []string{"example.com"})

	// Setup fake certificates
	cert, err := cert.GenerateTestCertificate("example.com")
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}
	retriever.SetCertificates("example.com", []*x509.Certificate{cert})

	req := httptest.NewRequest(http.MethodGet, "/v1/pins?domain=example.com", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Parse JWS response
	var jwsResp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&jwsResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Validate response structure
	jwsToken, ok := jwsResp["jws"]
	if !ok {
		t.Fatal("Missing 'jws' field in response")
	}

	if jwsToken == "" {
		t.Fatal("JWS token is empty")
	}

	// Verify token is in compact serialization format (header.payload.signature)
	parts := strings.Split(jwsToken, ".")
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
	if header["kid"] == nil || header["kid"] == "" {
		t.Error("Missing or empty kid in header")
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

	// Validate payload claims
	if payload["domain"] != "example.com" {
		t.Errorf("Expected domain 'example.com', got '%v'", payload["domain"])
	}

	// Verify pins array
	pinsArray, ok := payload["pins"].([]interface{})
	if !ok {
		t.Fatal("Pins is not an array")
	}
	if len(pinsArray) == 0 {
		t.Error("Expected at least one pin")
	}

	// Verify timestamps
	iat, ok := payload["iat"].(float64)
	if !ok {
		t.Fatal("Missing or invalid iat claim")
	}

	exp, ok := payload["exp"].(float64)
	if !ok {
		t.Fatal("Missing or invalid exp claim")
	}

	if exp <= iat {
		t.Error("exp should be after iat")
	}

	// Verify ttl_seconds
	ttlSeconds, ok := payload["ttl_seconds"].(float64)
	if !ok {
		t.Fatal("Missing or invalid ttl_seconds claim")
	}
	if ttlSeconds <= 0 {
		t.Errorf("Expected positive TTL, got %v", ttlSeconds)
	}
}

func TestHandleGetPins_WildcardDomain(t *testing.T) {
	server, retriever := createTestServerWithFakeRetrieverAndDomains(t, []string{"*.example.com"})

	// Setup fake certificates
	cert, err := cert.GenerateTestCertificate("www.example.com")
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}
	retriever.SetCertificates("www.example.com", []*x509.Certificate{cert})

	// Test subdomain that matches wildcard
	req := httptest.NewRequest(http.MethodGet, "/v1/pins?domain=www.example.com", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d for wildcard match, got %d", http.StatusOK, w.Code)
	}

	// Test domain that doesn't match (too many levels)
	req = httptest.NewRequest(http.MethodGet, "/v1/pins?domain=api.v2.example.com", nil)
	w = httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for non-matching wildcard, got %d", http.StatusForbidden, w.Code)
	}
}

// Helper function to create a test server with fake retriever
func createTestServer(t *testing.T) (*Server, *cert.FakeRetriever) {
	return createTestServerWithFakeRetriever(t, []string{"example.com"})
}

func createTestServerWithFakeRetriever(t *testing.T, domains []string) (*Server, *cert.FakeRetriever) {
	return createTestServerWithFakeRetrieverAndDomains(t, domains)
}

func createTestServerWithFakeRetrieverAndDomains(t *testing.T, domains []string) (*Server, *cert.FakeRetriever) {
	t.Helper()

	// Generate a test ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    domains,
		SignatureLifetime: 1 * time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		LogLevel:          "error", // Reduce noise in tests
	}

	// Create fake retriever
	fakeRetriever := cert.NewFakeRetriever()

	return NewWithRetriever(cfg, fakeRetriever), fakeRetriever
}

// TestHandleGetPins_BackupPins tests the include-backup-pins parameter
func TestHandleGetPins_BackupPins(t *testing.T) {
	server, retriever := createTestServerWithFakeRetriever(t, []string{"example.com"})

	// Generate a chain with leaf and intermediate certificates
	chain, err := cert.GenerateTestCertificateChain("example.com")
	if err != nil {
		t.Fatalf("Failed to generate test certificate chain: %v", err)
	}
	retriever.SetCertificates("example.com", chain)

	tests := []struct {
		name             string
		includeBackup    string
		expectedPinCount int
	}{
		{
			name:             "without_backup_pins",
			includeBackup:    "false",
			expectedPinCount: 1, // Only leaf cert
		},
		{
			name:             "with_backup_pins",
			includeBackup:    "true",
			expectedPinCount: 2, // Leaf + intermediate
		},
		{
			name:             "default_no_param",
			includeBackup:    "",
			expectedPinCount: 1, // Default is only leaf
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/v1/pins?domain=example.com"
			if tt.includeBackup != "" {
				url += "&include-backup-pins=" + tt.includeBackup
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			// Parse JWS and decode payload
			var jwsResp map[string]string
			if err := json.NewDecoder(w.Body).Decode(&jwsResp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			jwsToken := jwsResp["jws"]
			parts := strings.Split(jwsToken, ".")
			if len(parts) != 3 {
				t.Fatalf("Invalid JWS format")
			}

			payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err != nil {
				t.Fatalf("Failed to decode payload: %v", err)
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(payloadJSON, &payload); err != nil {
				t.Fatalf("Failed to parse payload: %v", err)
			}

			pinsArray, ok := payload["pins"].([]interface{})
			if !ok {
				t.Fatal("Pins is not an array")
			}

			if len(pinsArray) != tt.expectedPinCount {
				t.Errorf("Expected %d pins, got %d", tt.expectedPinCount, len(pinsArray))
			}
		})
	}
}

// TestHandleGetPins_RetrieverErrors tests error handling from certificate retrieval
func TestHandleGetPins_RetrieverErrors(t *testing.T) {
	tests := []struct {
		name           string
		setupError     error
		expectedStatus int
	}{
		{
			name:           "connection_failed",
			setupError:     fmt.Errorf("failed to connect to example.com: connection refused"),
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "timeout",
			setupError:     fmt.Errorf("dial tcp: i/o timeout"),
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "tls_handshake_failed",
			setupError:     fmt.Errorf("tls: handshake failure"),
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, retriever := createTestServerWithFakeRetriever(t, []string{"example.com"})

			// Set error to return
			retriever.SetError(tt.setupError)

			req := httptest.NewRequest(http.MethodGet, "/v1/pins?domain=example.com", nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var errorResp models.Error
			if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}

			if errorResp.Code != tt.expectedStatus {
				t.Errorf("Expected error code %d, got %d", tt.expectedStatus, errorResp.Code)
			}
		})
	}
}

// TestHandleGetPins_IPLiterals tests IP address rejection
func TestHandleGetPins_IPLiterals(t *testing.T) {
	tests := []struct {
		name            string
		domain          string
		allowIPLiterals bool
		expectedStatus  int
	}{
		{
			name:            "ipv4_rejected",
			domain:          "192.168.1.1",
			allowIPLiterals: false,
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "ipv6_rejected",
			domain:          "2001:db8::1",
			allowIPLiterals: false,
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "ipv6_brackets_rejected",
			domain:          "[2001:db8::1]",
			allowIPLiterals: false,
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "ipv4_allowed_with_flag",
			domain:          "192.168.1.1",
			allowIPLiterals: true,
			expectedStatus:  http.StatusUnprocessableEntity, // Will fail on cert retrieval, but passes validation
		},
		{
			name:            "localhost_ipv4",
			domain:          "127.0.0.1",
			allowIPLiterals: false,
			expectedStatus:  http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate key pair
			privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatalf("Failed to generate key: %v", err)
			}

			cfg := &config.Config{
				Port:              8080,
				AllowedDomains:    []string{tt.domain},
				SignatureLifetime: 1 * time.Hour,
				PrivateKey:        privateKey,
				PublicKey:         &privateKey.PublicKey,
				CertDialTimeout:   10 * time.Second,
				AllowIPLiterals:   tt.allowIPLiterals,
				LogLevel:          "error",
			}

			fakeRetriever := cert.NewFakeRetriever()
			server := NewWithRetriever(cfg, fakeRetriever)

			req := httptest.NewRequest(http.MethodGet, "/v1/pins?domain="+tt.domain, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus >= 400 {
				var errorResp models.Error
				if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}

				if errorResp.Code != tt.expectedStatus {
					t.Errorf("Expected error code %d, got %d", tt.expectedStatus, errorResp.Code)
				}
			}
		})
	}
}
