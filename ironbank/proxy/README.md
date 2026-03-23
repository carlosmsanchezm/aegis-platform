# aegis-proxy

Aegis Platform reverse proxy for secure WebSocket workspace tunneling.

## Description

The proxy provides authenticated WebSocket tunneling from users to GPU development workspaces running on spoke clusters. It:

- Validates JWT tokens issued by Keycloak
- Routes WebSocket connections to the correct workspace pod
- Supports multiple concurrent workspace sessions
- Provides Prometheus metrics for observability

## Base Image

`registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal:9.7`

Builder: `registry1.dso.mil/ironbank/google/golang/ubi9/golang-1.24:1.24.13`

## Ports

| Port | Protocol | Description |
|------|----------|-------------|
| 8080 | HTTP/WS | Proxy endpoint (HTTP + WebSocket upgrade) |

## Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aegis-proxy
  namespace: aegis-system
spec:
  replicas: 1
  template:
    spec:
      containers:
        - name: proxy
          image: registry1.dso.mil/ironbank/aegis/proxy:1.0.0
          ports:
            - containerPort: 8080
          resources:
            requests:
              cpu: 100m
              memory: 64Mi
            limits:
              cpu: 500m
              memory: 256Mi
          securityContext:
            runAsUser: 65532
            runAsGroup: 65532
            runAsNonRoot: true
```

## Resource Requirements

| Resource | Request | Limit |
|----------|---------|-------|
| CPU | 100m | 500m |
| Memory | 64Mi | 256Mi |
