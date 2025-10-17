package crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
)

// GenerateKeyID generates a unique identifier for a public key
// by hashing the public key bytes and taking the first 8 characters
func GenerateKeyID(publicKey *ecdsa.PublicKey) string {
	// Marshal the public key to DER format (SPKI)
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		// This should never fail for a valid ECDSA public key
		// Return empty string to indicate error
		return ""
	}
	hash := sha256.Sum256(pubKeyBytes)
	return hex.EncodeToString(hash[:])[:8]
}

// GetPublicKeyFromPrivate extracts the public key from an ECDSA private key
func GetPublicKeyFromPrivate(privateKey *ecdsa.PrivateKey) *ecdsa.PublicKey {
	return &privateKey.PublicKey
}
