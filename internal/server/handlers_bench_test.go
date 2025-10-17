package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pinning-server/internal/config"
)

// BenchmarkHandleGetPins benchmarks the /v1/pins endpoint
func BenchmarkHandleGetPins(b *testing.B) {
	// For benchmark, we'll skip the actual cert retrieval
	// and focus on the handler logic performance
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	// Use a domain that will be allowed but we won't actually fetch certs
	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"benchmark.local"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)

	// Note: This will fail cert retrieval but measures handler overhead
	req := httptest.NewRequest(http.MethodGet, "/v1/pins?domain=benchmark.local", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		srv.handleGetPins(w, req)
		// We expect 422 since we can't reach benchmark.local, but that's OK
		// We're measuring the handler logic, validation, etc.
	}
}

// BenchmarkHandleGetPinsValidationOnly benchmarks just the validation path
func BenchmarkHandleGetPinsValidationOnly(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"allowed.com"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)

	benchmarks := []struct {
		name     string
		domain   string
		expected int
	}{
		{"allowed_domain", "allowed.com", http.StatusUnprocessableEntity}, // Will fail on cert fetch
		{"forbidden_domain", "forbidden.com", http.StatusForbidden},
		{"missing_domain", "", http.StatusBadRequest},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			var req *http.Request
			if bm.domain == "" {
				req = httptest.NewRequest(http.MethodGet, "/v1/pins", nil)
			} else {
				req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/pins?domain=%s", bm.domain), nil)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				srv.handleGetPins(w, req)
			}
		})
	}
}

// BenchmarkHandleGetPinsParallel benchmarks parallel requests
func BenchmarkHandleGetPinsParallel(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"benchmark.local"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/v1/pins?domain=benchmark.local", nil)
		for pb.Next() {
			w := httptest.NewRecorder()
			srv.handleGetPins(w, req)
		}
	})
}

// BenchmarkHandleHealth benchmarks health check endpoint
func BenchmarkHandleHealth(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"example.com"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		srv.handleHealth(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkHandleHealthParallel benchmarks parallel health checks
func BenchmarkHandleHealthParallel(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"example.com"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		for pb.Next() {
			w := httptest.NewRecorder()
			srv.handleHealth(w, req)

			if w.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkHandleReadiness benchmarks readiness check endpoint
func BenchmarkHandleReadiness(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"example.com"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)
	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		srv.handleReadiness(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkHandleReadinessParallel benchmarks parallel readiness checks
func BenchmarkHandleReadinessParallel(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"example.com"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
		for pb.Next() {
			w := httptest.NewRecorder()
			srv.handleReadiness(w, req)

			if w.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkServeHTTP benchmarks the full HTTP handler pipeline
func BenchmarkServeHTTP(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"benchmark.local"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)
	req := httptest.NewRequest(http.MethodGet, "/v1/pins?domain=benchmark.local", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
	}
}

// BenchmarkServeHTTPHealth benchmarks ServeHTTP with health endpoint
func BenchmarkServeHTTPHealth(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"example.com"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// BenchmarkErrorHandling benchmarks error path performance
func BenchmarkErrorHandling(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"example.com"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)

	benchmarks := []struct {
		name     string
		request  *http.Request
		expected int
	}{
		{
			"missing_domain",
			httptest.NewRequest(http.MethodGet, "/v1/pins", nil),
			http.StatusBadRequest,
		},
		{
			"forbidden_domain",
			httptest.NewRequest(http.MethodGet, "/v1/pins?domain=forbidden.com", nil),
			http.StatusForbidden,
		},
		{
			"method_not_allowed",
			httptest.NewRequest(http.MethodPost, "/v1/pins?domain=example.com", nil),
			http.StatusMethodNotAllowed,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				srv.handleGetPins(w, bm.request)

				if w.Code != bm.expected {
					b.Fatalf("Expected status %d, got %d", bm.expected, w.Code)
				}
			}
		})
	}
}

// BenchmarkRoutingOverhead benchmarks the routing overhead
func BenchmarkRoutingOverhead(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	cfg := &config.Config{
		Port:              8080,
		AllowedDomains:    []string{"example.com"},
		SignatureLifetime: time.Hour,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		CertDialTimeout:   10 * time.Second,
		LogLevel:          "error",
	}

	srv := New(cfg)

	endpoints := []string{
		"/health",
		"/readiness",
		"/v1/pins?domain=example.com",
	}

	for _, endpoint := range endpoints {
		b.Run(endpoint, func(b *testing.B) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				srv.ServeHTTP(w, req)
			}
		})
	}
}
