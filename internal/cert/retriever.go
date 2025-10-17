package cert

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"
)

// Retriever retrieves TLS certificates for domains
type Retriever struct {
	dialTimeout time.Duration
}

// NewRetriever creates a new certificate retriever
func NewRetriever(dialTimeout time.Duration) *Retriever {
	return &Retriever{
		dialTimeout: dialTimeout,
	}
}

// GetCertificates retrieves the certificate chain for a domain
func (r *Retriever) GetCertificates(domain string) ([]*x509.Certificate, error) {
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
