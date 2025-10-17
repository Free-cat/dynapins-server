package server

import (
	"encoding/json"
	"net/http"
	"time"

	"pinning-server/internal/crypto"
	"pinning-server/internal/logger"
	"pinning-server/internal/models"
)

// handleGetPins handles GET /v1/pins?domain=example.com
func (s *Server) handleGetPins(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get domain from query parameter
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		writeError(w, "Missing required query parameter: domain", http.StatusBadRequest)
		return
	}

	logger.Info("Processing pins request", "domain", domain, "remote_addr", r.RemoteAddr)

	// Validate domain is in whitelist
	if !s.validator.IsAllowed(domain) {
		logger.Warn("Domain not in whitelist", "domain", domain)
		writeError(w, "Domain not found in whitelist", http.StatusForbidden)
		return
	}

	// Retrieve certificates for the domain
	certs, err := s.retriever.GetCertificates(domain)
	if err != nil {
		logger.Error("Failed to retrieve certificates", "domain", domain, "error", err)
		writeError(w, "Failed to retrieve certificate for domain", http.StatusUnprocessableEntity)
		return
	}

	// Generate SPKI hashes
	pins := crypto.GenerateSPKIHashes(certs)

	// Generate timestamps
	now := time.Now().UTC()
	expires := now.Add(s.config.SignatureLifetime)

	// Create signable payload
	signablePayload := crypto.SignablePayload{
		Domain:     domain,
		Pins:       pins,
		Created:    now.Format(time.RFC3339),
		Expires:    expires.Format(time.RFC3339),
		TTLSeconds: int(s.config.SignatureLifetime.Seconds()),
		KeyID:      s.keyID,
		Alg:        "Ed25519",
	}

	// Sign the payload
	signature, err := crypto.SignPayload(signablePayload, s.config.PrivateKey)
	if err != nil {
		logger.Error("Failed to sign payload", "domain", domain, "error", err)
		writeError(w, "Failed to generate signature", http.StatusInternalServerError)
		return
	}

	// Create response envelope
	envelope := models.PinEnvelope{
		Domain:     signablePayload.Domain,
		Pins:       signablePayload.Pins,
		Created:    signablePayload.Created,
		Expires:    signablePayload.Expires,
		TTLSeconds: signablePayload.TTLSeconds,
		KeyID:      signablePayload.KeyID,
		Alg:        signablePayload.Alg,
		Signature:  signature,
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		logger.Error("Failed to encode response", "error", err)
	}

	logger.Info("Successfully processed pins request",
		"domain", domain,
		"pin_count", len(pins),
		"expires", expires.Format(time.RFC3339))
}

// handleHealth handles GET /health - basic liveness check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
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
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not ready",
			"reason": "crypto keys not initialized",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "ready",
		"allowed_domains": len(s.config.AllowedDomains),
		"key_id":         s.keyID,
	})
}

// writeError writes an error response
func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.Error{
		Error: message,
		Code:  code,
	})
}
