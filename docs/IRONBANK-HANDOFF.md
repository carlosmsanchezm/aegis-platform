# Iron Bank Container Hardening — Claude Code Handoff

## Context

Aegis Platform has 5 container images that need `hardening_manifest.yaml` files for Iron Bank submission via repo1.dso.mil. All Dockerfiles are already Iron Bank compliant (non-root users, IB base images, proper OCI labels, multi-stage builds). The remaining work is creating the hardening manifests and generating SBOMs.

Owner has CAC access to repo1.dso.mil and can create projects under `dsop/`.

## The 5 Images

### 1. platform-api
- **Dockerfile:** `services/platform-api/Dockerfile`
- **Builder:** `registry1.dso.mil/ironbank/google/golang/ubi9/golang-1.24:1.24.13`
- **Base:** `registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal:9.7`
- **Binary:** Go static binary (`platform-api`)
- **External resources needed in hardening_manifest.yaml:**
  - `pulumi-linux-x64.tar.gz` — Pulumi CLI v3.226.0 from https://get.pulumi.com/releases/sdk/pulumi-v3.226.0-linux-x64.tar.gz
  - `awscli.zip` — AWS CLI v2 from https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip
  - `kubectl` — kubectl binary from https://dl.k8s.io/release/v1.33.0/bin/linux/amd64/kubectl
  - `jq` — jq binary from https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux-amd64
  - `pulumi-plugins-bundle.tar.gz` — optional, Pulumi AWS/EKS plugins
- **Ports:** 8080, 8081
- **Runs as:** UID 1000 (aegis)

### 2. k8s-agent
- **Dockerfile:** `agents/k8s-agent/Dockerfile`
- **Builder:** `registry1.dso.mil/ironbank/google/golang/ubi9/golang-1.24:1.24.13`
- **Base:** `registry1.dso.mil/ironbank/google/distroless/static:nonroot`
- **Binary:** Go static binary (`manager`)
- **External resources:** None (pure Go binary, no external deps)
- **Ports:** None (operator pattern, outbound only)
- **Runs as:** UID 65532 (nonroot)

### 3. proxy
- **Dockerfile:** `services/proxy/Dockerfile`
- **Builder:** `registry1.dso.mil/ironbank/google/golang/ubi9/golang-1.24:1.24.13`
- **Base:** `registry1.dso.mil/ironbank/google/distroless/static:nonroot`
- **Binary:** Go static binary (`aegis-auth-proxy`)
- **External resources:** None
- **Ports:** 8080
- **Runs as:** nonroot:nonroot

### 4. workspace (openssh-vscode)
- **Dockerfile:** `workspace-images/openssh-vscode/Dockerfile`
- **Base:** `registry1.dso.mil/ironbank/opensource/nvidia/cuda:12.6`
- **External resources needed in hardening_manifest.yaml:**
  - `tini-amd64` — Tini init from https://github.com/krallin/tini/releases/download/v0.19.0/tini-amd64
- **Ports:** 11111
- **Runs as:** UID 1000 (aegis)

### 5. vscode-reh-init
- **Dockerfile:** `workspace-images/vscode-reh-init/Dockerfile`
- **Base:** `registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal:9.7`
- **External resources needed in hardening_manifest.yaml:**
  - `vscode-server-linux-x64.tar.gz` — VS Code Server REH binary from https://update.code.visualstudio.com/commit:{COMMIT}/server-linux-x64/stable (pinned commit: `ce099c1ed25d9eb3076c11e4a280f3eb52b4fbeb`)
- **Runs as:** UID 65534

## Task: Create hardening_manifest.yaml for Each Image

Each image needs a `hardening_manifest.yaml` in its directory. The format follows Iron Bank's spec:

```yaml
apiVersion: v1

# Image metadata
name: "aegis-platform-api"  # image name
tags:
  - "latest"
args:
  BASE_REGISTRY: "registry1.dso.mil"
  BASE_IMAGE: "ironbank/redhat/ubi/ubi9-minimal"
  BASE_TAG: "9.7"
  BUILDER_REGISTRY: "registry1.dso.mil"
  BUILDER_IMAGE: "ironbank/google/golang/ubi9/golang-1.24"
  BUILDER_TAG: "1.24.13"

# Labels (must match Dockerfile LABELs)
labels:
  org.opencontainers.image.title: "aegis-platform-api"
  org.opencontainers.image.description: "..."
  org.opencontainers.image.vendor: "Aegis Platform"
  org.opencontainers.image.licenses: "Proprietary"
  mil.dso.ironbank.image.type: "opensource"
  mil.dso.ironbank.product.name: "aegis-platform-api"

# External resources the Iron Bank pipeline must pre-fetch
resources:
  - url: "https://get.pulumi.com/releases/sdk/pulumi-v3.226.0-linux-x64.tar.gz"
    filename: "pulumi-linux-x64.tar.gz"
    validation:
      type: sha256
      value: "<sha256 hash>"
  # ... more resources

# Health check and user
maintainers:
  - name: "Carlos Sanchez"
    email: "carlos@aegis-platform.tech"
    company: "Aegis Platform LLC"
```

## Additional Tasks

1. **Generate SBOMs** for each image using Syft:
   ```bash
   # Build images first, then generate SBOM
   syft <image> -o cyclonedx-json > sbom-<imagename>.json
   ```

2. **Run vulnerability scans** with Grype:
   ```bash
   grype <image> --output json > scan-<imagename>.json
   ```

3. **Verify sha256 hashes** for all external resources listed in the manifests. Download each resource and compute:
   ```bash
   sha256sum <file>
   ```

4. **Create repo1 project structure** — each image gets its own directory on repo1:
   ```
   dsop/aegis-platform/platform-api/
   dsop/aegis-platform/k8s-agent/
   dsop/aegis-platform/proxy/
   dsop/aegis-platform/workspace/
   dsop/aegis-platform/vscode-reh-init/
   ```

   Each directory contains:
   - `Dockerfile`
   - `hardening_manifest.yaml`
   - `README.md` (image description, usage, build instructions)
   - `LICENSE`

## Reference

- Iron Bank Container Hardening Guide: https://repo1.dso.mil/dsop/dccscr/-/tree/master/docs
- Example hardened image (vault-k8s): `repo1.dso.mil/dsop/hashicorp/vault-k8s`
- Aegis build script: `scripts/build-and-push-images.sh`
- Platform One Iron Bank Guide: `docs/business/sales/platform-one-iron-bank-guide.md`
