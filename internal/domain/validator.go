package domain

import (
	"strings"
)

// Validator validates domain names against a whitelist
type Validator struct {
	allowedDomains []string
}

// NewValidator creates a new domain validator
func NewValidator(allowedDomains []string) *Validator {
	return &Validator{
		allowedDomains: allowedDomains,
	}
}

// IsAllowed checks if a domain is in the whitelist
// Supports wildcards like "*.example.com"
func (v *Validator) IsAllowed(domain string) bool {
	domain = strings.ToLower(strings.TrimSpace(domain))

	for _, allowed := range v.allowedDomains {
		allowed = strings.ToLower(strings.TrimSpace(allowed))

		// Exact match
		if domain == allowed {
			return true
		}

		// Wildcard match (only single-level wildcard supported)
		if strings.HasPrefix(allowed, "*.") {
			suffix := allowed[2:] // Remove "*."
			// Check if domain ends with the suffix and has exactly one more level
			if strings.HasSuffix(domain, suffix) {
				// Ensure there's a dot before the suffix
				if len(domain) > len(suffix) && domain[len(domain)-len(suffix)-1] == '.' {
					// Ensure there's only one additional level (no extra dots)
					prefix := domain[:len(domain)-len(suffix)-1]
					if !strings.Contains(prefix, ".") {
						return true
					}
				}
			}
		}
	}

	return false
}
