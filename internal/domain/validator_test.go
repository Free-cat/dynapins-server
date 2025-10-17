package domain

import (
	"testing"
)

func TestValidator_IsAllowed(t *testing.T) {
	tests := []struct {
		name           string
		allowedDomains []string
		testDomain     string
		expected       bool
	}{
		{
			name:           "exact match",
			allowedDomains: []string{"example.com", "api.example.com"},
			testDomain:     "example.com",
			expected:       true,
		},
		{
			name:           "exact match - case insensitive",
			allowedDomains: []string{"example.com"},
			testDomain:     "EXAMPLE.COM",
			expected:       true,
		},
		{
			name:           "not in whitelist",
			allowedDomains: []string{"example.com"},
			testDomain:     "notallowed.com",
			expected:       false,
		},
		{
			name:           "wildcard match - single level",
			allowedDomains: []string{"*.example.com"},
			testDomain:     "api.example.com",
			expected:       true,
		},
		{
			name:           "wildcard match - multiple subdomains in whitelist",
			allowedDomains: []string{"*.example.com"},
			testDomain:     "www.example.com",
			expected:       true,
		},
		{
			name:           "wildcard no match - too many levels",
			allowedDomains: []string{"*.example.com"},
			testDomain:     "api.v2.example.com",
			expected:       false,
		},
		{
			name:           "wildcard no match - different domain",
			allowedDomains: []string{"*.example.com"},
			testDomain:     "example.org",
			expected:       false,
		},
		{
			name:           "wildcard no match - base domain",
			allowedDomains: []string{"*.example.com"},
			testDomain:     "example.com",
			expected:       false,
		},
		{
			name:           "mixed exact and wildcard",
			allowedDomains: []string{"example.com", "*.api.example.com"},
			testDomain:     "v1.api.example.com",
			expected:       true,
		},
		{
			name:           "whitespace handling",
			allowedDomains: []string{" example.com ", " *.api.example.com "},
			testDomain:     " example.com ",
			expected:       true,
		},
		{
			name:           "empty domain",
			allowedDomains: []string{"example.com"},
			testDomain:     "",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tt.allowedDomains)
			result := validator.IsAllowed(tt.testDomain)
			if result != tt.expected {
				t.Errorf("IsAllowed(%q) = %v, want %v", tt.testDomain, result, tt.expected)
			}
		})
	}
}

func TestValidator_IsAllowed_MultipleWildcards(t *testing.T) {
	validator := NewValidator([]string{"*.example.com", "*.test.org", "exact.domain.net"})

	tests := []struct {
		domain   string
		expected bool
	}{
		{"api.example.com", true},
		{"www.example.com", true},
		{"api.test.org", true},
		{"exact.domain.net", true},
		{"sub.api.example.com", false}, // Too many levels
		{"example.com", false},         // Base domain not allowed
		{"notallowed.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			result := validator.IsAllowed(tt.domain)
			if result != tt.expected {
				t.Errorf("IsAllowed(%q) = %v, want %v", tt.domain, result, tt.expected)
			}
		})
	}
}
