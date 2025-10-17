package domain

import (
	"testing"
)

// BenchmarkIsAllowed benchmarks domain validation
func BenchmarkIsAllowed(b *testing.B) {
	allowedDomains := []string{
		"example.com",
		"*.api.example.com",
		"google.com",
		"*.google.com",
		"github.com",
		"*.github.com",
		"cloudflare.com",
		"*.cloudflare.com",
	}

	validator := NewValidator(allowedDomains)

	benchmarks := []struct {
		name   string
		domain string
	}{
		{"exact_match", "example.com"},
		{"wildcard_match", "v1.api.example.com"},
		{"no_match", "notallowed.com"},
		{"subdomain_no_match", "sub.example.com"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				validator.IsAllowed(bm.domain)
			}
		})
	}
}

// BenchmarkIsAllowedParallel benchmarks parallel domain validation
func BenchmarkIsAllowedParallel(b *testing.B) {
	allowedDomains := []string{
		"example.com",
		"*.api.example.com",
		"google.com",
		"*.google.com",
	}

	validator := NewValidator(allowedDomains)

	b.RunParallel(func(pb *testing.PB) {
		domains := []string{"example.com", "v1.api.example.com", "notallowed.com", "google.com"}
		i := 0
		for pb.Next() {
			validator.IsAllowed(domains[i%len(domains)])
			i++
		}
	})
}

// BenchmarkIsAllowedLargeDomainList benchmarks with large domain list
func BenchmarkIsAllowedLargeDomainList(b *testing.B) {
	// Create a large list of allowed domains
	allowedDomains := make([]string, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			allowedDomains[i] = "example" + string(rune(i)) + ".com"
		} else {
			allowedDomains[i] = "*.example" + string(rune(i)) + ".com"
		}
	}

	validator := NewValidator(allowedDomains)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.IsAllowed("example50.com")
	}
}
