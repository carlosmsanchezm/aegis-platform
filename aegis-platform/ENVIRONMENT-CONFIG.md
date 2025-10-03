# Backstage Environment Configuration

## Quick Start

### Local Development
```bash
# Switch to local config
cp app-config.local-dev.yaml app-config.local.yaml
yarn start
```
- Connects to local platform-api at `http://localhost:8080`
- Use for all local testing without GPU requirements

### Cloud Testing
```bash
# Switch to cloud config
cp app-config.cloud.yaml app-config.local.yaml
yarn start
```
- Connects to AWS EKS Load Balancer
- Use for GPU workloads and connectivity testing

## Configuration Files

- **`app-config.yaml`** - Base configuration (shared)
- **`app-config.local.yaml`** - Active config (gitignored, copy from templates below)
- **`app-config.local-dev.yaml`** - Template for local development
- **`app-config.cloud.yaml`** - Template for cloud deployment

## Testing Strategy

1. **Local First**: Do most development and testing locally
2. **Cloud When Needed**: Only test on cloud for:
   - GPU workload execution
   - Load balancer connectivity
   - Production-like environment validation

## How It Works

The proxy configuration in `app-config.local.yaml` determines where API calls are routed:

**Local**: `/api/proxy/aegis/*` → `http://localhost:8080/*`
**Cloud**: `/api/proxy/aegis/*` → `http://<load-balancer>:8080/*`

Backstage backend (port 7008) proxies requests to the appropriate target.
