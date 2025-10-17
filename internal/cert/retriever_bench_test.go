package cert

import (
	"testing"
	"time"
)

// Note: Benchmarks for GetCertificates with real TLS connections are intentionally
// omitted because they would require external network access or complex mocking.
// These benchmarks would also be unstable due to network conditions.
//
// Instead, we focus on benchmarking the retriever creation and other isolated operations.

// BenchmarkNewRetriever benchmarks retriever creation
func BenchmarkNewRetriever(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewRetriever(10*time.Second, 5*time.Minute)
	}
}

// BenchmarkNewRetrieverWithDifferentTimeouts benchmarks creation with various timeouts
func BenchmarkNewRetrieverWithDifferentTimeouts(b *testing.B) {
	timeouts := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		10 * time.Second,
		30 * time.Second,
		60 * time.Second,
	}

	for _, timeout := range timeouts {
		b.Run(timeout.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				NewRetriever(timeout, 5*time.Minute)
			}
		})
	}
}

// BenchmarkNewRetrieverParallel benchmarks parallel retriever creation
func BenchmarkNewRetrieverParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewRetriever(10*time.Second, 5*time.Minute)
		}
	})
}
