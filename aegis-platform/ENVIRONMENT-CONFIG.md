# Backstage Environment Configuration

## Quick Start

### Local Development (Default)
```bash
yarn dev
```
- Connects to local platform-api at `http://localhost:8080`
- Use for all local testing without GPU requirements

### Cloud Testing
```bash
yarn dev:cloud
```
- Connects to AWS EKS Load Balancer
- Use for GPU workloads and connectivity testing

## Configuration Files

- **`app-config.yaml`** - Base configuration (shared)
- **`app-config.local-dev.yaml`** - Local development (localhost:8080)
- **`app-config.cloud.yaml`** - Cloud deployment (AWS Load Balancer)

## Testing Strategy

1. **Local First**: Do most development and testing locally (`yarn dev`)
2. **Cloud When Needed**: Only test on cloud for:
   - GPU workload execution
   - Load balancer connectivity
   - Production-like environment validation

## How It Works

The proxy configuration in each environment file determines where API calls are routed:

**Local**: `/api/proxy/aegis/*` → `http://localhost:8080/*`
**Cloud**: `/api/proxy/aegis/*` → `http://<load-balancer>:8080/*`

Backstage backend (port 7008) proxies requests to the appropriate target.
