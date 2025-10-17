package cert

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"sync"
	"time"
)

// CertRetriever is an interface for retrieving TLS certificates
// This allows for easy testing with fake implementations
type CertRetriever interface {
	GetCertificates(domain string) ([]*x509.Certificate, error)
}

// cacheEntry holds cached certificates with expiry
type cacheEntry struct {
	certs     []*x509.Certificate
	expiresAt time.Time
}

// Retriever retrieves TLS certificates for domains
type Retriever struct {
	dialTimeout time.Duration
	cacheTTL    time.Duration
	cache       map[string]*cacheEntry
	mu          sync.RWMutex
}

// NewRetriever creates a new certificate retriever
func NewRetriever(dialTimeout time.Duration, cacheTTL time.Duration) *Retriever {
	return &Retriever{
		dialTimeout: dialTimeout,
		cacheTTL:    cacheTTL,
		cache:       make(map[string]*cacheEntry),
	}
}

// GetCertificates retrieves the certificate chain for a domain
// Uses cache if TTL > 0 and entry is still valid
func (r *Retriever) GetCertificates(domain string) ([]*x509.Certificate, error) {
	// Check cache if TTL is enabled (> 0)
	if r.cacheTTL > 0 {
		r.mu.RLock()
		entry, found := r.cache[domain]
		r.mu.RUnlock()

		if found && time.Now().Before(entry.expiresAt) {
			// Cache hit - return cached certificates
			return entry.certs, nil
		}
	}

	// Cache miss or expired - retrieve certificates
	certs, err := r.fetchCertificates(domain)
	if err != nil {
		return nil, err
	}

	// Store in cache if TTL is enabled
	if r.cacheTTL > 0 {
		r.mu.Lock()
		r.cache[domain] = &cacheEntry{
			certs:     certs,
			expiresAt: time.Now().Add(r.cacheTTL),
		}
		r.mu.Unlock()
	}

	return certs, nil
}

// fetchCertificates retrieves certificates from the domain via TLS connection
func (r *Retriever) fetchCertificates(domain string) ([]*x509.Certificate, error) {
	// Connect to the domain over TLS
	dialer := &net.Dialer{
		Timeout: r.dialTimeout,
	}

	conn, err := tls.DialWithDialer(
		dialer,
		"tcp",
		domain+":443",
		&tls.Config{
			ServerName:         domain,
			InsecureSkipVerify: false, // We want to verify the cert chain
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", domain, err)
	}
	defer conn.Close()

	// Get the peer certificates
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found for domain: %s", domain)
	}

	return certs, nil
}
