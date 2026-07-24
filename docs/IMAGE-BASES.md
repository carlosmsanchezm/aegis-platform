# Image base flavors: public vs Iron Bank

Aegis service Dockerfiles accept **build-args** for builder and runtime bases.  
You choose the flavor at build time — same Dockerfiles, two supply chains.

| Flavor | When to use | Base registries |
|--------|-------------|-----------------|
| **`public`** (default for lab/cloud smoke) | Demo, commercial lab, prove deploy without IB access | `docker.io` (golang), `registry.access.redhat.com` (UBI), `docker.io/nvidia/cuda` (workspace) |
| **`ironbank`** | Gov customer / Platform One / IB-only pull policy | `registry1.dso.mil` Iron Bank images |

Iron Bank is **not** required to run Aegis. It is required when a **customer policy** only allows IB images.

## Quick commands

```bash
# Show which bases will be used
make print-image-flavor
make print-image-flavor IMAGE_FLAVOR=ironbank

# Lab / ECR push (public bases — no registry1.dso.mil)
export AWS_PROFILE=aegis-lab
export AWS_ECR_REGISTRY=471147325433.dkr.ecr.us-east-1.amazonaws.com
make ecr-login
make push-cloud-images                 # IMAGE_FLAVOR=public by default
# or explicitly:
make push-cloud-images-public

# Government / IB pipeline (needs IB registry auth + network)
make push-cloud-images-ironbank
# equivalent:
make push-cloud-images IMAGE_FLAVOR=ironbank
```

## Build-args (all service Dockerfiles)

| Arg | Iron Bank default | Public default |
|-----|-------------------|----------------|
| `BUILDER_REGISTRY` | `registry1.dso.mil` | `docker.io` |
| `BUILDER_IMAGE` | `ironbank/google/golang/ubi9/golang-1.24` | `library/golang` |
| `BUILDER_TAG` | `1.24.13` | `1.24` |
| `BASE_REGISTRY` | `registry1.dso.mil` | `registry.access.redhat.com` |
| `BASE_IMAGE` | `ironbank/redhat/ubi/ubi9-minimal` | `ubi9/ubi-minimal` |
| `BASE_TAG` | `9.7` | `latest` |
| `FIPS_ENABLED` | `true` | `false` (stock golang; set true only if your builder supports BoringCrypto) |

Workspace image (`workspace-images/openssh-vscode`):

| Arg | Iron Bank | Public |
|-----|-----------|--------|
| `BASE_REGISTRY` | `registry1.dso.mil` | `docker.io` |
| `BASE_IMAGE` | `ironbank/opensource/nvidia/cuda` | `nvidia/cuda` |
| `BASE_TAG` | `12.6` | `12.6.0-runtime-ubi9` |

## Override any single base

```bash
make build-proxy IMAGE_FLAVOR=public \
  BASE_REGISTRY=docker.io BASE_IMAGE=library/nginx BASE_TAG=stable-alpine
# (only if that Dockerfile supports a generic base — proxy/agent expect UBI-like or install tools via RUN)
```

Prefer **`IMAGE_FLAVOR`** over one-off overrides unless you know the Dockerfile package manager path.

## Local vs cloud

| Target | Default flavor |
|--------|----------------|
| `build-*-local` / `build-local-all` | **public** |
| `push-cloud-images` | **public** (lab-friendly) |
| IB hardening tree under `ironbank/` | Always IB (submission artifacts) |

## UI image

Backstage UI (`aegis-ui` `Dockerfile.cloud`) is separate and already uses public Node bases. Flavor switch does not change UI.

## Customer messaging

- **Commercial / lab:** “We deploy with publicly available base images.”  
- **DoD IB-only:** “Same product, rebuilt from Iron Bank bases and submitted to Iron Bank / P1.”  
- Do **not** claim runtime is IB-approved unless images are actually approved on Iron Bank.
