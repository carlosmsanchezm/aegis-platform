# Keycloak Authentication Troubleshooting Guide

This document explains the Keycloak OIDC authentication configuration for the Aegis platform, common issues, and how to resolve them.

## Architecture Overview

```
┌─────────────────────┐     ┌──────────────────────┐     ┌─────────────────────┐
│   Backstage UI      │────▶│   Platform API       │────▶│     Keycloak        │
│  (localhost:3000)   │     │  (aegis-system)      │     │   (keycloak ns)     │
└─────────────────────┘     └──────────────────────┘     └─────────────────────┘
         │                            │                           │
         │                            │                           │
         ▼                            ▼                           ▼
   Gets token from            Validates token via         Issues tokens with
   Keycloak via               JWKS endpoint               iss: aegis-platform.tech
   localtest.me
```

### Cloudflare Tunnel Architecture (for Remote Spoke)

```
┌─────────────────────┐                    ┌─────────────────────┐
│  Remote EKS Spoke   │                    │   Local Cluster     │
│  (AWS us-east-1)    │                    │   (Docker Desktop)  │
└─────────────────────┘                    └─────────────────────┘
         │                                          │
         │ HTTPS                                    │
         ▼                                          ▼
┌─────────────────────┐                    ┌─────────────────────┐
│ keycloak.aegis-     │◀── Cloudflare ────▶│ aegis-services-     │
│ platform.tech       │     Tunnel         │ keycloak (keycloak) │
└─────────────────────┘                    └─────────────────────┘
         │
         │ Also used for
         ▼
┌─────────────────────┐
│ remote.aegis-       │◀── Platform API gRPC
│ platform.tech       │
└─────────────────────┘
```

## Key Configuration Points

### 1. Keycloak Realm Issuer

The Keycloak realm is configured with `keycloak.aegis-platform.tech` as the issuer URL. This is set in:
- `charts/aegis-services/files/keycloak/aegis-realm.json`

**All tokens issued by Keycloak will have:**
```json
{
  "iss": "https://keycloak.aegis-platform.tech/realms/aegis",
  ...
}
```

This issuer is used because:
1. Remote spokes (EKS clusters) access Keycloak via Cloudflare tunnel at `keycloak.aegis-platform.tech`
2. Tokens must have a consistent issuer regardless of where they're obtained

### 2. Platform API OIDC Configuration

The Platform API validates tokens using these environment variables:

| Variable | Value | Purpose |
|----------|-------|---------|
| `OIDC_ISSUER_URL` | `https://keycloak.aegis-platform.tech/realms/aegis` | Must match the `iss` claim in tokens |
| `OIDC_JWKS_URL` | `https://aegis-services-keycloak.keycloak.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/certs` | Internal endpoint to fetch signing keys |
| `OIDC_AUDIENCE` | `backstage,aegis-platform` | Valid audiences for tokens |

**CRITICAL:** The `OIDC_ISSUER_URL` must match exactly what's in the token's `iss` claim, NOT where you access Keycloak from.

### 3. Backstage Configuration

In `aegis-ui/app-config.local-dev.yaml`:

```yaml
backend:
  auth:
    externalAccess:
      - type: jwks
        options:
          # Use localtest.me to fetch JWKS (the actual reachable endpoint)
          url: https://keycloak.localtest.me/realms/aegis/protocol/openid-connect/certs
          # Issuer must match the "iss" claim in the token
          issuer: https://keycloak.aegis-platform.tech/realms/aegis
          audience: backstage

auth:
  providers:
    keycloak:
      development:
        # Use localtest.me for metadata since that's the accessible endpoint locally
        metadataUrl: https://keycloak.localtest.me/realms/aegis/.well-known/openid-configuration
        # Issuer must match token's iss claim
        issuer: https://keycloak.aegis-platform.tech/realms/aegis
```

## Common Issues and Solutions

### Issue 1: "invalid bearer token" - Issuer Mismatch

**Symptoms:**
```
{"error":"Unauthenticated","message":"invalid bearer token"}
```

**Cause:** The `OIDC_ISSUER_URL` doesn't match the `iss` claim in the token.

**Solution:** Ensure `OIDC_ISSUER_URL` is set to `https://keycloak.aegis-platform.tech/realms/aegis` (NOT `localtest.me`).

```yaml
# In charts/aegis-services/values/local.yaml
platformApi:
  env:
    OIDC_ISSUER_URL: "https://keycloak.aegis-platform.tech/realms/aegis"
```

### Issue 2: "failed to verify certificate" - TLS Certificate SAN Mismatch

**Symptoms:**
```
tls: failed to verify certificate: x509: certificate is valid for
keycloak.localtest.me, aegis-services-keycloak.keycloak.svc.cluster.local,
not aegis-services-keycloak-service.keycloak.svc.cluster.local
```

**Cause:** The JWKS URL uses a hostname not covered by the Keycloak TLS certificate's Subject Alternative Names (SANs).

**Solution:** Use the correct internal service name that matches the certificate:

```yaml
# CORRECT - matches certificate SAN
OIDC_JWKS_URL: "https://aegis-services-keycloak.keycloak.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/certs"

# WRONG - not in certificate SAN
OIDC_JWKS_URL: "https://aegis-services-keycloak-service.keycloak.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/certs"
```

**Quick fix if needed:**
```bash
kubectl set env deployment/aegis-services-platform-api -n aegis-system \
  OIDC_JWKS_URL="https://aegis-services-keycloak.keycloak.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/certs"
```

### Issue 3: "Missing credentials" in Backstage

**Symptoms:**
```
AuthenticationError: Missing credentials
```

**Cause:** The Backstage `SignInPage` component is not enabled, allowing guest access which fails on protected API calls.

**Solution:** Enable auto-login in `aegis-ui/packages/app/src/App.tsx`:

```tsx
components: {
  SignInPage: props => (
    <SignInPage {...props} auto provider={keycloakSignInProvider} />
  ),
},
```

### Issue 4: Keycloak Not Deployed

**Symptoms:** No pods in keycloak namespace after `helm upgrade`.

**Cause:** Using `local.yaml` instead of `local-tls.yaml`. The `local.yaml` file doesn't include the Keycloak deployment configuration.

**Solution:** Use the correct values file:
```bash
make deploy-local-tls  # Uses local-tls.yaml which includes Keycloak
```

## Cloudflare Tunnel Configuration

The Cloudflare tunnel (`cloudflared-config.yaml`) exposes:

| External URL | Internal Service |
|--------------|------------------|
| `keycloak.aegis-platform.tech` | `aegis-services-keycloak-service.keycloak.svc.cluster.local:8443` |
| `remote.aegis-platform.tech` | `aegis-services-platform-api.aegis-system.svc.cluster.local:8081` |

This allows remote spokes to:
1. Obtain tokens from Keycloak via `keycloak.aegis-platform.tech`
2. Connect to Platform API via `remote.aegis-platform.tech`

## Configuration Checklist

When setting up or troubleshooting authentication:

- [ ] Keycloak realm issuer is `https://keycloak.aegis-platform.tech/realms/aegis`
- [ ] Platform API `OIDC_ISSUER_URL` matches the realm issuer exactly
- [ ] Platform API `OIDC_JWKS_URL` uses hostname in the TLS certificate SAN (`aegis-services-keycloak`, NOT `aegis-services-keycloak-service`)
- [ ] Backstage `SignInPage` is enabled for auto-login
- [ ] Backstage `issuer` config matches the realm issuer
- [ ] Backstage `metadataUrl` and JWKS `url` use the accessible endpoint (`localtest.me` for local dev)
- [ ] Cloudflare tunnel is running if remote spokes need access

## Files Reference

| File | Purpose |
|------|---------|
| `charts/aegis-services/values/local.yaml` | Local dev without Keycloak deployment |
| `charts/aegis-services/values/local-tls.yaml` | Local dev WITH Keycloak deployment |
| `charts/aegis-services/files/keycloak/aegis-realm.json` | Keycloak realm configuration |
| `aegis-ui/app-config.local-dev.yaml` | Backstage local dev config |
| `aegis-ui/packages/app/src/App.tsx` | Backstage SignInPage config |
| `cloudflared-config.yaml` | Cloudflare tunnel routes |

## Debugging Commands

```bash
# Check Platform API OIDC config
kubectl get deployment aegis-services-platform-api -n aegis-system -o yaml | grep -A1 "OIDC_"

# Check Platform API auth logs
kubectl logs -n aegis-system deploy/aegis-services-platform-api | grep -E "auth|token|OIDC"

# Test Keycloak JWKS endpoint internally
kubectl run -n aegis-system debug-curl --rm -i --restart=Never --image=curlimages/curl:latest -- \
  curl -sk https://aegis-services-keycloak.keycloak.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/certs

# Check Keycloak's reported issuer
curl -sk https://keycloak.localtest.me/realms/aegis/.well-known/openid-configuration | jq .issuer

# Verify Cloudflare tunnel is running
kubectl logs -n aegis-system deploy/cloudflared --tail=20
```
