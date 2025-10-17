package models

// Error represents an API error response
type Error struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}
