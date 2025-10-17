package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// CreateJWS creates a JWS token with the given parameters using ECDSA P-256 (ES256)
func CreateJWS(privateKey *ecdsa.PrivateKey, keyID string, domain string, pins []string, ttl time.Duration) (string, error) {
	// Create a new JWT token
	token := jwt.New()

	// Set required claims
	if err := token.Set("domain", domain); err != nil {
		return "", fmt.Errorf("failed to set domain claim: %w", err)
	}
	if err := token.Set("pins", pins); err != nil {
		return "", fmt.Errorf("failed to set pins claim: %w", err)
	}

	// Set standard JWT claims
	now := time.Now().UTC()
	if err := token.Set(jwt.IssuedAtKey, now.Unix()); err != nil {
		return "", fmt.Errorf("failed to set iat claim: %w", err)
	}
	if err := token.Set(jwt.ExpirationKey, now.Add(ttl).Unix()); err != nil {
		return "", fmt.Errorf("failed to set exp claim: %w", err)
	}
	if err := token.Set("ttl_seconds", int(ttl.Seconds())); err != nil {
		return "", fmt.Errorf("failed to set ttl_seconds claim: %w", err)
	}

	// Create JWS headers
	headers := jws.NewHeaders()
	if err := headers.Set(jws.AlgorithmKey, jwa.ES256); err != nil {
		return "", fmt.Errorf("failed to set algorithm header: %w", err)
	}
	if err := headers.Set(jws.KeyIDKey, keyID); err != nil {
		return "", fmt.Errorf("failed to set kid header: %w", err)
	}

	// Sign the token with ES256 (ECDSA P-256 + SHA-256)
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, privateKey, jws.WithProtectedHeaders(headers)))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return string(signed), nil
}
