package server

import (
	"crypto/x509"
	"encoding/json"
	"net/http"
	"time"

	"pinning-server/internal/crypto"
	"pinning-server/internal/logger"
	"pinning-server/internal/models"
)

// handleGetPins handles GET /v1/pins?domain=example.com
func (s *Server) handleGetPins(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Only allow GET requests
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", http.StatusMethodNotAllowed,
			"duration_ms", time.Since(start).Milliseconds())
		return
	}

	// Get domain from query parameter
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		writeError(w, "Missing required query parameter: domain", http.StatusBadRequest)
		logger.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", http.StatusBadRequest,
			"error", "missing_domain",
			"duration_ms", time.Since(start).Milliseconds())
		return
	}

	// Validate domain format (basic validation for malformed domains)
	if len(domain) == 0 || len(domain) > 253 {
		writeError(w, "Invalid domain parameter", http.StatusBadRequest)
		return
	}

	// Check if backup pins should be included
	includeBackupStr := r.URL.Query().Get("include-backup-pins")
	includeBackup := includeBackupStr == "true"

	logger.Info("Processing pins request", "domain", domain, "remote_addr", r.RemoteAddr)

	// Validate domain is in whitelist
	if !s.validator.IsAllowed(domain) {
		logger.Warn("Domain not in whitelist", "domain", domain)
		writeError(w, "Domain not found in whitelist", http.StatusForbidden)
		logger.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"domain", domain,
			"status", http.StatusForbidden,
			"error", "domain_not_allowed",
			"duration_ms", time.Since(start).Milliseconds())
		return
	}

	// Retrieve certificates for the domain
	certs, err := s.retriever.GetCertificates(domain)
	if err != nil {
		logger.Error("Failed to retrieve certificates", "domain", domain, "error", err)
		writeError(w, "Failed to retrieve certificate for domain", http.StatusUnprocessableEntity)
		logger.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"domain", domain,
			"status", http.StatusUnprocessableEntity,
			"error", "cert_retrieval_failed",
			"duration_ms", time.Since(start).Milliseconds())
		return
	}

	// Determine which certificates to use for pin generation
	var certsForPinning []*x509.Certificate
	if includeBackup && len(certs) > 1 {
		// Use leaf and intermediate certificate
		certsForPinning = certs[:2]
	} else if len(certs) > 0 {
		// Use only leaf certificate
		certsForPinning = certs[:1]
	}

	// Generate SPKI hashes in TrustKit format: base64(SHA256(SPKI))
	pins := crypto.GenerateSPKIHashes(certsForPinning)

	// Create JWS token
	jwsToken, err := crypto.CreateJWS(
		s.config.PrivateKey,
		s.keyID,
		domain,
		pins,
		s.config.SignatureLifetime,
	)
	if err != nil {
		logger.Error("Failed to create JWS token", "domain", domain, "error", err)
		writeError(w, "Failed to generate signed token", http.StatusInternalServerError)
		logger.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"domain", domain,
			"status", http.StatusInternalServerError,
			"error", "jws_creation_failed",
			"duration_ms", time.Since(start).Milliseconds())
		return
	}

	// Create JWS response
	response := map[string]string{
		"jws": jwsToken,
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", "error", err)
	}

	logger.Info("Request completed",
		"method", r.Method,
		"path", r.URL.Path,
		"domain", domain,
		"status", http.StatusOK,
		"pin_count", len(pins),
		"include_backup", includeBackup,
		"duration_ms", time.Since(start).Milliseconds())
}

// handleHealth handles GET /health - basic liveness check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	}); err != nil {
		logger.Error("Failed to encode health response", "error", err)
	}
}

// handleReadiness handles GET /readiness - readiness check with crypto validation
func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Verify crypto components are initialized
	if s.config.PrivateKey == nil || s.config.PublicKey == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "not ready",
			"reason": "crypto keys not initialized",
		}); err != nil {
			logger.Error("Failed to encode readiness error response", "error", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "ready",
		"allowed_domains": len(s.config.AllowedDomains),
		"key_id":          s.keyID,
	}); err != nil {
		logger.Error("Failed to encode readiness response", "error", err)
	}
}

// writeError writes an error response
func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(models.Error{
		Error: message,
		Code:  code,
	}); err != nil {
		logger.Error("Failed to encode error response", "error", err)
	}
}
