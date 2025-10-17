# Dynapins Server Server

A Go-based HTTP server that provides signed TLS certificate pins for dynamic SSL pinning. This server retrieves TLS certificates for whitelisted domains, generates SHA-256 hashes of their Subject Public Key Info (SPKI), and returns them in a signed response.

## Features

- **Dynamic SSL Pinning**: Get certificate pins for domains without hardcoding them
- **Signature-Verified Trust**: All responses are signed with Ed25519 for verification
- **Domain Whitelist**: Only serves pins for explicitly allowed domains (supports wildcards)
- **Stateless**: No database required, fully stateless operation
- **High Performance**: Built in Go with minimal dependencies

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
| `IDLE_TIMEOUT` | Maximum time to wait for the next request when keep-alives are enabled | No | `60s` | `60s`, `2m` |
| `SHUTDOWN_TIMEOUT` | Maximum time to wait for graceful server shutdown | No | `10s` | `10s`, `30s` |
| **Domain & Security** |
| `ALLOWED_DOMAINS` | Comma-separated list of domains and wildcards to allow | **Yes** | - | `"example.com,*.example.com,api.anotherexample.com"` |
| `SIGNATURE_LIFETIME` | The validity period of the generated signature | No | `1h` | `1h`, `30m`, `2h30m` |
| `PRIVATE_KEY_PEM` | The PEM-encoded Ed25519 private key for signing | **Yes** | - | `"-----BEGIN PRIVATE KEY-----..."` |
| **Certificate Retrieval** |
| `CERT_DIAL_TIMEOUT` | Maximum time to wait when connecting to retrieve certificates | No | `10s` | `10s`, `15s`, `30s` |
| **Logging** |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | No | `info` | `info`, `debug`, `error` |

### Duration Format

Duration values support Go's duration format:
- `s` = seconds (e.g., `30s`)
- `m` = minutes (e.g., `5m`)
- `h` = hours (e.g., `2h`)
- Combined: `1h30m`, `2h30m45s`

### Generating an Ed25519 Key Pair

To generate a new Ed25519 key pair for signing:

```bash
# Generate private key
openssl genpkey -algorithm Ed25519 -out private_key.pem

# Extract public key
openssl pkey -in private_key.pem -pubout -out public_key.pem
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

**Example Request:**

```bash
curl "http://localhost:8080/v1/pins?domain=example.com"
```

**Example Response (200 OK):**

```json
{
  "domain": "example.com",
  "pins": [
    "b7f3e6a1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9",
    "c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9"
  ],
  "created": "2025-10-17T08:00:00Z",
  "expires": "2025-10-17T09:00:00Z",
  "ttl_seconds": 3600,
  "keyId": "a1b2c3d4",
  "alg": "Ed25519",
  "signature": "MEQCIG3..."
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
docker pull freecats/dynapins-server:v1.0.0
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

## Security Considerations

1. **Private Key Protection**: Keep your Ed25519 private key secure. Never commit it to version control.
2. **HTTPS Only**: This server should be deployed behind a reverse proxy with TLS termination.
3. **Whitelist Management**: Only add trusted domains to the whitelist.
4. **Signature Verification**: Clients must verify the signature using the public key before trusting pins.
5. **Certificate Validation**: The server validates certificates during retrieval (no `InsecureSkipVerify`).

## License

This project is part of a Dynamic SSL Pinning system.

