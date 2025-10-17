package crypto

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
)

// GenerateSPKIHash generates a SHA-256 hash of a certificate's Subject Public Key Info
func GenerateSPKIHash(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return hex.EncodeToString(hash[:])
}

// GenerateSPKIHashes generates SHA-256 hashes for all certificates in a chain
func GenerateSPKIHashes(certs []*x509.Certificate) []string {
	hashes := make([]string, 0, len(certs))
	for _, cert := range certs {
		hashes = append(hashes, GenerateSPKIHash(cert))
	}
	return hashes
}
