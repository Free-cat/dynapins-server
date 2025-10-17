# Dynamic SSL Pinning API

**Provides signed TLS certificate pins for mobile applications with dynamic SSL pinning.**

[![Docker Image Size](https://img.shields.io/docker/image-size/freecats/dynapins-server/latest)](https://hub.docker.com/r/freecats/dynapins-server)
[![Docker Pulls](https://img.shields.io/docker/pulls/freecats/dynapins-server)](https://hub.docker.com/r/freecats/dynapins-server)

## Quick Start

```bash
# Generate ECDSA P-256 key
openssl ecparam -genkey -name prime256v1 -noout -out private_key.pem

# Run the server
docker run -p 8080:8080 \
  -e ALLOWED_DOMAINS="example.com,*.example.com" \
  -e PRIVATE_KEY_PEM="$(cat private_key.pem)" \
  freecats/dynapins-server:latest
```

Test it:
```bash
curl "http://localhost:8080/v1/pins?domain=example.com"
```

## Features

- 🔐 **ECDSA P-256 Signing** - JWS (ES256) signatures for certificate pins
- 🌐 **Domain Whitelist** - Control which domains are allowed (with wildcard support)
- 🚀 **High Performance** - Built in Go, certificate caching, <200ms response times
- 📦 **Stateless** - No database required
- 🏥 **Health Checks** - Built-in `/health` and `/readiness` endpoints
- 🔧 **Fully Configurable** - All settings via environment variables
- 🔒 **Security Hardened** - Non-root user, IP literal blocking, configurable timeouts

## Supported Platforms

- `linux/amd64` - Intel/AMD (x86_64)
- `linux/arm64` - ARM 64-bit (Apple Silicon, AWS Graviton, Raspberry Pi 4+)

## Configuration

All configuration is done via environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | Server port |
| `ALLOWED_DOMAINS` | **Yes** | - | Comma-separated domains (supports wildcards) |
| `PRIVATE_KEY_PEM` | **Yes** | - | ECDSA P-256 private key (PEM format) |
| `SIGNATURE_LIFETIME` | No | `1h` | JWS signature validity period |
| `READ_TIMEOUT` | No | `10s` | HTTP read timeout |
| `WRITE_TIMEOUT` | No | `10s` | HTTP write timeout |
| `READ_HEADER_TIMEOUT` | No | `5s` | Read header timeout (Slowloris protection) |
| `MAX_HEADER_BYTES` | No | `1048576` | Max header size in bytes (1MB) |
| `CERT_DIAL_TIMEOUT` | No | `10s` | TLS connection timeout |
| `CERT_CACHE_TTL` | No | `5m` | Certificate cache TTL (0 to disable) |
| `ALLOW_IP_LITERALS` | No | `false` | Allow IP addresses (dev only) |
| `LOG_LEVEL` | No | `info` | Log level (debug/info/warn/error) |

## API Endpoints

### Get Certificate Pins
```bash
# Get primary pin only
GET /v1/pins?domain=example.com

# Include backup pin (leaf + intermediate)
GET /v1/pins?domain=example.com&include-backup-pins=true
```

**Response:**
```json
{
  "jws": "eyJhbGciOiJFUzI1NiIsImtpZCI6ImExYjJjM2Q0In0.eyJkb21haW4i..."
}
```

**Decoded JWS Payload:**
```json
{
  "domain": "example.com",
  "pins": ["b7f3e6a1c2d3..."],
  "iat": 1729588800,
  "exp": 1729592400,
  "ttl_seconds": 3600
}
```

### Health Check
```bash
GET /health          # Liveness probe
GET /readiness       # Readiness probe
```

## Examples

### Basic Usage

```bash
docker run -d -p 8080:8080 \
  -e ALLOWED_DOMAINS="api.example.com" \
  -e PRIVATE_KEY_PEM="$(cat private_key.pem)" \
  --name ssl-pinning \
  freecats/dynapins-server:latest
```

### With Wildcard Domains

```bash
docker run -d -p 8080:8080 \
  -e ALLOWED_DOMAINS="example.com,*.api.example.com,*.cdn.example.com" \
  -e SIGNATURE_LIFETIME="30m" \
  -e PRIVATE_KEY_PEM="$(cat private_key.pem)" \
  freecats/dynapins-server:latest
```

### Production Configuration

```bash
docker run -d -p 8080:8080 \
  -e ALLOWED_DOMAINS="api.example.com,*.api.example.com" \
  -e SIGNATURE_LIFETIME="15m" \
  -e READ_TIMEOUT="5s" \
  -e WRITE_TIMEOUT="5s" \
  -e CERT_DIAL_TIMEOUT="10s" \
  -e LOG_LEVEL="warn" \
  -e PRIVATE_KEY_PEM="$(cat private_key.pem)" \
  --restart unless-stopped \
  --name ssl-pinning-prod \
  freecats/dynapins-server:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  ssl-pinning:
    image: freecats/dynapins-server:latest
    ports:
      - "8080:8080"
    environment:
      ALLOWED_DOMAINS: "example.com,*.example.com"
      SIGNATURE_LIFETIME: "1h"
      LOG_LEVEL: "info"
      PRIVATE_KEY_PEM: |
        -----BEGIN PRIVATE KEY-----
        MC4CAQAwBQYDK2VwBCIEIC...
        -----END PRIVATE KEY-----
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
```

## Generate ECDSA P-256 Key

```bash
# Generate private key (ECDSA P-256)
openssl ecparam -genkey -name prime256v1 -noout -out private_key.pem

# Extract public key (for client verification)
openssl ec -in private_key.pem -pubout -out public_key.pem

# Optional: Convert to PKCS#8 format
openssl pkcs8 -topk8 -nocrypt -in private_key.pem -out private_key_pkcs8.pem
```

## Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ssl-pinning-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ssl-pinning
  template:
    metadata:
      labels:
        app: ssl-pinning
    spec:
      containers:
      - name: api
        image: freecats/dynapins-server:latest
        ports:
        - containerPort: 8080
        env:
        - name: ALLOWED_DOMAINS
          value: "api.example.com,*.api.example.com"
        - name: PRIVATE_KEY_PEM
          valueFrom:
            secretKeyRef:
              name: ssl-pinning-key
              key: private_key
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readiness
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Domain Whitelist

The `ALLOWED_DOMAINS` configuration supports:

- **Exact matches**: `example.com` allows only `example.com`
- **Single-level wildcards**: `*.example.com` allows `api.example.com`, `www.example.com` but NOT `api.v2.example.com`

Example:
```bash
ALLOWED_DOMAINS="example.com,*.api.example.com,cdn.example.org"
```

This allows:
- ✅ `example.com`
- ✅ `v1.api.example.com`
- ✅ `v2.api.example.com`
- ✅ `cdn.example.org`

But NOT:
- ❌ `www.example.com` (not in whitelist)
- ❌ `api.example.com` (wildcard requires one subdomain)
- ❌ `prod.v1.api.example.com` (too many levels)

## Security

⚠️ **Important Security Considerations:**

1. **Never expose private keys** - Keep `PRIVATE_KEY_PEM` secure
2. **Use Kubernetes Secrets** or similar for production
3. **Run behind HTTPS** - Use reverse proxy with TLS
4. **Restrict allowed domains** - Only whitelist domains you control
5. **Verify JWS signatures** - Clients must verify signatures before trusting pins
6. **Monitor access logs** - Track usage patterns
7. **Rotate keys periodically** - Update keys and distribute new public keys to clients
8. **Disable IP literals** - Keep `ALLOW_IP_LITERALS=false` in production

## Performance

- **Response Time**: <200ms average
- **Throughput**: 1000+ requests/minute per instance
- **Resource Usage**: ~50MB RAM, minimal CPU
- **Concurrent Connections**: Handles thousands

## Available Tags

- `latest` - Latest stable release
- `v0.0.1`, `v0.0` - Semantic version tags
- `main` - Latest development build
- `sha-<commit>` - Specific commit builds

## Links

- 📖 **Full Documentation**: [GitHub Repository](https://github.com/Free-cat/ssl_pinning)
- 🐛 **Issues & Support**: [GitHub Issues](https://github.com/Free-cat/ssl_pinning/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/Free-cat/ssl_pinning/discussions)

## License

Open source project. See repository for details.

---

**Made with ❤️ for mobile app security**

