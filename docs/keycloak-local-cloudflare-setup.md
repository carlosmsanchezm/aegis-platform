# Keycloak Local + Cloudflare Tunnel Configuration

**Last Updated:** December 2, 2025
**Status:** ✅ Working Configuration - DO NOT MODIFY WITHOUT READING THIS FIRST

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│  Local Docker Desktop Kubernetes                            │
│                                                              │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐        │
│  │  Keycloak   │  │ Platform API │  │    Proxy    │        │
│  │  (Port 8443)│  │              │  │             │        │
│  └──────┬──────┘  └──────────────┘  └─────────────┘        │
│         │                                                    │
│  ┌──────▼─────────────────────────────────────────────┐    │
│  │  Cloudflare Tunnel (cloudflared pod)               │    │
│  │  Exposes: keycloak.aegis-platform.tech             │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
           ▲                              ▲
           │                              │
    Internal (fast)              External (via Cloudflare)
           │                              │
    ┌──────┴──────────┐          ┌───────┴──────────────┐
    │   Backstage     │          │   External Users     │
    │  (localhost)    │          │   Remote Services    │
    └─────────────────┘          └──────────────────────┘
  keycloak.localtest.me        keycloak.aegis-platform.tech
```

### Components

1. **Keycloak**: Authentication server running in local Docker Desktop k8s
   - Internal access: `https://keycloak.localtest.me`
   - External access: `https://keycloak.aegis-platform.tech`

2. **Backstage (aegis-ui)**: Developer portal running on local laptop
   - Uses `keycloak.localtest.me` for authentication (fast, direct)

3. **Cloudflare Tunnel**: Exposes local services to the internet
   - Routes `keycloak.aegis-platform.tech` → local Keycloak service

4. **Remote AWS EKS**: Workload clusters for ML workspaces
   - Managed by local platform-api
   - Does NOT run Keycloak

## Current Working Configuration

### 1. Keycloak Helm Values
**File:** `charts/aegis-services/values/local-tls.yaml`

```yaml
keycloak:
  hostname:
    hostname: https://keycloak.localtest.me
    admin: ""
    adminUrl: https://keycloak.localtest.me
    strict: false
  additionalOptions:
    - name: hostname-admin
      value: https://keycloak.localtest.me
    - name: hostname-admin-url
      value: https://keycloak.localtest.me/
```

**⚠️ Note:** The `https://` prefix is technically incorrect for Keycloak 26.2, but it works. Don't change it.

### 2. Backstage Configuration
**File:** `aegis-ui/app-config.local.yaml`

```yaml
auth:
  session:
    secret: dev-session-secret
    cookie:
      secure: false
      sameSite: lax
      domain: localhost
  providers:
    keycloak:
      development:
        clientId: backstage
        clientSecret: local-backstage-client-secret
        issuer: https://keycloak.localtest.me/realms/aegis
        metadataUrl: https://keycloak.localtest.me/realms/aegis/.well-known/openid-configuration
```

### 3. Cloudflare Tunnel Configuration
**File:** `cloudflared-config.yaml`

```yaml
- hostname: keycloak.aegis-platform.tech
  service: https://aegis-services-keycloak-service.keycloak.svc.cluster.local:8443
  originRequest:
    noTLSVerify: true
    httpHostHeader: keycloak.aegis-platform.tech
    originServerName: keycloak.aegis-platform.tech
```

### 4. Realm Configuration
**File:** `charts/aegis-services/files/keycloak/aegis-realm.json`

The realm JSON is the source of truth. After deployment, the realm is imported into Keycloak.

**Critical:** Do NOT modify realm settings via the Keycloak admin API in development. Changes will be lost on redeploy and create inconsistencies.

## ❌ What NOT to Change

**NEVER modify these without understanding the consequences:**

1. ❌ Keycloak `hostname` configuration in Helm values
   - Keep: `https://keycloak.localtest.me`
   - Changing this breaks local Backstage authentication

2. ❌ Backstage OIDC `issuer` and `metadataUrl`
   - Keep: `https://keycloak.localtest.me/realms/aegis`
   - Changing this causes issuer mismatch errors

3. ❌ Realm `frontendUrl` via Keycloak Admin API
   - The realm configuration should only come from the JSON file
   - API changes create inconsistencies that persist in the database

4. ❌ Session cookie `domain` in Backstage
   - Keep: `domain: localhost`
   - Removing this breaks cookie persistence in popup OAuth flow

## ✅ What You CAN Change Safely

1. ✅ **Cloudflare tunnel routes** for other services
   - Add platform-api, proxy, or other service routes

2. ✅ **Platform-api configuration** for remote EKS clusters
   - Configure remote cluster connections
   - Update kubeconfig secrets

3. ✅ **Keycloak client configurations** via realm JSON
   - Add new clients
   - Update redirect URIs
   - Modify client secrets
   - **Then redeploy:** `make clean-local && make deploy-local-tls`

4. ✅ **Backstage plugins and features**
   - Add new plugins
   - Configure integrations
   - Update UI components

## Access Patterns

### Local Access (Fast, Direct)
- **Backstage → Keycloak**: Uses `keycloak.localtest.me`
- **Platform-api → Keycloak**: Uses cluster-local service DNS
- **DNS Resolution**: `keycloak.localtest.me` → `127.0.0.1`
- **Benefits**: Fast, no external dependencies

### External Access (Via Cloudflare Tunnel)
- **External Users → Keycloak**: Uses `keycloak.aegis-platform.tech`
- **Routing**: Internet → Cloudflare → Tunnel → Local cluster
- **Use Cases**:
  - Remote developers accessing login page
  - External services authenticating
  - Testing production-like URLs

## Known Limitations

### External Service Authentication Issue

When external services access Keycloak via `keycloak.aegis-platform.tech`, they receive OIDC metadata with internal URLs:

```json
{
  "issuer": "https://keycloak.localtest.me/realms/aegis",
  "token-service": "https://keycloak.localtest.me/realms/aegis/protocol/openid-connect"
}
```

**Impact:**
- ✅ Browser-based authentication works (users can see login page)
- ✅ Local services work fine (use `keycloak.localtest.me` directly)
- ❌ External services can't resolve `keycloak.localtest.me` URLs

**Workarounds:**
1. External services should use `keycloak.localtest.me` if they can resolve it
2. Set up DNS override for external services to resolve `keycloak.localtest.me` to Cloudflare tunnel IP
3. Configure a reverse proxy to rewrite URLs (not implemented)
4. Accept limitation and only use Cloudflare tunnel for browser-based access

## Deployment Commands

### Normal Deployment
```bash
make deploy-local-tls
```

### Clean Slate Deployment (When Things Break)
```bash
make clean-local
make deploy-local-tls
```

**What this does:**
1. Deletes all Keycloak resources (pods, CRs, PVCs)
2. Fresh realm import from `aegis-realm.json`
3. Resets all configuration to git repo state
4. Clears any stale API modifications

### Start Backstage
```bash
cd ~/code/aegis-ui
NODE_EXTRA_CA_CERTS=~/aegis-local-trust.pem yarn dev
```

Then access: http://localhost:3000

## Troubleshooting

### Authentication Errors

**Symptom:** "Invalid client or Invalid client credentials"

**Causes:**
1. Realm configuration inconsistency (API modifications)
2. Issuer mismatch between Backstage and Keycloak
3. Session cookie issues

**Solution:**
```bash
# 1. Clean redeploy Keycloak
make clean-local
make deploy-local-tls

# 2. Verify Keycloak issuer
curl -s https://keycloak.localtest.me/realms/aegis/.well-known/openid-configuration | jq -r '.issuer'
# Should return: https://keycloak.localtest.me/realms/aegis

# 3. Verify Backstage config matches
grep "issuer:" ~/code/aegis-ui/app-config.local.yaml

# 4. Clear browser cookies
# Chrome DevTools (F12) → Application → Cookies → Delete All

# 5. Restart Backstage
cd ~/code/aegis-ui
NODE_EXTRA_CA_CERTS=~/aegis-local-trust.pem yarn dev
```

### "Missing session cookie" Error

**Causes:**
1. Browser blocking cookies in popup
2. Session cookie domain mismatch
3. Incognito mode with strict settings

**Solutions:**
1. Allow popups for `localhost:3000`
2. Ensure `domain: localhost` is set in Backstage config
3. Try regular browser window (not incognito)
4. Clear all browser data and restart

### Path Duplication in OIDC URLs

**Symptom:** URLs like `https://keycloak.localtest.me/realms/aegis/realms/aegis`

**Cause:** Incorrect `frontendUrl` in realm configuration

**Solution:**
```bash
# The realm JSON should NOT have frontendUrl set, or it should be:
# "frontendUrl": "https://keycloak.localtest.me"
# (without /realms/aegis at the end)

# Fix: Clean redeploy
make clean-local
make deploy-local-tls
```

### Cloudflare Tunnel Not Working

**Check tunnel status:**
```bash
kubectl -n aegis-system get pods -l app=cloudflared
kubectl -n aegis-system logs -l app=cloudflared --tail=50
```

**Restart tunnel:**
```bash
kubectl -n aegis-system rollout restart deploy/cloudflared
```

**Test external access:**
```bash
curl -s https://keycloak.aegis-platform.tech/realms/aegis/ | jq '.realm'
# Should return: "aegis"
```

## Verification Checklist

After deployment or configuration changes:

- [ ] Keycloak pod is running
  ```bash
  kubectl -n keycloak get pods -l app=keycloak
  ```

- [ ] Keycloak is accessible locally
  ```bash
  curl -k -s https://keycloak.localtest.me/realms/aegis/ | jq '.realm'
  ```

- [ ] Keycloak issuer is correct
  ```bash
  curl -s https://keycloak.localtest.me/realms/aegis/.well-known/openid-configuration | jq -r '.issuer'
  # Expected: https://keycloak.localtest.me/realms/aegis
  ```

- [ ] Cloudflare tunnel is running
  ```bash
  kubectl -n aegis-system get pods -l app=cloudflared
  ```

- [ ] External access works
  ```bash
  curl -s https://keycloak.aegis-platform.tech/realms/aegis/ | jq '.realm'
  ```

- [ ] Backstage config matches Keycloak issuer
  ```bash
  grep "issuer:" ~/code/aegis-ui/app-config.local.yaml
  ```

- [ ] Backstage starts without errors
  ```bash
  cd ~/code/aegis-ui && NODE_EXTRA_CA_CERTS=~/aegis-local-trust.pem yarn dev
  ```

- [ ] Can login to Backstage
  - Go to http://localhost:3000
  - Click "Sign In"
  - Redirects to Keycloak
  - Login succeeds and returns to Backstage

## Important Files Reference

| File | Purpose | Can Modify? |
|------|---------|-------------|
| `charts/aegis-services/values/local-tls.yaml` | Keycloak config for local TLS | ❌ No (hostname) |
| `charts/aegis-services/values/common.yaml` | Common Keycloak config | ❌ No (hostname) |
| `charts/aegis-services/files/keycloak/aegis-realm.json` | Realm definition | ✅ Yes (clients, users) |
| `aegis-ui/app-config.local.yaml` | Backstage local config | ❌ No (OIDC settings) |
| `cloudflared-config.yaml` | Cloudflare tunnel routes | ✅ Yes (add routes) |

## When to Use Clean Redeploy

Use `make clean-local && make deploy-local-tls` when:

1. ✅ Authentication completely broken
2. ✅ Made realm changes via Keycloak Admin UI/API
3. ✅ Modified realm JSON and want to apply changes
4. ✅ Issuer mismatch errors persist
5. ✅ Path duplication in OIDC URLs
6. ❌ NOT for: Simple config changes to non-Keycloak services

## Getting Help

If authentication breaks again:

1. **Don't panic** - it can be fixed with a clean redeploy
2. **Check this document** for troubleshooting steps
3. **Verify the basics:**
   - Is Keycloak running?
   - What's the issuer?
   - Does Backstage config match?
4. **Clean redeploy** if configuration is inconsistent
5. **Document what broke** - add to this file for future reference

## Summary

**Golden Rule:** If it's working, don't change the Keycloak or Backstage OIDC configuration.

The current configuration is a working state achieved after extensive troubleshooting. While not textbook-perfect (the `https://` in hostname is technically wrong), it works reliably.

Focus configuration efforts on:
- Remote EKS cluster connections
- Platform-api workload management
- Backstage plugins and features
- Cloudflare tunnel routes for other services

Leave the auth configuration alone unless absolutely necessary.
