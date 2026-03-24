# Iron Bank Getting Started Form — Answers

**Date:** 2026-03-23
**Form URL:** https://docs.google.com/forms/d/e/1FAIpQLSeWkruUNDFoVaBKzpa1MtS6YX8WFVTQn93jhGYYRWKQu6qI5g/viewform

---

## Qualifying Questions

**I have attended an onboarding session or have reviewed the virtual materials.**
> **Yes**
> (Review the virtual materials at https://docs-ironbank.dso.mil/quickstart/contributor-onboarding/ before submitting if you haven't already)

**Are you okay with your containers being publicly available?**
> **Yes**
> (Aegis containers will be on IL2. No classified components — this is a commercial Kubernetes control plane product.)

**Are you okay if your security findings are public?**
> **Yes**
> (Findings and scans will be visible in VAT and the Iron Bank registry. We have no issue with this.)

**If your application requires a license, do your containers implement a license model to prevent unauthorized usage?**
> **Yes**
> (Aegis Platform uses commercial licensing. The platform-api enforces license validation at runtime — the containers themselves are usable but require a valid Aegis Platform license key to operate in production. Iron Bank does not need to manage licensing.)

**Do you agree to submit your application upon every update, the day of the update?**
> **Yes**

**Do you agree to provide timely justifications on any findings discovered during our hardening and scanning process?**
> **Yes**

**Do you agree to work towards correcting any findings that may be identified as unacceptable based on the 'Findings Approvers' review process?**
> **Yes**

**Is your application currently containerized?**
> **Yes**
> (All 6 images are currently containerized and running in production on EKS.)

**Does your application run on Linux containers?**
> **Yes**
> (All containers are Linux-based — UBI9-minimal and distroless/static bases.)

**Is your application running on or able to be rebased to another Iron Bank Image?**
> **Yes**
> (All containers already use Iron Bank base images: ironbank/redhat/ubi/ubi9-minimal, ironbank/google/golang/ubi9/golang-1.24, ironbank/google/distroless/static, ironbank/opensource/nvidia/cuda.)

**Can your application build and run in an offline/air-gapped environment?**
> **Yes**
> (All external dependencies — Pulumi CLI, AWS CLI, kubectl, jq, tini, VS Code Server — are declared in hardening_manifest.yaml with sha256 hashes for pipeline pre-fetch. Source code is packaged as a release tarball resource. No runtime downloads required.)

**Is your container FIPS compliant?**
> **Yes**
> (Go services — platform-api, proxy, k8s-agent — are compiled with GOEXPERIMENT=boringcrypto and CGO_ENABLED=1 for FIPS 140-2 validated cryptography. Runtime startup checks verify BoringCrypto is linked. The workspace and vscode-reh-init containers do not handle encryption directly — they rely on the platform's mTLS layer — but they can run on FIPS-enabled nodes.)

**Can your application include DNF commands to ensure ALL the latest updates are applied?**
> **Yes**
> (UBI9-based containers use `microdnf update -y` or `dnf upgrade -y`. Two containers — k8s-agent and proxy — use UBI9-minimal in their final stage with microdnf. The workspace image uses full dnf. Note: the original Dockerfiles targeted distroless/static for k8s-agent and proxy, but we have switched to UBI9-minimal for Iron Bank compatibility so DNF updates can run.)
>
> If you prefer to keep distroless for k8s-agent and proxy, answer: **"My container uses a Distroless image"** for those two, but note this may create friction with the pipeline. The current Dockerfiles already use UBI9-minimal.

**Have you checked that any containerized dependencies are already hardened within Iron Bank?**
> **Yes**
> (Verified: ubi9-minimal, golang-1.24 builder, distroless/static, nvidia/cuda 12.6 all exist in Iron Bank registry. No additional containerized dependencies are required.)

---

## Open-Ended Questions

**Are there any end of life dependencies your application requires?**
> No. All dependencies are current: Go 1.24, Node.js 22, CUDA 12.6, UBI9, Pulumi v3.226.0, AWS CLI v2, kubectl v1.33.0, jq v1.8.0, tini v0.19.0. No EOL components.

**Do any of your containers require 'root' user to start? If so, which ones and why?**
> No. All containers run as non-root in their final stage:
> - platform-api: UID 1000 (aegis)
> - k8s-agent: UID 65532 (nonroot)
> - proxy: UID 65532 (nonroot)
> - workspace: UID 1000 (aegis)
> - vscode-reh-init: UID 65534 (nobody)
> - ui: UID 1001 (node)
>
> Build stages use root for package installation, but final runtime stages are all non-root.

**Do any files within your containers include special permissions such as SUID or SGID? If so, which ones and why?**
> No. The workspace image explicitly strips all SUID/SGID bits: `find / -xdev -perm /6000 -type f -exec chmod a-s {} \;`. The Go binary containers (platform-api, k8s-agent, proxy) contain only the static binary and system packages — no SUID/SGID files. The vscode-reh-init container is a minimal UBI9 image with no special permissions.

**Are there any import/export controls regarding your application? If yes, please list which containers and which controls.**
> No. Aegis Platform uses only standard Go cryptography (BoringCrypto/FIPS via Go standard library) and gRPC TLS. No EAR/ITAR-controlled encryption libraries, no custom cryptographic implementations, no export-restricted algorithms. All crypto is provided by NIST-validated BoringSSL through Go's native boringcrypto experiment.

---

## Contact Information

**Name of company (Vendor) or program office:**
> Aegis Platform LLC

**Name of primary company POC:**
> Carlos Sanchez

**Email of primary company POC:**
> carlos@aegis-platform.tech

**Phone number of primary company POC:**
> [YOUR PHONE NUMBER]

**Name of primary engineer for the container:**
> Carlos Sanchez

**Email of primary engineer for the container:**
> carlos@aegis-platform.tech

**Phone number of primary engineer for the container:**
> [YOUR PHONE NUMBER]

**What are the Repo1 usernames for ALL of your engineers working the container?**
> cms553

**Name of government sponsor working with Platform1:**
> N/A
> (We are onboarding as a commercial vendor without a current government sponsor. We are pursuing DIU and DoD contracts and seeking ATO through Iron Bank hardening as part of that effort.)

**Email of government sponsor working with Platform1:**
> N/A

---

## Container Details

**What containers need to be added?**
> 6 containers:
>
> 1. **aegis-platform-api** (Application) v1.0.0 — gRPC+REST control plane API server. Base: UBI9-minimal 9.7. Builder: IB golang-1.24.
> 2. **aegis-k8s-agent** (Application) v1.0.0 — Kubernetes operator agent for spoke clusters. Base: UBI9-minimal 9.7. Builder: IB golang-1.24.
> 3. **aegis-proxy** (Application) v1.0.0 — Reverse proxy for secure WebSocket workspace tunneling. Base: UBI9-minimal 9.7. Builder: IB golang-1.24.
> 4. **aegis-workspace** (Application) v1.0.0 — GPU development workspace with CUDA runtime. Base: IB nvidia/cuda 12.6.
> 5. **aegis-vscode-reh-init** (Init Container) v1.0.0 — Init container injecting VS Code Remote Extension Host. Base: UBI9-minimal 9.7.
> 6. **aegis-ui** (Application) v1.0.0 — Backstage-based developer portal. Base: IB Node.js 22 (UBI9).

**What would you like your repo naming convention to be?**
> aegis-platform/platform-api
> aegis-platform/k8s-agent
> aegis-platform/proxy
> aegis-platform/workspace
> aegis-platform/vscode-reh-init
> aegis-platform/ui

---

## Important Notes

1. **Government Sponsor:** The form asks for a .gov email. If you don't have one yet, "N/A" is acceptable for initial onboarding, but you may need one for final ATO approval. If you have a DIU or DoD contact who can sponsor, use their info instead.

2. **Phone Number:** Fill in your actual phone number in the two placeholder spots.

3. **Licensing Answer:** If Aegis doesn't have runtime license enforcement yet, change the licensing answer to: "No — we are implementing a licensing model and will have it in place before container publication." Be honest here — they're asking whether someone can just pull your container and use it without paying.

4. **Review Virtual Materials:** Before checking "Yes" on the first question, at minimum skim the contributor onboarding docs at https://docs-ironbank.dso.mil/quickstart/contributor-onboarding/ so you can honestly say you've reviewed them.
