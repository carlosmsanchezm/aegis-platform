# Aegis Proxy — Architecture Overview

## Purpose

The Aegis Proxy enables secure, authenticated access to GPU development workspaces running on spoke clusters. It handles WebSocket protocol upgrades and routes traffic based on JWT claims.

## Authentication Flow

1. User authenticates with Keycloak and obtains a JWT
2. Client initiates WebSocket connection to proxy with JWT in Authorization header
3. Proxy validates the JWT signature and claims
4. Proxy routes the WebSocket connection to the target workspace pod
5. Bidirectional WebSocket traffic flows through the proxy

## Security

- Runs as non-root (UID 65532)
- FIPS 140-2 compliant (BoringCrypto)
- JWT validation with Keycloak public keys
- No persistent state — stateless proxy
