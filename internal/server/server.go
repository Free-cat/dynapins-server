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
	retriever cert.CertRetriever
	keyID     string
	mux       *http.ServeMux
}

// New creates a new HTTP server
func New(cfg *config.Config) *Server {
	return NewWithRetriever(cfg, cert.NewRetriever(cfg.CertDialTimeout, cfg.CertCacheTTL))
}

// NewWithRetriever creates a new HTTP server with a custom certificate retriever
// This is useful for testing with fake retrievers
func NewWithRetriever(cfg *config.Config, retriever cert.CertRetriever) *Server {
	s := &Server{
		config:    cfg,
		validator: domain.NewValidatorWithOptions(cfg.AllowedDomains, cfg.AllowIPLiterals),
		retriever: retriever,
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
