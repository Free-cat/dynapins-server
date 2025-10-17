package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"
)

// MockTLSServer creates a local TLS server for testing
type MockTLSServer struct {
	listener net.Listener
	cert     *x509.Certificate
	address  string
}

// TestingTB is a subset of testing.TB interface
type TestingTB interface {
	Helper()
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// NewMockTLSServer creates a new mock TLS server
func NewMockTLSServer(t TestingTB) *MockTLSServer {
	t.Helper()

	// Generate RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "127.0.0.1"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		t.Fatal(err)
	}

	// Create TLS config
	tlsCert := tls.Certificate{
		Certificate: [][]byte{certBytes},
		PrivateKey:  privateKey,
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	}

	// Start listener
	listener, err := tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	if err != nil {
		t.Fatal(err)
	}

	server := &MockTLSServer{
		listener: listener,
		cert:     cert,
		address:  listener.Addr().String(),
	}

	// Start accepting connections
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	return server
}

// Close shuts down the mock server
func (m *MockTLSServer) Close() {
	if m.listener != nil {
		m.listener.Close()
	}
}

// Address returns the server address (host:port)
func (m *MockTLSServer) Address() string {
	return m.address
}

// Host returns just the hostname
func (m *MockTLSServer) Host() string {
	host, _, _ := net.SplitHostPort(m.address)
	return host
}

// Certificate returns the server's certificate
func (m *MockTLSServer) Certificate() *x509.Certificate {
	return m.cert
}
