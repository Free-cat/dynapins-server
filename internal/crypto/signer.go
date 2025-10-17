package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// SignPayload signs a payload with an Ed25519 private key
func SignPayload(payload interface{}, privateKey ed25519.PrivateKey) (string, error) {
	// Marshal the payload to JSON (canonical form for signing)
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Sign the data
	signature := ed25519.Sign(privateKey, data)

	// Return base64-encoded signature
	return base64.StdEncoding.EncodeToString(signature), nil
}

// SignablePayload represents the data that will be signed
type SignablePayload struct {
	Domain     string   `json:"domain"`
	Pins       []string `json:"pins"`
	Created    string   `json:"created"`
	Expires    string   `json:"expires"`
	TTLSeconds int      `json:"ttl_seconds"`
	KeyID      string   `json:"keyId"`
	Alg        string   `json:"alg"`
}
