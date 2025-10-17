# Security

## Reporting Security Issues

If you discover a security vulnerability, please use [GitHub Security Advisories](https://github.com/Free-cat/dynapins-server/security/advisories/new) to report it privately.

Alternatively, you can open an issue in the repository if the vulnerability is not critical.

## Supported Versions

Only the latest version receives security updates. Please upgrade to the newest release.

| Version | Supported |
| ------- | --------- |
| 0.2.x   | ✅        |
| < 0.2   | ❌        |

## Deployment Best Practices

### 1. Protect Your Private Key

```bash
# Never commit private keys to git
# Use environment variables or secret management
export PRIVATE_KEY_PEM="$(cat private_key.pem)"

# For Kubernetes, use secrets
kubectl create secret generic dynapins-key --from-file=private_key.pem
```

### 2. Production Configuration

```bash
# Recommended production settings
export ALLOWED_DOMAINS="api.example.com,*.api.example.com"
export SIGNATURE_LIFETIME="15m"        # Shorter is more secure
export ALLOW_IP_LITERALS="false"        # Block IP addresses
export READ_HEADER_TIMEOUT="3s"        # Prevent Slowloris attacks
export LOG_LEVEL="info"
```

### 3. Run Behind HTTPS

Always deploy behind a reverse proxy with TLS:

```nginx
# Nginx example with rate limiting
limit_req_zone $binary_remote_addr zone=dynapins:10m rate=10r/s;

server {
    listen 443 ssl http2;
    server_name api.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        limit_req zone=dynapins burst=20 nodelay;
        proxy_pass http://localhost:8080;
    }
}
```

### 4. Client Must Verify JWS Signatures

Mobile clients **must** verify the JWS signature before trusting pins. See the iOS and Android SDK documentation for implementation details.

## Known Limitations

- **No built-in rate limiting** - implement at reverse proxy/API gateway level
- **Certificate rotation** - clients need backup pins for graceful rotation
- **Caching** - balance `CERT_CACHE_TTL` and `SIGNATURE_LIFETIME` for your needs

