# Research Prompt: Making the Aegis Workspace Image Iron Bank Compliant

## Your Task

Research and propose a concrete, implementable solution for making the Aegis Platform's GPU workspace container image compliant with the DoD's Iron Bank (registry1.dso.mil) hardening requirements. This image is currently based on Ubuntu and won't pass Iron Bank's pipeline. I need you to find a path that preserves all functionality while meeting Iron Bank's strict requirements.

---

## Background: What Is Aegis Platform?

Aegis is a **self-hosted, multi-cluster GPU workload orchestration platform** for DoD and defense environments. It deploys inside the customer's Kubernetes clusters and schedules GPU workloads (interactive VS Code development workspaces and batch training jobs) across distributed spoke clusters.

**Three products work together:**
1. **Aegis Platform** (Go services) — Hub-and-spoke control plane: gRPC API, Keycloak auth, workload scheduling, budget enforcement
2. **Aegis UI** (Backstage/React) — Web interface for managing workloads, clusters, projects
3. **Sovran** (VS Code Extension) — Connects developers to remote GPU workspaces via WebSocket tunnels

The **workspace image** is the container that runs on GPU nodes in spoke clusters. It's what data scientists and ML engineers actually work inside — it provides an interactive VS Code development environment with GPU access.

## Why Iron Bank Matters

Iron Bank is Platform One's (P1) hardened container image repository for the DoD. Any software running on P1 infrastructure or Big Bang clusters **must** use Iron Bank-approved images. We are targeting DoD customers who run Big Bang or P1 infrastructure. Without Iron Bank images, we cannot deploy to their clusters.

We've already identified a clear path for our 3 control plane images (platform-api, proxy, k8s-agent) — they're small Go binaries that can run on UBI9-minimal or Iron Bank's distroless. The workspace image is the hard one.

---

## The Current Workspace Image

**Dockerfile base:** `nvidia/cuda:12.2.0-runtime-ubuntu22.04`

**What it runs (two services in parallel):**

### 1. OpenSSH Server
- Port: 2222 (configurable via `SSH_PORT`)
- User: `aegis` (UID 1000, GID 1000) with NOPASSWD sudo
- Password auth enabled by default (`aegis:aegis123`, configurable)
- Used for: file sync, git operations, remote terminal, secondary connections

### 2. VS Code Remote Extension Host (REH)
- Port: 11111 (configurable via `VSCODE_SERVER_PORT`)
- **Critical: Downloaded at container startup** from `https://update.code.visualstudio.com`
- The `start-reh.sh` script:
  1. Checks if VS Code server binary exists at `/reh/bin/current`
  2. If not, downloads it from Microsoft's CDN based on `VSCODE_QUALITY` (stable/insider) and `VSCODE_COMMIT`
  3. Starts `code-server` listening on `0.0.0.0:11111`
  4. Uses connection token file at `/reh/token` (content: `hello`)
  5. Disables telemetry
- This runtime download is a **major Iron Bank violation** — Iron Bank requires all content to be present at build time

### Entrypoint flow:
```
tini → entrypoint.sh → starts sshd (background) → exec start-reh.sh (foreground)
```

### Installed packages (Ubuntu):
openssh-server, curl, ca-certificates, sudo, tini, bash, jq, xz-utils, git

### Why CUDA is in the image:
CUDA is a **convenience layer for user GPU workloads**, NOT used by SSH or VS Code server. Users SSH into the container and run their own CUDA-accelerated code (PyTorch, TensorFlow, etc.). The VS Code server and SSH daemon are pure CPU services. Kubernetes exposes GPUs to the pod via `nvidia.com/gpu` resource limits + NVIDIA device plugin.

---

## How Sovran Connects to the Workspace

The Sovran VS Code extension on the developer's machine connects to the workspace through this chain:

```
Developer's VS Code (Sovran Extension)
  → OIDC/PKCE auth with Keycloak
  → gRPC CreateConnectionSession to Platform API
  → Gets JWT proxy ticket (5-min TTL, one-time use)
  → Opens WebSocket Secure (WSS) to spoke-proxy
  → Proxy validates JWT, routes to pod
  → VS Code Remote Authority protocol over tunnel → port 11111 (VS Code server)
  → SSH connections → port 2222 (for terminal, file ops)
```

**What the extension expects from the workspace container:**
1. VS Code Server listening on port 11111 with connection token auth
2. SSH server on port 2222
3. Both services running and healthy
4. VS Code server version compatible with the client's VS Code version
5. Container stays running (long-lived interactive session, not a batch job)

**If either service is missing or broken, the workspace is unusable.** The VS Code server is the primary interface; SSH is secondary but required for full functionality.

---

## Iron Bank Requirements (From Real Repo Analysis)

We've analyzed actual hardening repos on repo1.dso.mil. Here's what the Iron Bank pipeline enforces:

### Base image
Must use an Iron Bank-approved base. Common options:
- `ironbank/redhat/ubi/ubi9-minimal` (most common)
- `ironbank/redhat/ubi/ubi9` (when more packages needed)
- `ironbank/google/distroless/static:nonroot` (for static binaries)

### Build-time only
**All content must be present at build time.** The pipeline:
1. Pre-fetches all resources declared in `hardening_manifest.yaml`
2. Builds in an **offline environment** — no network access during `docker build`
3. Every external file needs a URL + SHA256 hash in the manifest

### No runtime downloads
This is the biggest problem for our workspace image. The VS Code server binary is currently downloaded at container startup. Iron Bank will not allow this.

### Scanning
6 scanners run automatically:
- Anchore/Grype (CVE scanning)
- Trivy (secondary CVE scan)
- OpenSCAP (STIG compliance)
- ClamAV (malware)
- RapidFort (attack surface)
- Twistlock/Prisma Cloud (runtime)

### Non-root
Container must not run as root. (We already run as `aegis` user, but sshd startup may need root briefly.)

### Hardening manifest example (from vault-k8s):
```yaml
apiVersion: v1
name: "hashicorp/vault/vault-k8s"
tags:
  - "v1.7.3"
  - "latest"
args:
  BASE_IMAGE: "redhat/ubi/ubi9-minimal"
  BASE_TAG: "9.7"
labels:
  org.opencontainers.image.title: "vault-k8s"
  org.opencontainers.image.description: "..."
  org.opencontainers.image.licenses: "MPL-2.0"
  org.opencontainers.image.vendor: "Hashicorp"
  mil.dso.ironbank.image.type: "opensource"
resources:
  - filename: vault-k8s.tar.gz
    url: https://github.com/hashicorp/vault-k8s/archive/refs/tags/v1.7.3.tar.gz
    validation:
      type: sha256
      value: 0decebfb3725bbbc053e51b616296b1a3681b64dd5d7fd7b44f7075fc241dfae
maintainers:
  - email: "ironbank@dsop.io"
    name: "Jeff Weatherford"
    username: "jweatherford"
    cht_member: true
```

---

## The 4 Core Challenges You Need to Solve

### Challenge 1: NVIDIA CUDA on UBI9/RHEL

NVIDIA publishes official CUDA Docker images for Ubuntu and CentOS/Rocky, but **not for UBI9 directly**. Iron Bank requires UBI9 or an approved base.

**Questions to research:**
- Does Iron Bank already have NVIDIA CUDA images? Search `registry1.dso.mil` and `repo1.dso.mil` for existing NVIDIA/CUDA hardening repos
- Can CUDA runtime libraries be installed on UBI9 via NVIDIA's RHEL9 RPM repos? (RHEL9 and UBI9 share the same package ecosystem)
- Is there an `nvidia/cuda:12.x-runtime-ubi9` variant anywhere?
- Could we use UBI9 + manually install `cuda-cudart-12-2`, `libcudnn8`, etc. from NVIDIA's RHEL9 repo?
- What's the minimum set of CUDA packages needed for runtime (not development/compilation)?
- Is there a CUDA runtime RPM bundle we could declare in the hardening manifest?

### Challenge 2: VS Code Server Runtime Download

The VS Code Remote Extension Host (REH) binary is currently downloaded at container startup from Microsoft's CDN. This is an Iron Bank violation.

**Questions to research:**
- Can the VS Code server binary be pre-baked into the image at build time?
- Where are the VS Code server releases published? Is there a stable download URL with versioned archives?
- The download URL pattern is: `https://update.code.visualstudio.com/commit:<commit-hash>/server-linux-x64/stable` — can we pin a specific commit and include it in the hardening manifest?
- Does Microsoft publish checksums (SHA256) for VS Code server releases?
- How do we handle version compatibility? The VS Code client and server must be compatible versions. If we pin the server version in the image, how do we keep it in sync with the client?
- Alternative: Could we use `code-server` (coder/code-server, an open-source VS Code fork) instead, which is a single binary and already exists in some Iron Bank repos?
- Alternative: Could we use OpenVSCode Server (gitpod-io/openvscode-server) which is designed for remote deployment?
- What are the licensing implications of pre-packaging Microsoft's VS Code server binary?

### Challenge 3: OpenSSH on UBI9

Ubuntu's `openssh-server` package is well-tested. UBI9 has OpenSSH too, but the configuration differs.

**Questions to research:**
- Does UBI9-minimal include openssh-server, or does it need to be installed?
- Are there any STIG/hardening differences for OpenSSH on RHEL9 vs Ubuntu 22.04?
- Can sshd start as non-root on port 2222? (Ports > 1024 don't require root)
- What's the minimal openssh config needed for our use case (password auth + key auth, non-standard port)?

### Challenge 4: Image Size and Package Footprint

The current Ubuntu+CUDA image is large. Iron Bank prefers minimal images.

**Questions to research:**
- What's the minimum image size achievable with UBI9 + CUDA runtime + openssh + VS Code server?
- Can we use a multi-stage build where CUDA libs are copied from an NVIDIA image into UBI9?
- What packages can be eliminated? (We need: openssh-server, bash, tini, ca-certificates, curl or wget for nothing if we pre-bake everything, git)

---

## What I Need You to Deliver

### 1. Recommended Architecture
A clear recommendation for the Iron Bank-compliant workspace image architecture:
- Which base image to use
- How to get CUDA runtime on it
- How to handle VS Code server (pre-bake vs alternative)
- How to handle SSH

### 2. Draft Dockerfile
A working Dockerfile (or close to working) that follows Iron Bank patterns.

### 3. Draft Hardening Manifest
A `hardening_manifest.yaml` with all required resources declared.

### 4. Risk Assessment
For each approach, what are the risks?
- Version pinning/staleness
- Compatibility issues
- Licensing concerns
- Maintenance burden

### 5. Alternative Approaches
If the straightforward approach (UBI9 + CUDA RPMs + pre-baked VS Code server) doesn't work, what are the alternatives?
- Separate CUDA and non-CUDA workspace images?
- code-server instead of Microsoft VS Code server?
- Init container that installs VS Code server at pod startup (outside the Iron Bank image)?
- Two-image approach: Iron Bank base + sidecar with CUDA?

### 6. Existing Iron Bank Precedent
Are there any existing Iron Bank images that solve similar problems? Specifically:
- Images with NVIDIA/CUDA on RHEL/UBI base
- Images with VS Code server or code-server pre-installed
- Images designed as interactive development environments
- Images that run SSH daemons

---

## Constraints

- The solution must work with Kubernetes on EKS (AWS) with NVIDIA GPU nodes
- The NVIDIA device plugin handles GPU driver injection — the image only needs CUDA runtime userspace libraries
- The solution must be maintainable by a small team (1-2 engineers)
- The workspace image must remain compatible with the Sovran VS Code extension (WebSocket tunnel → port 11111 for VS Code, port 2222 for SSH)
- Iron Bank images are rebuilt/rescanned nightly — the solution must be resilient to scan findings
- We have access to repo1.dso.mil (Platform One GitLab) and registry1.dso.mil (Iron Bank registry) via CAC

## Useful Search Targets on repo1.dso.mil (Public API)

You can query Iron Bank repos without auth:
```bash
# Search for NVIDIA/CUDA images
curl -s "https://repo1.dso.mil/api/v4/projects?search=nvidia&visibility=public&per_page=20"
curl -s "https://repo1.dso.mil/api/v4/projects?search=cuda&visibility=public&per_page=20"

# Search for VS Code / code-server images
curl -s "https://repo1.dso.mil/api/v4/projects?search=code-server&visibility=public&per_page=20"
curl -s "https://repo1.dso.mil/api/v4/projects?search=vscode&visibility=public&per_page=20"

# Search for SSH-based images
curl -s "https://repo1.dso.mil/api/v4/projects?search=openssh&visibility=public&per_page=20"

# Get files from a specific repo
curl -s "https://repo1.dso.mil/api/v4/projects/<url-encoded-path>/repository/tree?ref=development"
curl -s "https://repo1.dso.mil/api/v4/projects/<url-encoded-path>/repository/files/Dockerfile/raw?ref=development"
curl -s "https://repo1.dso.mil/api/v4/projects/<url-encoded-path>/repository/files/hardening_manifest.yaml/raw?ref=development"
```

## Priority

This is a Tier 1 blocker for selling Aegis to DoD customers. The workspace image is the product — it's what users actually interact with. Without an Iron Bank-compliant workspace image, we can demonstrate the control plane but not the actual developer experience. Getting this right is critical.
