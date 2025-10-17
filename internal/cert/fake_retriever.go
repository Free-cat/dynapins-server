package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

// FakeRetriever is a test double for CertRetriever
type FakeRetriever struct {
	certs map[string][]*x509.Certificate
	err   error
}

// NewFakeRetriever creates a fake retriever with default test certificates
func NewFakeRetriever() *FakeRetriever {
	return &FakeRetriever{
		certs: make(map[string][]*x509.Certificate),
	}
}

// SetCertificates sets the certificates to return for a domain
func (f *FakeRetriever) SetCertificates(domain string, certs []*x509.Certificate) {
	f.certs[domain] = certs
}

// SetError sets an error to return for all GetCertificates calls
func (f *FakeRetriever) SetError(err error) {
	f.err = err
}

// GetCertificates implements CertRetriever interface
func (f *FakeRetriever) GetCertificates(domain string) ([]*x509.Certificate, error) {
	if f.err != nil {
		return nil, f.err
	}

	certs, ok := f.certs[domain]
	if !ok {
		return nil, fmt.Errorf("no certificates configured for domain: %s", domain)
	}

	return certs, nil
}

// GenerateTestCertificate creates a self-signed certificate for testing
func GenerateTestCertificate(commonName string) (*x509.Certificate, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			CommonName:   commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{commonName},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// GenerateTestCertificateChain creates a chain with leaf and intermediate certs
func GenerateTestCertificateChain(commonName string) ([]*x509.Certificate, error) {
	// Generate intermediate cert
	intermediateCert, err := GenerateTestCertificate("Intermediate CA")
	if err != nil {
		return nil, err
	}

	// Generate leaf cert
	leafCert, err := GenerateTestCertificate(commonName)
	if err != nil {
		return nil, err
	}

	return []*x509.Certificate{leafCert, intermediateCert}, nil
}
