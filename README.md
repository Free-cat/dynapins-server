# Dynapins Server

[![CI](https://github.com/Free-cat/dynapins-server/actions/workflows/ci.yml/badge.svg)](https://github.com/Free-cat/dynapins-server/actions/workflows/ci.yml)
[![Docker Pulls](https://img.shields.io/docker/pulls/freecats/dynapins-server)](https://hub.docker.com/r/freecats/dynapins-server)
[![Go Report Card](https://goreportcard.com/badge/github.com/Free-cat/dynapins-server)](https://goreportcard.com/report/github.com/Free-cat/dynapins-server)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go-based HTTP server that provides signed TLS certificate pins for dynamic SSL pinning. This server retrieves TLS certificates for whitelisted domains, generates SHA-256 hashes of their Subject Public Key Info (SPKI), and returns them in a JWS (JSON Web Signature) signed response.

## Features

- **Dynamic SSL Pinning**: Get certificate pins for domains without hardcoding them
- **Signature-Verified Trust**: All responses are signed with ECDSA P-256 (ES256) for verification
- **Certificate Caching**: Optional TTL-based caching to reduce TLS handshakes and improve performance
- **Domain Whitelist**: Only serves pins for explicitly allowed domains (supports wildcards)
- **Stateless**: No database required, fully stateless operation
- **High Performance**: Built in Go with minimal dependencies
- **Security Hardened**: Runs as non-root user, validates domains, configurable timeouts

## Prerequisites

- **Go 1.25+** (for local development)
- **Docker** (for containerized deployment)

## Configuration

The server is configured entirely via environment variables for maximum flexibility:

### Configuration Reference

| Variable | Description | Required | Default | Example |
|----------|-------------|----------|---------|---------|
| **Server Settings** |
| `PORT` | The port the server listens on | No | `8080` | `8080`, `3000` |
| `READ_TIMEOUT` | Maximum duration for reading the entire request | No | `10s` | `10s`, `30s`, `1m` |
| `WRITE_TIMEOUT` | Maximum duration before timing out writes of the response | No | `10s` | `10s`, `30s` |
| `READ_HEADER_TIMEOUT` | Maximum duration for reading request headers (Slowloris protection) | No | `5s` | `5s`, `10s` |
| `IDLE_TIMEOUT` | Maximum time to wait for the next request when keep-alives are enabled | No | `60s` | `60s`, `2m` |
| `SHUTDOWN_TIMEOUT` | Maximum time to wait for graceful server shutdown | No | `10s` | `10s`, `30s` |
| `MAX_HEADER_BYTES` | Maximum size of request headers in bytes | No | `1048576` (1MB) | `1048576`, `524288` |
| **Domain & Security** |
| `ALLOWED_DOMAINS` | Comma-separated list of domains and wildcards to allow | **Yes** | - | `"example.com,*.example.com,api.anotherexample.com"` |
| `SIGNATURE_LIFETIME` | The validity period of the generated JWS signature | No | `1h` | `1h`, `30m`, `2h30m` |
| `PRIVATE_KEY_PEM` | The PEM-encoded ECDSA P-256 private key for signing | **Yes** | - | `"-----BEGIN PRIVATE KEY-----..."` |
| `ALLOW_IP_LITERALS` | Allow IP addresses as domains (for development only) | No | `false` | `true`, `false` |
| **Certificate Retrieval & Caching** |
| `CERT_DIAL_TIMEOUT` | Maximum time to wait when connecting to retrieve certificates | No | `10s` | `10s`, `15s`, `30s` |
| `CERT_CACHE_TTL` | Certificate cache TTL (0 to disable caching) | No | `5m` | `5m`, `10m`, `0` (disabled) |
| **Logging** |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | No | `info` | `info`, `debug`, `error` |

### Duration Format

Duration values support Go's duration format:
- `s` = seconds (e.g., `30s`)
- `m` = minutes (e.g., `5m`)
- `h` = hours (e.g., `2h`)
- Combined: `1h30m`, `2h30m45s`

### Generating an ECDSA P-256 Key Pair

To generate a new ECDSA P-256 key pair for signing:

```bash
# Generate private key (ECDSA P-256)
openssl ecparam -genkey -name prime256v1 -noout -out private_key.pem

# Extract public key
openssl ec -in private_key.pem -pubout -out public_key.pem

# Optional: Convert to PKCS#8 format
openssl pkcs8 -topk8 -nocrypt -in private_key.pem -out private_key_pkcs8.pem
```

## Running the Server

### Option 1: Pre-built Docker Image

Pull and run the pre-built image:

```bash
docker pull freecats/dynapins-server:latest

docker run -p 8080:8080 \
  -e ALLOWED_DOMAINS="example.com,*.example.com" \
  -e PRIVATE_KEY_PEM="$(cat private_key.pem)" \
  freecats/dynapins-server:latest
```

### Option 2: Build from Source

1. **Build the Docker image:**

   ```bash
   docker build -t dynapins-server .
   # or use make
   make docker-build
   ```

2. **Run the container:**

   Replace the example values with your actual configuration.

   ```bash
   docker run -p 8080:8080 \
     -e PORT=8080 \
     -e ALLOWED_DOMAINS="example.com,*.example.com" \
     -e SIGNATURE_LIFETIME="1h" \
     -e PRIVATE_KEY_PEM="$(cat private_key.pem)" \
     dynapins-server
   ```

3. **Run with custom timeouts (optional):**

   ```bash
   docker run -p 8080:8080 \
     -e PORT=8080 \
     -e ALLOWED_DOMAINS="example.com,*.example.com" \
     -e SIGNATURE_LIFETIME="1h" \
     -e READ_TIMEOUT="15s" \
     -e WRITE_TIMEOUT="15s" \
     -e IDLE_TIMEOUT="120s" \
     -e CERT_DIAL_TIMEOUT="15s" \
     -e LOG_LEVEL="debug" \
     -e PRIVATE_KEY_PEM="$(cat private_key.pem)" \
     dynapins-server
   ```

### Option 2: Local Development

1. **Set environment variables:**

   ```bash
   export PORT=8080
   export ALLOWED_DOMAINS="google.com,*.google.com"
   export SIGNATURE_LIFETIME="1h"
   export PRIVATE_KEY_PEM=$(cat private_key.pem)
   ```

2. **Run the server:**

   ```bash
   go run ./cmd/server
   ```

## Configuration Examples

### Production Configuration

For production deployments with stricter timeouts and info logging:

```bash
export PORT=8080
export ALLOWED_DOMAINS="api.example.com,*.api.example.com,example.com"
export SIGNATURE_LIFETIME="30m"
export READ_TIMEOUT="5s"
export WRITE_TIMEOUT="5s"
export IDLE_TIMEOUT="30s"
export CERT_DIAL_TIMEOUT="10s"
export SHUTDOWN_TIMEOUT="15s"
export LOG_LEVEL="info"
export PRIVATE_KEY_PEM=$(cat /path/to/production-key.pem)
```

### Development Configuration

For local development with relaxed timeouts and debug logging:

```bash
export PORT=3000
export ALLOWED_DOMAINS="*.local,localhost,google.com,*.google.com"
export SIGNATURE_LIFETIME="1h"
export READ_TIMEOUT="30s"
export WRITE_TIMEOUT="30s"
export IDLE_TIMEOUT="120s"
export CERT_DIAL_TIMEOUT="30s"
export LOG_LEVEL="debug"
export PRIVATE_KEY_PEM=$(cat ./dev-key.pem)
```

### High-Traffic Configuration

For high-traffic scenarios with optimized timeouts:

```bash
export PORT=8080
export ALLOWED_DOMAINS="api.example.com"
export SIGNATURE_LIFETIME="15m"
export READ_TIMEOUT="3s"
export WRITE_TIMEOUT="3s"
export IDLE_TIMEOUT="20s"
export CERT_DIAL_TIMEOUT="5s"
export SHUTDOWN_TIMEOUT="30s"
export LOG_LEVEL="warn"
export PRIVATE_KEY_PEM=$(cat /path/to/key.pem)
```

## API Usage

### Get Certificate Pins

Retrieve signed certificate pins for a domain.

**Endpoint:** `GET /v1/pins`

**Query Parameters:**
- `domain` (required): The fully qualified domain name to get pins for
- `include-backup-pins` (optional): Include backup pin from intermediate cert (`true` or `false`, default: `false`)

**Example Request:**

```bash
# Get primary pin only (leaf certificate)
curl "http://localhost:8080/v1/pins?domain=example.com"

# Include backup pin (leaf + intermediate)
curl "http://localhost:8080/v1/pins?domain=example.com&include-backup-pins=true"
```

**Example Response (200 OK):**

```json
{
  "jws": "eyJhbGciOiJFUzI1NiIsImtpZCI6ImExYjJjM2Q0In0.eyJkb21haW4iOiJleGFtcGxlLmNvbSIsInBpbnMiOlsiYjdmM2U2YTFjMmQzZTRmNWE2YjdjOGQ5ZTBmMWEyYjNjNGQ1ZTZmN2E4YjljMGQxZTJmM2E0YjVjNmQ3ZThmOSJdLCJpYXQiOjE3Mjk1ODg4MDAsImV4cCI6MTcyOTU5MjQwMCwidHRsX3NlY29uZHMiOjM2MDB9.MEQCIG3..."
}
```

**JWS Token Contents** (when decoded):

Header:
```json
{
  "alg": "ES256",
  "kid": "a1b2c3d4"
}
```

Payload:
```json
{
  "domain": "example.com",
  "pins": [
    "b7f3e6a1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9"
  ],
  "iat": 1729588800,
  "exp": 1729592400,
  "ttl_seconds": 3600
}
```

**Error Responses:**

- **400 Bad Request**: Missing or invalid `domain` parameter
- **403 Forbidden**: Domain not in whitelist
- **422 Unprocessable Entity**: Failed to retrieve certificate for domain

### Health Check Endpoints

#### Liveness Check

**Endpoint:** `GET /health`

Simple liveness check for Kubernetes/Docker health monitoring.

```bash
curl "http://localhost:8080/health"
```

**Response (200 OK):**
```json
{
  "status": "healthy"
}
```

#### Readiness Check

**Endpoint:** `GET /readiness`

Readiness check that verifies crypto components are initialized.

```bash
curl "http://localhost:8080/readiness"
```

**Response (200 OK):**
```json
{
  "status": "ready",
  "allowed_domains": 3,
  "key_id": "a1b2c3d4"
}
```

**Response (503 Service Unavailable):**
```json
{
  "status": "not ready",
  "reason": "crypto keys not initialized"
}
```

## Documentation

### API Specification

Full OpenAPI 3.0 specification: [api/openapi.yaml](api/openapi.yaml)

Browse the interactive API docs:
- [OpenAPI Viewer](https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/Free-cat/dynapins-server/main/api/openapi.yaml)

### Deployment Examples

Ready-to-use configurations:
- **[Docker Compose](examples/docker-compose/)** - Simple setup for single-server deployments
- **[Kubernetes](examples/kubernetes/)** - Production-ready K8s manifests with health checks

### Additional Resources

- **[CHANGELOG.md](CHANGELOG.md)** - Version history and release notes
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - How to contribute to the project
- **[SECURITY.md](SECURITY.md)** - Security policy and best practices

## Development

### Using Make

The project includes a `Makefile` for common development tasks:

```bash
# Show all available commands
make help

# Build the server
make build

# Run tests
make test

# Run tests with coverage report
make test-coverage

# Format code
make fmt

# Run linters
make lint

# Build Docker image
make docker-build

# Clean build artifacts
make clean
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

## Domain Whitelist

The `ALLOWED_DOMAINS` configuration supports:

- **Exact matches**: `example.com` only allows `example.com`
- **Single-level wildcards**: `*.example.com` allows `api.example.com`, `www.example.com`, etc.
  - Does NOT match `api.v2.example.com` (too many levels)
  - Does NOT match `example.com` (base domain)

**Example:**

```bash
ALLOWED_DOMAINS="example.com,*.api.example.com,google.com"
```

This allows:
- `example.com` (exact)
- `v1.api.example.com` (wildcard)
- `v2.api.example.com` (wildcard)
- `google.com` (exact)

This does NOT allow:
- `www.example.com` (not in whitelist)
- `api.example.com` (wildcard requires one more level)
- `prod.v1.api.example.com` (too many levels for wildcard)

## Docker Images

Pre-built multi-platform images: [freecats/dynapins-server](https://hub.docker.com/r/freecats/dynapins-server)

**Platforms:** `linux/amd64`, `linux/arm64`

```bash
docker pull freecats/dynapins-server:latest

# Or specific version
docker pull freecats/dynapins-server:v0.2.0
```

### Build Your Own

```bash
docker build -t dynapins-server .
docker run -p 8080:8080 \
  -e ALLOWED_DOMAINS="example.com" \
  -e PRIVATE_KEY_PEM="$(cat private_key.pem)" \
  dynapins-server
```

## Architecture

```
pinning-server/
├── cmd/
│   └── server/          # Application entry point
│       └── main.go
├── internal/
│   ├── config/          # Configuration loading
│   ├── logger/          # Structured logging
│   ├── models/          # API data models
│   ├── domain/          # Domain validation
│   ├── cert/            # Certificate retrieval
│   ├── crypto/          # Cryptographic operations
│   └── server/          # HTTP server and handlers
├── Dockerfile
└── README.md
```

## Client Integration

### Verifying JWS Signatures

Clients **must** verify the JWS signature before trusting the pins. Here's how:

#### iOS (Swift)

```swift
import CryptoKit

// Your server's public key (extract with: openssl ec -in private_key.pem -pubout)
let publicKeyPEM = """
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...
-----END PUBLIC KEY-----
"""

// Parse and verify JWS using CryptoKit or a library like JOSESwift
// The dynapins-ios SDK handles this automatically
```

#### Android (Kotlin)

```kotlin
import java.security.KeyFactory
import java.security.spec.X509EncodedKeySpec
import java.util.Base64

// Your server's public key
val publicKeyPEM = """
    -----BEGIN PUBLIC KEY-----
    MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...
    -----END PUBLIC KEY-----
""".trimIndent()

// Parse and verify JWS
// The dynapins-android SDK handles this automatically
```

### Caching Strategy

- **Server-side caching**: Enabled by default (`CERT_CACHE_TTL=5m`)
  - Reduces TLS handshake overhead
  - Improves response time and reduces load
  
- **Client-side caching**: Based on signature expiry (`exp` claim in JWS)
  - Clients should cache pins until signature expires
  - Refresh pins before expiry to avoid gaps

## Security Considerations

1. **Private Key Protection**: Keep your ECDSA P-256 private key secure. Never commit it to version control.
2. **HTTPS Only**: This server should be deployed behind a reverse proxy with TLS termination.
3. **Whitelist Management**: Only add trusted domains to the whitelist.
4. **Signature Verification**: Clients **must** verify the JWS signature using the public key before trusting pins.
5. **Certificate Validation**: The server validates certificates during retrieval (no `InsecureSkipVerify`).
6. **IP Literal Blocking**: By default, IP addresses are rejected (`ALLOW_IP_LITERALS=false` for production).
7. **Non-root Execution**: Docker image runs as user 65532 (non-root) for security.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

Copyright (c) 2025 Artem Melnikov

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.

