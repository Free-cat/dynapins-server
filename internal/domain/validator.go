package domain

import (
	"net"
	"strings"
)

// Validator validates domain names against a whitelist
type Validator struct {
	allowedDomains  []string
	allowIPLiterals bool
}

// NewValidator creates a new domain validator
func NewValidator(allowedDomains []string) *Validator {
	return &Validator{
		allowedDomains:  allowedDomains,
		allowIPLiterals: false,
	}
}

// NewValidatorWithOptions creates a validator with custom options
func NewValidatorWithOptions(allowedDomains []string, allowIPLiterals bool) *Validator {
	return &Validator{
		allowedDomains:  allowedDomains,
		allowIPLiterals: allowIPLiterals,
	}
}

// IsAllowed checks if a domain is in the whitelist
// Supports wildcards like "*.example.com"
// Rejects IP literals unless allowIPLiterals is true
func (v *Validator) IsAllowed(domain string) bool {
	domain = strings.ToLower(strings.TrimSpace(domain))

	// Reject IP literals (IPv4 and IPv6) unless explicitly allowed
	if !v.allowIPLiterals {
		if net.ParseIP(domain) != nil {
			return false
		}
		// Also check for [IPv6] format
		if strings.HasPrefix(domain, "[") && strings.HasSuffix(domain, "]") {
			ip := domain[1 : len(domain)-1]
			if net.ParseIP(ip) != nil {
				return false
			}
		}
	}

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
