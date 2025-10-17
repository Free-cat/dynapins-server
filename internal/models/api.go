package models

// PinEnvelope represents the response containing signed certificate pins
type PinEnvelope struct {
	Domain     string   `json:"domain"`
	Pins       []string `json:"pins"`
	Created    string   `json:"created"`
	Expires    string   `json:"expires"`
	TTLSeconds int      `json:"ttl_seconds"`
	KeyID      string   `json:"keyId"`
	Alg        string   `json:"alg"`
	Signature  string   `json:"signature"`
}

// Error represents an API error response
type Error struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}
