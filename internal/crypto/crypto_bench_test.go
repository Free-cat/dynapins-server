package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

// BenchmarkGenerateSPKIHashes benchmarks SPKI hash generation
func BenchmarkGenerateSPKIHashes(b *testing.B) {
	// Generate test certificates
	certs := generateTestCerts(b, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateSPKIHashes(certs)
	}
}

// BenchmarkGenerateSPKIHashesSingleCert benchmarks SPKI hash generation for single cert
func BenchmarkGenerateSPKIHashesSingleCert(b *testing.B) {
	certs := generateTestCerts(b, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateSPKIHashes(certs)
	}
}

// BenchmarkGenerateSPKIHashesMultipleCerts benchmarks SPKI hash generation for chain
func BenchmarkGenerateSPKIHashesMultipleCerts(b *testing.B) {
	certs := generateTestCerts(b, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateSPKIHashes(certs)
	}
}

// BenchmarkCreateJWS benchmarks JWS token creation with ECDSA P-256
func BenchmarkCreateJWS(b *testing.B) {
	// Generate ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}

	keyID := "test-key-id"
	domain := "example.com"
	pins := []string{
		"r/mIkG3eEpVdm+u/ko/cwxzOMo1bk4TyHIlByibiA5E=",
		"YLh1dUR9y6Kja30RrAn7JKnbQG/uEtLMkBgFF2Fuihg=",
	}
	lifetime := time.Hour

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CreateJWS(privateKey, keyID, domain, pins, lifetime)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCreateJWSWithDifferentPinCounts benchmarks JWS creation with varying pin counts
func BenchmarkCreateJWSWithDifferentPinCounts(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}

	keyID := "test-key-id"
	domain := "example.com"
	lifetime := time.Hour

	benchmarks := []struct {
		name     string
		pinCount int
	}{
		{"1pin", 1},
		{"2pins", 2},
		{"5pins", 5},
		{"10pins", 10},
	}

	for _, bm := range benchmarks {
		pins := make([]string, bm.pinCount)
		for i := 0; i < bm.pinCount; i++ {
			pins[i] = "r/mIkG3eEpVdm+u/ko/cwxzOMo1bk4TyHIlByibiA5E="
		}

		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := CreateJWS(privateKey, keyID, domain, pins, lifetime)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkGenerateKeyID benchmarks key ID generation
func BenchmarkGenerateKeyID(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateKeyID(publicKey)
	}
}

// BenchmarkParallelJWSCreation benchmarks parallel JWS creation
func BenchmarkParallelJWSCreation(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}

	keyID := "test-key-id"
	domain := "example.com"
	pins := []string{
		"r/mIkG3eEpVdm+u/ko/cwxzOMo1bk4TyHIlByibiA5E=",
		"YLh1dUR9y6Kja30RrAn7JKnbQG/uEtLMkBgFF2Fuihg=",
	}
	lifetime := time.Hour

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := CreateJWS(privateKey, keyID, domain, pins, lifetime)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkGetPublicKeyFromPrivate benchmarks public key extraction
func BenchmarkGetPublicKeyFromPrivate(b *testing.B) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetPublicKeyFromPrivate(privateKey)
	}
}

// BenchmarkKeyGeneration benchmarks ECDSA P-256 key generation
func BenchmarkKeyGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSPKIHashWithDifferentAlgorithms benchmarks hash generation with different key types
func BenchmarkSPKIHashWithDifferentAlgorithms(b *testing.B) {
	// RSA 2048
	b.Run("RSA2048", func(b *testing.B) {
		certs := generateTestCertsRSA(b, 1, 2048)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GenerateSPKIHashes(certs)
		}
	})

	// RSA 4096
	b.Run("RSA4096", func(b *testing.B) {
		certs := generateTestCertsRSA(b, 1, 4096)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GenerateSPKIHashes(certs)
		}
	})

	// ECDSA P-256
	b.Run("ECDSAP256", func(b *testing.B) {
		certs := generateTestCertsECDSA(b, 1, elliptic.P256())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GenerateSPKIHashes(certs)
		}
	})

	// ECDSA P-384
	b.Run("ECDSAP384", func(b *testing.B) {
		certs := generateTestCertsECDSA(b, 1, elliptic.P384())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GenerateSPKIHashes(certs)
		}
	})
}

// Helper function to generate test certificates with RSA
func generateTestCerts(b *testing.B, count int) []*x509.Certificate {
	b.Helper()
	return generateTestCertsRSA(b, count, 2048)
}

func generateTestCertsRSA(b *testing.B, count int, bits int) []*x509.Certificate {
	b.Helper()

	certs := make([]*x509.Certificate, count)
	for i := 0; i < count; i++ {
		// Generate RSA key
		privateKey, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			b.Fatal(err)
		}

		// Create certificate template
		template := x509.Certificate{
			SerialNumber: big.NewInt(int64(i + 1)),
			Subject: pkix.Name{
				Organization: []string{"Test Org"},
				CommonName:   "test.example.com",
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(365 * 24 * time.Hour),
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		// Create certificate
		certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			b.Fatal(err)
		}

		cert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			b.Fatal(err)
		}

		certs[i] = cert
	}

	return certs
}

func generateTestCertsECDSA(b *testing.B, count int, curve elliptic.Curve) []*x509.Certificate {
	b.Helper()

	certs := make([]*x509.Certificate, count)
	for i := 0; i < count; i++ {
		// Generate ECDSA key
		privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			b.Fatal(err)
		}

		// Create certificate template
		template := x509.Certificate{
			SerialNumber: big.NewInt(int64(i + 1)),
			Subject: pkix.Name{
				Organization: []string{"Test Org"},
				CommonName:   "test.example.com",
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(365 * 24 * time.Hour),
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		// Create certificate
		certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			b.Fatal(err)
		}

		cert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			b.Fatal(err)
		}

		certs[i] = cert
	}

	return certs
}
