package server

import (
	"net/http"

	"pinning-server/internal/cert"
	"pinning-server/internal/config"
	"pinning-server/internal/crypto"
	"pinning-server/internal/domain"
)

// Server represents the HTTP server
type Server struct {
	config    *config.Config
	validator *domain.Validator
	retriever *cert.Retriever
	keyID     string
	mux       *http.ServeMux
}

// New creates a new HTTP server
func New(cfg *config.Config) *Server {
	s := &Server{
		config:    cfg,
		validator: domain.NewValidator(cfg.AllowedDomains),
		retriever: cert.NewRetriever(cfg.CertDialTimeout),
		keyID:     crypto.GenerateKeyID(cfg.PublicKey),
		mux:       http.NewServeMux(),
	}

	// Register routes
	s.mux.HandleFunc("/v1/pins", s.handleGetPins)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/readiness", s.handleReadiness)

	return s
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
