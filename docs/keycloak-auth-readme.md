# Aegis Identity, Cryptography, and Workspace Access Architecture

This document explains how authentication, authorization, and encrypted transport work across the Aegis local stack. The same design applies in cloud environments; the only difference is which Kubernetes cluster or ingress endpoint you target. Use this as a reference for day‑to‑day development, onboarding new engineers, and supplying architectural detail to compliance reviewers.

---

## 1. Core Components

| Component | Purpose | Key Details |
|-----------|---------|-------------|
| **Keycloak (realm `aegis`)** | Identity Provider (IdP) and MFA enforcement. Issues OIDC tokens to any trusted client. | Clients: `backstage` (confidential), `vscode-extension` (public). Realm stores users, MFA required actions, and client scopes (includes `offline_access`). |
| **Aegis Platform API** | Control plane for workspaces, exposed via gRPC/HTTPS (`platform-api.localtest.me`). | Validates Keycloak tokens, authorizes requests, provisions `AegisWorkload` CRDs, and mints short-lived proxy session tokens. |
| **Workspace Proxy** | Terminates HTTPS from clients and tunnels traffic to workspace pods. | Validates the Platform API session token and opens TCP streams to the pod (e.g., port 11111). |
| **`aegis-connect`** | CLI/extension helper that handles proxy negotiation. | Receives the session token, opens the HTTPS tunnel to the proxy, and pipes SSH traffic through it. |
| **Workspace Pod** | Runs the developer environment (OpenSSH + VS Code Remote Extension Host). | Built from `workspace-images/openssh-vscode`. Listens on SSH port 2222 and VS Code port 11111. |
| **Backstage UI** | Developer portal. Acts as the “workspace client” when launched from the browser. | Uses the same Keycloak realm, maps identities to catalog users, and calls the Platform API on behalf of the logged-in user. |
| **VS Code Extension (`aegis-remote`)** | Native desktop client. Handles Keycloak auth, gRPC calls, and launches VS Code Remote. | Stores `offline_access` refresh tokens to keep sessions alive. |

---

## 2. Authentication and Token Lifecycle

### 2.1 Keycloak Login

1. **Initiation** – Backstage and the VS Code extension both redirect to Keycloak’s `/realms/aegis` endpoints using OIDC Authorization Code Flow with PKCE.
2. **MFA & Credentials** – Users complete MFA (OTP) and password entry. Required actions in Keycloak enforce TOTP enrollment.
3. **Token Issuance** – Keycloak returns:
   - **ID Token** (JWT, RS256) – identity claims for the client.
   - **Access Token** (JWT, RS256) – scopes include `openid profile email offline_access`.
   - **Refresh Token** – long-lived, because the client is allowed to use `offline_access`.
4. **Client Storage** – Backstage stores tokens in its session cookie; the VS Code extension stores them securely in the local keychain.

### 2.2 Using the Access Token with the Platform API

1. Clients call the gRPC endpoint (`platform-api.localtest.me:8443`).
2. The Platform API validates the access token by fetching Keycloak’s JWKS and verifying the RS256 signature.
3. Identity claims (`sub`, `email`, realm roles) are extracted and attached to the request context.
4. Authorization policies (group membership, realm roles, etc.) decide whether the action is allowed.

### 2.3 Session Token for Workspace Proxy

After the Platform API trusts the caller, it mints a short-lived *session token* used only by the workspace proxy.

| Field | Example | Notes |
|-------|---------|-------|
| `sub` | `user:default/cms553` | Backstage catalog identity (namespace/name). |
| `wid` | `w-ee4cc281` | Workspace ID returned by `SubmitWorkload`. |
| `dest` | `aegis-w-w-ee4cc281.aegis-workloads-local.svc.cluster.local:11111` | Pod DNS + port. |
| `aud` | `["aegis-proxy"]` | Ensures only the proxy accepts it. |
| `exp` | `Now + 3 minutes` | Short-lived, single-use. |
| Signature | `HS256` over payload, keyed by a secret shared between Platform API and proxy. | Provides tamper-proof validation without involving Keycloak. |

*Why not signed by Keycloak?* Because the proxy and Platform API already trust the Keycloak-issued access token used to request the session ticket. The session token just carries routing metadata and an expiration; keeping it internal avoids extra network calls to Keycloak and lets the control plane encode implementation-specific fields.

---

## 3. Transport Security Layers

### 3.1 TLS Everywhere

- **Backstage/VS Code → Proxy** – HTTPS (`https://proxy.localtest.me`) with TLS 1.2+. Certificates are generated locally (`aegis-local-trust.pem`) or via a production CA in cloud environments. Satisfies NIST/FedRAMP SC-8/SC-12/SC-13 controls.
- **Backstage/VS Code → Platform API** – Same certificate authority; gRPC with ALPN `h2` over TLS.
- **Proxy → Workspace Pod** – TCP stream inside the cluster (unencrypted). Because this traffic never leaves Kubernetes, it is protected by the cluster network and the outer TLS tunnel.

### 3.2 SSH Inside the Tunnel

The VS Code Remote Extension Host communicates using SSH semantics (the Microsoft server speaks an SSH-like protocol). We run standard OpenSSH in the workspace image, so once the TLS tunnel is established:

1. `ssh` launches `aegis-connect` via `ProxyCommand`.
2. `aegis-connect` sends the session token to the proxy over HTTPS.
3. The proxy validates the token and pipes the TCP stream to port 11111.
4. SSH key exchange and encryption take place over that tunneled connection.

*FIPS Status:* The Debian-based workspace image uses standard OpenSSH/OpenSSL. TLS on the outer tunnel already meets transit encryption requirements; a backlog ticket exists to replace the base image with a FIPS-enabled variant for environments that require FIPS mode end-to-end.

---

## 4. Workspace Provisioning Flow (End-to-End)

1. **Submit Workload**
   - Client calls `SubmitWorkload` with the Keycloak access token and user header `x-aegis-user`.
   - Platform API schedules an `AegisWorkload` CRD; controller creates a job/pod running the workspace image (`aegis-workspace:dev` locally).

2. **Pod Ready Check**
   - Script `scripts/test-workspace-connection.sh` (or Backstage) watches the pod until status `Running` and logs show “Extension host agent listening”.

3. **StartWorkload (optional)**
   - Once the pod is `Ready`, Platform API is notified or the client calls `StartWorkload` to mark the workload as `RUNNING`.

4. **Create Connection Session**
   - Client calls `CreateConnectionSession` → Platform API mints the proxy session token and returns:
     - VS Code deep link (`vscode://aegis.aegis-remote/...`)
     - SSH command snippet (`ssh ... -o ProxyCommand="aegis-connect ..."`).

5. **Connect**
   - Backstage automatically opens the VS Code deep link; the extension invokes `aegis-connect` with the session token.
   - Manual option: run the presented SSH command.

6. **Tunnel Established**
   - TLS between client ↔ proxy, SSH (VS Code server) between client ↔ pod.
   - Session token is invalidated after first use or when the expiry passes.

7. **Cleanup**
   - When finished, delete the workload (`kubectl delete aegisworkload <id>`). Automation may do this on session end.

---

## 5. Token Summary

| Token Type | Issuer | Lifetime | Audience | Purpose |
|------------|--------|----------|----------|---------|
| ID Token | Keycloak | Minutes | Client (Backstage/VS Code) | Display profile info, optional. |
| Access Token | Keycloak | Minutes | Platform API | Authorize control-plane gRPC calls. |
| Refresh Token | Keycloak | Days | Keycloak | Renew access token without relogin. |
| Session Token | Platform API | Minutes, single use | Workspace proxy | Prove workspace access & route to pod. |

---

## 6. Environment Parity (Local vs Cloud)

| Layer | Local (kind/Docker Desktop) | Cloud (Kubernetes / Managed) | Notes |
|-------|------------------------------|------------------------------|-------|
| Identity | Same Keycloak realm/export (Helm chart `charts/aegis-services/files/keycloak/aegis-realm.json`). | Same realm import, possibly hosted Keycloak. | Users/passwords preserved by realm export. |
| Platform API | `platform-api.localtest.me` via port-forward | Public LoadBalancer / ingress (HTTPS) | TLS certificates from local CA vs. ACM/Let’s Encrypt. |
| Proxy | `proxy.localtest.me` via port-forward | Public/Private ingress | Session token semantics identical. |
| Workspace Pods | `aegis-workloads-local` namespace, kind nodes. | Production cluster namespace with cloud storage, node pools. | Images must exist in the target registry; same entrypoint scripts run. |

The architecture, token flow, and cryptographic protections remain identical across environments. Only the endpoints and certificate issuance change.

---

## 7. Operational Checklist

1. **Realm import** – `charts/aegis-services/files/keycloak/aegis-realm.json` must be deployed so Keycloak has the correct clients, scopes, users, and required actions.
2. **Workspace image** – `workspace-images/openssh-vscode` must be built and loaded into the cluster (`kind load` or push to registry). For kind, avoid the `:latest` tag to prevent `imagePullPolicy=Always`.
3. **Port-forward (local dev)** – `make port-forward` for ports 10080/10081/10085.
4. **Smoke test** – `scripts/test-workspace-connection.sh` to validate gRPC, workspace spin-up, proxy token flow, and VS Code server readiness.
5. **Backstage upstream** – Catalog must contain a `User` entity whose email matches the Keycloak account, otherwise the login resolver will fail.

---

## 8. Future Work (Backlog)

| Opportunity | Summary |
|-------------|---------|
| FIPS-capable workspace image | Build on RHEL/UBI FIPS base, enable OS FIPS mode so SSH/VS Code server run entirely in FIPS context. |
| Auto-provision catalog users | Implement Backstage sign-in resolver that creates `User` entities on first Keycloak login (no manual YAML). |
| Token telemetry | Add audit logging for session token issuance/consumption to simplify incident response. |

---

## 9. Useful Commands

```bash
# Build workspace image
make build-workspace

# Load image into kind
docker save aegis-workspace:dev -o /tmp/aegis-workspace-dev.tar
for node in desktop-control-plane desktop-worker desktop-worker2 desktop-worker3; do
  cat /tmp/aegis-workspace-dev.tar | docker exec -i "$node" sh -c 'cat > /tmp/aegis-workspace-dev.tar'
  docker exec "$node" ctr -n=k8s.io images import /tmp/aegis-workspace-dev.tar
  docker exec "$node" rm /tmp/aegis-workspace-dev.tar
done

# Smoke test workspace access
export VSCODE_COMMIT=$(code --version | sed -n '2p' | awk '{print $NF}')
GRPC_TLS=1 \
GRPC_CA=$HOME/aegis-local-trust.pem \
GRPC_TLS_SERVER_NAME=platform-api.localtest.me \
AEGIS_WORKSPACE_USER=cms553@cornell.edu \
WORKSPACE_IMAGE=aegis-workspace:dev \
./scripts/test-workspace-connection.sh
```

---

## 10. Glossary

| Term | Definition |
|------|------------|
| **REH (Remote Extension Host)** | The VS Code server process running inside the workspace container. |
| **Session Token** | Short-lived proxy authorization token issued by the Platform API. |
| **`aegis-connect`** | Command-line tool that wraps the HTTPS tunnel to the proxy and forwards SSH traffic. |
| **`offline_access`** | Keycloak scope/realm role allowing clients to receive refresh tokens that survive browser restarts. |
| **FIPS Mode** | OS configuration that restricts cryptographic modules to those validated under FIPS 140-3. |

---

With this reference you should be able to reason about every step of the login, token issuance, tunnel establishment, and workspace connection flows, regardless of whether you are running locally on kind or in a managed cloud cluster.

---

## Appendix: Authentication & Workspace Access Diagram

```mermaid
sequenceDiagram
    autonumber
    participant User
    participant Browser as Backstage UI
    participant VSCode as VS Code Ext.
    participant Keycloak
    participant PlatformAPI as Platform API (gRPC)
    participant Proxy as Workspace Proxy
    participant Pod as Workspace Pod (SSH/REH)

    rect rgb(235, 245, 255)
        Note over Browser,VSCode: Either client initiates the login flow
        User->>Browser: Open Backstage
        User->>VSCode: Launch extension
    end

    Browser->>Keycloak: OIDC Authorization Code flow (PKCE)
    VSCode->>Keycloak: OIDC Authorization Code flow (PKCE)
    Keycloak-->>Browser: ID + Access + Refresh tokens (incl. offline_access)
    Keycloak-->>VSCode: ID + Access + Refresh tokens (incl. offline_access)

    Browser->>PlatformAPI: SubmitWorkload/GetWorkload (Keycloak access token + x-aegis-user)
    VSCode->>PlatformAPI: SubmitWorkload/GetWorkload (Keycloak access token + x-aegis-user)
    PlatformAPI->>Keycloak: Validate JWT (JWKS, RS256)
    Keycloak-->>PlatformAPI: JWT valid
    PlatformAPI->>PlatformAPI: Authorize request and schedule workspace pod
    PlatformAPI->>Pod: Create AegisWorkload / workspace pod

    PlatformAPI-->>Browser: Workspace status (Ready)
    PlatformAPI-->>VSCode: Workspace status (Ready)

    Browser->>PlatformAPI: CreateConnectionSession
    VSCode->>PlatformAPI: CreateConnectionSession
    PlatformAPI->>PlatformAPI: Mint session token (HS256, short-lived, aud=aegis-proxy)
    PlatformAPI-->>Browser: Session token + VS Code deep link + SSH snippet
    PlatformAPI-->>VSCode: Session token + VS Code URI

    Browser->>VSCode: Forward deep link (vscode://...)

    VSCode->>Proxy: HTTPS (TLS) via aegis-connect + session token
    Proxy->>PlatformAPI: Validate token (shared secret) *(optional)*
    PlatformAPI-->>Proxy: Token valid *(if checked)*
    Proxy->>Pod: Forward TCP stream to port 11111
    VSCode<->>Pod: SSH handshake & VS Code Remote traffic inside tunnel

    Note over Proxy,Pod: Inner SSH runs inside the outer TLS tunnel
    Note over Browser,VSCode: Refresh tokens (offline_access) keep sessions alive
```

---

## Appendix B: Flowchart Overview

```mermaid
flowchart TD
    subgraph Clients
        B[Backstage UI]:::client
        V[VS Code Extension]:::client
    end

    K[Keycloak (OIDC)]:::keycloak
    PAPI[Platform API (gRPC + authz)]:::control
    PROXY[Workspace Proxy (HTTPS)]:::proxy
    POD[Workspace Pod (SSH + VS Code REH)]:::workspace

    B -->|OIDC Auth Code Flow| K
    V -->|OIDC Auth Code Flow| K
    K -->|ID/Access/Refresh Tokens<br/>incl. offline_access| B
    K -->|ID/Access/Refresh Tokens<br/>incl. offline_access| V

    B -->|SubmitWorkload/GetWorkload<br/>Keycloak access token| PAPI
    V -->|SubmitWorkload/GetWorkload<br/>Keycloak access token| PAPI
    PAPI -->|Validate JWT (JWKS)| K
    PAPI -->|Authorize + Schedule workspace| POD

    B -->|CreateConnectionSession| PAPI
    V -->|CreateConnectionSession| PAPI
    PAPI -->|Mint session token (HS256)| B
    PAPI -->|Mint session token (HS256)| V

    B -->|VS Code deep link| V

    V -->|HTTPS tunnel + session token| PROXY
    PROXY -->|Validate token| PAPI
    PROXY -->|Forward TCP to port 11111| POD
    V -->|SSH / VS Code Remote inside tunnel| POD

    classDef client fill:#e1efff,stroke:#4a78d4,stroke-width:1px;
    classDef keycloak fill:#ffe8cc,stroke:#f08a24,stroke-width:1px;
    classDef control fill:#e8f8f3,stroke:#28a17b,stroke-width:1px;
    classDef proxy fill:#f3e8ff,stroke:#8d42f0,stroke-width:1px;
    classDef workspace fill:#fff3e0,stroke:#bf7100,stroke-width:1px;
```
