# Kubernetes Example

Deploy Dynapins Server to Kubernetes.

## Quick Start

1. **Generate a private key:**

```bash
openssl ecparam -genkey -name prime256v1 -noout -out private_key.pem
```

2. **Create secret from file:**

```bash
kubectl create secret generic dynapins-key \
  --from-file=private_key.pem=private_key.pem
```

Or edit `secret.yaml` and apply it:

```bash
kubectl apply -f secret.yaml
```

3. **Deploy:**

```bash
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

4. **Check status:**

```bash
kubectl get pods -l app=dynapins-server
kubectl logs -l app=dynapins-server -f
```

5. **Test (from inside cluster):**

```bash
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl "http://dynapins-server/v1/pins?domain=example.com"
```

## Expose Externally

### Option 1: Port Forward (dev/testing)

```bash
kubectl port-forward svc/dynapins-server 8080:80
curl "http://localhost:8080/v1/pins?domain=example.com"
```

### Option 2: Ingress (production)

Create `ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dynapins-server
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.example.com
    secretName: dynapins-tls
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: dynapins-server
            port:
              number: 80
```

Apply:

```bash
kubectl apply -f ingress.yaml
```

## Cleanup

```bash
kubectl delete -f service.yaml
kubectl delete -f deployment.yaml
kubectl delete secret dynapins-key
```



