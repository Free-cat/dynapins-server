package crypto

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
)

// GenerateSPKIHash generates a base64-encoded SHA-256 hash of a certificate's SPKI
// This matches TrustKit's pin format: base64(SHA256(SPKI))
// SPKI (SubjectPublicKeyInfo) includes both the algorithm identifier and the public key
func GenerateSPKIHash(cert *x509.Certificate) string {
	// Use the certificate's RawSubjectPublicKeyInfo (SPKI in DER format)
	// This works for all key types (EC, RSA, etc.)
	hash := sha256.Sum256(cert.RawSubjectPublicKeyInfo)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// GenerateSPKIHashes generates base64-encoded SHA-256 hashes for all certificates in a chain
// This matches TrustKit's pin format
func GenerateSPKIHashes(certs []*x509.Certificate) []string {
	hashes := make([]string, 0, len(certs))
	for _, cert := range certs {
		hashes = append(hashes, GenerateSPKIHash(cert))
	}
	return hashes
}
