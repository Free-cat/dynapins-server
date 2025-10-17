# Deployment Examples

Example configurations for deploying Dynapins Server.

## Available Examples

### [Docker Compose](./docker-compose/)

Simple setup for local development or single-server deployments.

```bash
cd docker-compose
export PRIVATE_KEY_PEM="$(cat private_key.pem)"
docker-compose up -d
```

### [Kubernetes](./kubernetes/)

Production-ready Kubernetes deployment with health checks and proper security context.

```bash
cd kubernetes
kubectl create secret generic dynapins-key --from-file=private_key.pem
kubectl apply -f deployment.yaml -f service.yaml
```

## Before Deploying

1. **Generate ECDSA P-256 key pair:**

```bash
# Private key
openssl ecparam -genkey -name prime256v1 -noout -out private_key.pem

# Public key (for clients)
openssl ec -in private_key.pem -pubout -out public_key.pem
```

2. **Configure allowed domains:**

Edit `ALLOWED_DOMAINS` in the example files to include your domains.

3. **Review security settings:**

- Set `ALLOW_IP_LITERALS=false` in production
- Use appropriate `SIGNATURE_LIFETIME` (recommended: 15m-1h)
- Enable rate limiting at reverse proxy level

## Need Help?

See the main [README](../README.md) for full documentation.


