# Cloudflare Tunnel Setup for Aegis Platform

## Overview

This guide covers setting up Cloudflare Tunnel to expose your local Aegis Platform and Keycloak to remote spoke clusters (e.g., in AWS).

**⚠️ Important**: This setup is for **DEVELOPMENT ONLY**. Production air-gapped DoD environments do NOT use Cloudflare tunnels. See [Air-Gapped Deployment](#air-gapped-deployment) below.

## Architecture

```
┌─────────────────────────┐
│  Local Kubernetes       │
│  ┌──────────────────┐   │      ┌─────────────┐      ┌──────────────┐
│  │  Platform-API    │◄──┼──────┤ Cloudflared ├──────┤  Cloudflare  │
│  │  Keycloak        │   │      │   Tunnel    │      │     Edge     │
│  └──────────────────┘   │      └─────────────┘      └──────┬───────┘
└─────────────────────────┘                                   │
                                                              │ HTTPS
                                                              │
                                         ┌────────────────────▼───────┐
                                         │  Remote Spoke Clusters     │
                                         │  (AWS, etc.)               │
                                         │  ┌──────────────────┐      │
                                         │  │ Aegis Spoke      │      │
                                         │  │ Agent            │      │
                                         │  └──────────────────┘      │
                                         └────────────────────────────┘
```

## Prerequisites

1. **Cloudflare Account** with domain managed by Cloudflare (e.g., `aegis-platform.tech`)
2. **cloudflared CLI** installed:
   ```bash
   brew install cloudflared
   ```
3. **kubectl** with access to your local cluster
4. **Aegis Platform** deployed locally:
   ```bash
   make deploy-local-tls
   ```

## Quick Start

### 1. Login to Cloudflare

```bash
cloudflared tunnel login
```

This opens a browser to authenticate. Select your domain.

### 2. Run Setup Script

```bash
./scripts/setup-cloudflare-tunnels.sh
```

**That's it!** The script will:
- Create the tunnel (if it doesn't exist)
- Configure DNS routes
- Deploy cloudflared to Kubernetes
- Verify connectivity

### 3. Update Spoke Configuration

Use the endpoints shown in the script output:

```yaml
spokeAgent:
  grpc:
    endpoint: remote.aegis-platform.tech:443
    insecure: false
    skipVerify: true
    serverName: remote.aegis-platform.tech

  oidc:
    tokenUrl: https://keycloak.aegis-platform.tech/realms/aegis/protocol/openid-connect/token
    clientId: spoke-agent
    clientSecret: <your-secret>
```

## Advanced Usage

### Custom Domain

```bash
./scripts/setup-cloudflare-tunnels.sh --domain example.com
```

### Custom Hostnames

```bash
./scripts/setup-cloudflare-tunnels.sh \
  --platform-host api \
  --keycloak-host auth
```

This creates:
- `api.aegis-platform.tech` → Platform API
- `auth.aegis-platform.tech` → Keycloak

### Clean Up and Recreate

```bash
./scripts/setup-cloudflare-tunnels.sh --clean
```

### Script Options

```
OPTIONS:
    -t, --tunnel-name NAME      Tunnel name (default: aegis-platform)
    -d, --domain DOMAIN         Domain name (default: aegis-platform.tech)
    -p, --platform-host HOST    Platform hostname (default: remote)
    -k, --keycloak-host HOST    Keycloak hostname (default: keycloak)
    -n, --namespace NAMESPACE   Kubernetes namespace (default: aegis-system)
    -c, --clean                 Clean up existing tunnel and recreate
    -h, --help                  Show help message
```

## Verification

### Test Endpoints

```bash
# Platform API (should return gRPC error - that's correct!)
curl -I https://remote.aegis-platform.tech

# Keycloak (should redirect to /admin/)
curl -I https://keycloak.aegis-platform.tech

# Keycloak token endpoint (should return 405 Method Not Allowed for HEAD)
curl -I https://keycloak.aegis-platform.tech/realms/aegis/protocol/openid-connect/token
```

### Check Tunnel Status

```bash
# View tunnel info
cloudflared tunnel info aegis-platform

# Check cloudflared logs
kubectl logs -n aegis-system deployment/cloudflared -f

# Check tunnel connections (should show 4 active)
cloudflared tunnel info aegis-platform | grep connection
```

### Test from Spoke

From your spoke cluster in AWS:

```bash
# Test gRPC endpoint
grpcurl -insecure remote.aegis-platform.tech:443 list

# Expected: "Unauthenticated: missing bearer token" (this is good!)

# Test OIDC token endpoint
curl -X POST https://keycloak.aegis-platform.tech/realms/aegis/protocol/openid-connect/token \
  -d grant_type=client_credentials \
  -d client_id=spoke-agent \
  -d client_secret=<secret>

# Expected: JSON with access_token
```

## Troubleshooting

### DNS Not Resolving

DNS propagation can take 1-5 minutes. Check:

```bash
# Check DNS record
dig remote.aegis-platform.tech

# Should return Cloudflare IPs (104.x.x.x or 172.x.x.x)
```

### 502 Bad Gateway

- Check cloudflared logs: `kubectl logs -n aegis-system deployment/cloudflared`
- Verify services are running: `kubectl get pods -n aegis-system`
- Check service names match in config

### Connection Refused

- Ensure gRPC is enabled in Cloudflare:
  - Go to Cloudflare Dashboard → domain → Network
  - Enable "gRPC"
- Verify DNS records are **Proxied** (orange cloud)

### Authentication Errors

If spoke shows "missing bearer token":
- ✅ This is correct! The tunnel is working
- Implement OIDC authentication (see Jira ticket)

If spoke shows "HTTP/1.1 header" errors:
- Check `AEGIS_CP_GRPC_INSECURE` is set to `false`
- Check `AEGIS_CP_GRPC_SKIP_VERIFY` is set to `true`
- Check `AEGIS_CP_GRPC_SERVER_NAME` matches hostname

## Maintenance

### Restart Cloudflared

```bash
kubectl rollout restart deployment/cloudflared -n aegis-system
```

### Update Configuration

1. Edit `cloudflared-config.yaml`
2. Apply changes:
   ```bash
   kubectl apply -f cloudflared-config.yaml
   kubectl rollout restart deployment/cloudflared -n aegis-system
   ```

### Add New Service

To expose another service (e.g., Backstage):

1. Create DNS route:
   ```bash
   cloudflared tunnel route dns aegis-platform backstage
   ```

2. Update `cloudflared-config.yaml`:
   ```yaml
   ingress:
     # ... existing entries ...
     - hostname: backstage.aegis-platform.tech
       service: http://backstage.aegis-system.svc.cluster.local:7007
     - service: http_status:404
   ```

3. Apply and restart:
   ```bash
   kubectl apply -f cloudflared-config.yaml
   kubectl rollout restart deployment/cloudflared -n aegis-system
   ```

## Cost

- **Cloudflare Tunnel**: FREE (included in Free plan)
- **Cloudflare DNS**: FREE
- **No bandwidth charges** from Cloudflare

## Security Considerations

### Development (Current Setup)

- ✅ TLS encryption via Cloudflare
- ✅ gRPC support enabled
- ⚠️ Public internet exposure (acceptable for dev)
- ⚠️ Skip TLS verification (acceptable for dev)

### Production Requirements

**DO NOT use Cloudflare tunnels in production air-gapped environments!**

Instead:
- Deploy Platform-API inside the enclave
- Use internal Kubernetes services or LoadBalancers
- No public internet exposure
- Proper TLS with DoD PKI certificates
- See [Air-Gapped Deployment](#air-gapped-deployment)

## Air-Gapped Deployment

For production DoD environments, the architecture is completely different:

```
┌─────────────────────────────────────────────┐
│  Secure Enclave (No Internet)               │
│                                             │
│  ┌────────────┐         ┌────────────┐     │
│  │ Platform   │◄────────┤   Spoke    │     │
│  │   API      │ Internal│  Cluster   │     │
│  └────────────┘   k8s   └────────────┘     │
│                   DNS                       │
└─────────────────────────────────────────────┘
```

**Configuration for Air-Gapped**:

```yaml
# No Cloudflare tunnel needed!
spokeAgent:
  grpc:
    # Internal service DNS
    endpoint: aegis-platform-api.aegis-system.svc.cluster.local:8081
    insecure: false
    skipVerify: false  # Proper TLS validation

  oidc:
    # Internal Keycloak
    tokenUrl: https://keycloak.aegis-system.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/token

  tls:
    # DoD PKI certificates
    caCertBundle: |
      -----BEGIN CERTIFICATE-----
      [DoD Root CA 3]
      -----END CERTIFICATE-----
```

## Files

- `scripts/setup-cloudflare-tunnels.sh` - Setup automation
- `cloudflared-config.yaml` - Generated Kubernetes manifest
- `~/.cloudflared/cert.pem` - Cloudflare auth certificate
- `~/.cloudflared/<tunnel-id>.json` - Tunnel credentials

## References

- [Cloudflare Tunnel Documentation](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- [cloudflared CLI Reference](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/tunnel-guide/)
- [gRPC Support in Cloudflare](https://developers.cloudflare.com/cloudflare-one/applications/non-http/arbitrary-tcp/grpc/)

## Support

For issues with:
- **Cloudflare Tunnel**: Check cloudflared logs and tunnel status
- **Kubernetes Resources**: Check pod logs and events
- **DNS Resolution**: Wait 5 minutes for propagation, check Cloudflare dashboard
- **Authentication**: See authentication implementation Jira ticket

## Cleanup

To completely remove the tunnel:

```bash
# Delete Kubernetes resources
kubectl delete deployment cloudflared -n aegis-system
kubectl delete configmap cloudflared-config -n aegis-system
kubectl delete secret cloudflared-credentials -n aegis-system

# Delete DNS routes (in Cloudflare dashboard or CLI)
# Delete tunnel
cloudflared tunnel delete aegis-platform

# Delete local files
rm cloudflared-config.yaml
rm ~/.cloudflared/<tunnel-id>.json
```

---

**Remember**: This setup is for development convenience. Production air-gapped deployments use internal networking only! 🔒
