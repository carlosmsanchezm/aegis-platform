# Aegis Platform Discovery Endpoint

## Overview

The Platform API serves a public discovery document at `/api/v1/discovery` that enables VS Code extension (Sovran) clients to auto-configure all connection parameters from a single platform URL.

## Endpoint

```
GET /api/v1/discovery
```

**Authentication:** None required. This endpoint returns public metadata only.

**Response:**
```json
{
  "platform_version": "1.0.0",
  "grpc_endpoint": "platform-api.aegis-platform.tech:8081",
  "auth": {
    "authority": "https://keycloak.aegis-platform.tech/realms/aegis",
    "client_id": "vscode-extension"
  },
  "pki": {
    "root_ca_url": "https://platform-api.aegis-platform.tech/api/v1/pki/root-ca"
  }
}
```

## Security Model

The discovery endpoint follows the same security pattern as OIDC `.well-known/openid-configuration`:

- **Public metadata only** — Returns DNS-resolvable endpoint URLs. No secrets, credentials, or internal state.
- **No authentication** — Knowing endpoint URLs does not grant access. Authentication happens through Keycloak OIDC (PKCE + MFA).
- **TLS protected** — Served over HTTPS via the platform's public ingress (Cloudflare, ACM, or similar publicly-trusted certificate).
- **Cache-friendly** — Response includes `Cache-Control: public, max-age=300` (5 minutes).

## Configuration

Set `AEGIS_DISCOVERY_GRPC_ENDPOINT` on the platform-api to define the public gRPC endpoint returned in the discovery response. If not set, the endpoint is derived from the HTTP request's Host header.

```yaml
# Helm values example
platformApi:
  env:
    AEGIS_DISCOVERY_GRPC_ENDPOINT: "platform-api.aegis-platform.tech:8081"
```

The auth authority is derived from the existing `OIDC_ISSUER_URL` env var. The PKI URL is derived from the request's Host header.

## Client Usage (Sovran Extension)

Users configure a single VS Code setting:

```json
{
  "aegisRemote.platform.url": "platform-api.aegis-platform.tech"
}
```

The extension:
1. Fetches `https://{platform.url}/api/v1/discovery`
2. Auto-configures gRPC endpoint, auth authority, and client ID
3. Fetches root CA from the PKI URL in the discovery response
4. Caches discovery results for offline resilience

**Backward compatible:** If `grpcEndpoint` and `auth.authority` are explicitly set in VS Code settings, they take precedence over discovery.
