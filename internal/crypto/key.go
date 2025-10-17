package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateKeyID generates a unique identifier for a public key
// by hashing the public key bytes and taking the first 8 characters
func GenerateKeyID(publicKey ed25519.PublicKey) string {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:])[:8]
}
