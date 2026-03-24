# Workspace Image: Iron Bank Compliance — Architecture Decision Document

**Date:** 2026-03-14
**Status:** Decision Needed
**Author:** Carlos Sanchez + AI research team

---

## 1. The Problem

The Aegis workspace image (`workspace-images/openssh-vscode/Dockerfile`) currently:
- Uses `nvidia/cuda:12.2.0-runtime-ubuntu22.04` as base (Ubuntu — won't pass Iron Bank)
- Downloads Microsoft's VS Code server binary at container startup from `update.code.visualstudio.com` (runtime download — Iron Bank violation)
- Runs OpenSSH server on port 2222 (secondary service)
- Runs VS Code Remote Extension Host (REH) on port 11111 (primary service)

To deploy on Platform One or Big Bang clusters, all images must come from Iron Bank (`registry1.dso.mil`). We need to make this image compliant.

---

## 3. What Needs to Stay the Same (Non-Negotiable)

These are the core architectural constraints for any solution:

1. **The WebSocket tunnel through the spoke proxy must be preserved.** This is Aegis's entire security value: Keycloak OIDC authentication → RBAC authorization → short-lived JWT proxy ticket (5-min, one-time use) → WebSocket Secure tunnel → proxy-side session management → idle timeouts → structured audit logging. Every workspace connection goes through this chain. It cannot be bypassed.

2. **Sovran extension connects via `ManagedResolvedAuthority`** to the VS Code REH binary protocol. This is the working, tested, secure connection flow. It should require zero or minimal changes.

3. **GPU access** must work (NVIDIA CUDA runtime for user ML workloads).

4. **The workspace must be interactive** — long-lived sessions where data scientists write code, train models, and iterate.

---

## 4. The Two Viable Approaches

After researching Iron Bank's repos (repo1.dso.mil public API), analyzing two independent AI research reports, and exploring the Sovran codebase, there are **two viable approaches**. An SSH-based approach was considered and rejected because direct SSH access would bypass Aegis's security chain.

### Approach A: Init-Container (Recommended)

**How it works:**

```
┌─────────────────────────────────────────────────────┐
│ Pod Spec (created by k8s-agent)                      │
│                                                      │
│ initContainers:                                      │
│   - name: vscode-installer                           │
│     image: internal-registry/vscode-reh:pinned-sha   │
│     command: ["cp", "-r", "/reh", "/shared/reh"]     │
│     volumeMounts:                                    │
│       - name: reh-volume                             │
│         mountPath: /shared                           │
│                                                      │
│ containers:                                          │
│   - name: workspace                                  │
│     image: registry1.dso.mil/ironbank/aegis/workspace│
│     volumeMounts:                                    │
│       - name: reh-volume                             │
│         mountPath: /reh                              │
│                                                      │
│ volumes:                                             │
│   - name: reh-volume                                 │
│     emptyDir: {}                                     │
└─────────────────────────────────────────────────────┘
```

**What's in the Iron Bank image (aegis/workspace):**
- Base: `ironbank/opensource/nvidia/cuda:12.6` (already exists in Iron Bank, UBI9-based)
- Packages: bash, git, ca-certificates, procps-ng (from UBI9 repos via dnf)
- Optional: openssh-server (only if needed for secondary file sync — see Section 5)
- Entrypoint: starts VS Code REH from `/reh/bin/current` (mounted from init container)
- NO Microsoft binaries baked into the image
- NO runtime downloads

**What's in the init container (vscode-reh-init):**
- Simple image containing a pre-downloaded, pinned VS Code REH binary
- Copies the binary to a shared emptyDir volume at pod startup
- This image is NOT in Iron Bank — it's in your internal registry (ECR)
- Version pinned to a specific VS Code commit SHA

**Impact on Sovran:** ZERO changes. The extension still connects via WebSocket tunnel to port 11111. The REH binary is the same — it just arrived via init container instead of runtime download.

| Pro | Con |
|-----|-----|
| Iron Bank image is 100% clean | Init container image is not Iron Bank-approved |
| Zero Sovran changes | Two images to maintain instead of one |
| Full Microsoft Marketplace (Copilot, Pylance) | Legal gray area: MS REH in init container (not redistributed publicly, deployed to government infra) |
| Same developer experience as today | Customer's AO must accept non-Iron-Bank init container |
| Fastest to implement (1-2 weeks) | Version drift: must update init container when VS Code updates |
| Version decoupled from Iron Bank image lifecycle | |

---

### Approach B: coder/code-server (Full Iron Bank Native)

**How it works:**
- Replace Microsoft's VS Code REH with coder/code-server (MIT-licensed) in the workspace image
- code-server RPM installed at build time, declared in hardening manifest
- Everything in the image is Iron Bank-approved
- Already has Iron Bank precedent: `dsop/aiml/jupyter/jupyterlab-gpu-codeserver-proxy` (project ID 15566)

**What's in the Iron Bank image:**
- Base: `ironbank/opensource/nvidia/cuda:12.6`
- code-server RPM (pre-fetched by Iron Bank pipeline, SHA256-validated)
- Packages: bash, git, ca-certificates
- Optional: openssh-server
- Entrypoint: starts code-server on port 11111
- 100% Iron Bank compliant, zero external dependencies

**Impact on Sovran:** MAJOR REWRITE REQUIRED. The Sovran extension uses VS Code's `ManagedResolvedAuthority` to connect to Microsoft's REH protocol. coder/code-server uses a different protocol (browser-oriented HTTP + WebSocket). You would need to:
1. Replace the `ManagedResolvedAuthority` resolver with a new connection model
2. Implement code-server's WebSocket protocol in the extension
3. Handle extension marketplace differences (Open VSX vs Microsoft)
4. Test all VS Code Remote functionality works through the new protocol

This is not a minor refactor — it's a fundamentally different connection architecture.

| Pro | Con |
|-----|-----|
| 100% Iron Bank native (single image) | **Major Sovran rewrite** (4-6 weeks minimum) |
| Zero legal risk (MIT license) | **Loses Microsoft Marketplace** (no Copilot, no Pylance, no IntelliCode) |
| Already has Iron Bank precedent | Different developer experience (browser-oriented protocol) |
| Simpler Kubernetes pod spec (no init container) | Uncertain protocol compatibility with VS Code desktop client |
| | Open VSX marketplace is smaller and missing popular extensions |

---

## 5. The SSH Question

**Should the workspace image run OpenSSH?**

The workspace currently runs sshd on port 2222 alongside VS Code REH on port 11111. SSH is used for:
- File synchronization (secondary to VS Code's built-in file system)
- Git operations (also available through VS Code REH)
- Remote terminal (also available through VS Code REH)

**Security analysis:**
- SSH is NOT directly accessible from outside the cluster
- All access goes through the Aegis WebSocket tunnel → spoke proxy → JWT validation → session management
- SSH inside the pod is only reachable via the proxy, which enforces all Aegis security controls
- So SSH does not bypass Aegis's security chain — it's behind it

**However:** VS Code REH already provides file system access, terminal, and git. SSH is redundant for most use cases. Removing it:
- Shrinks the image
- Reduces attack surface (fewer listening services, fewer packages)
- Eliminates an Iron Bank finding source (OpenSCAP SSH STIG checks)
- Simplifies the entrypoint (just start REH, no sshd management)

**Recommendation:** Start without SSH. Add it later only if users demonstrate a concrete need that VS Code REH can't satisfy. The proxy tunnel already routes to port 11111 — if users need raw SSH for some reason, a second proxy route to port 2222 can be added later.

---

## 6. Decision: Init-Container with Both Images in Iron Bank

**Use Approach A (Init-Container).**

Two Iron Bank images:
1. **`aegis/workspace`** — Iron Bank CUDA base + shell tools. The workspace environment.
2. **`aegis/vscode-reh-init`** — Contains pre-downloaded, pinned VS Code REH binary. Injected via init container.

Both images go through Iron Bank hardening. The VS Code REH binary is declared as a resource in the init container's `hardening_manifest.yaml` with SHA256. This is the same pattern Coder Enterprise used with their `envbuilder` image.

**Why this is the right approach:**
1. **Zero Sovran changes** — Extension works exactly as today
2. **Full Microsoft Marketplace** — Copilot, Pylance, IntelliCode preserved
3. **Same developer experience** — Data scientists see no difference
4. **Both images are Iron Bank compliant** — Passes Big Bang's `restrict-image-registries` Kyverno policy for all container types (init, regular, ephemeral)
5. **Proven precedent** — Coder Enterprise's `envbuilder` (Iron Bank project 703) does the same thing
6. **NVIDIA and Istio both hardened their injectors** — This is the expected pattern on P1

**Why code-server (Approach B) was rejected:**
- Incompatible protocol with Sovran's `ManagedResolvedAuthority`
- Loses Microsoft Marketplace (no Copilot, no Pylance)
- Would require major Sovran rewrite
- No clear benefit over init-container approach when both images are in Iron Bank

**Iron Bank's own review process will evaluate the Microsoft licensing** during onboarding of the init container image. If they accept the VS Code REH binary as a declared resource (like they accept other vendor binaries), we're clear. If they reject it on licensing grounds, that's a bridge we cross at onboarding time — not something that should delay starting the workspace base image submission.

---

## 8. Open Questions

1. **Big Bang checks ALL images in a pod — including init containers.** Research confirmed that Big Bang's Kyverno policy `restrict-image-registries` explicitly validates `initContainers`, `containers`, and `ephemeralContainers` against an allowed registry list (default: `registry1.dso.mil/` only). Default enforcement is **Audit** (logged, not blocked), but production environments typically set **Enforce** (blocked). This means the init container image must ALSO be published to Iron Bank for full compliance. Coder Enterprise solved this exact problem — their `envbuilder` image (`dsop/coder-enterprise/coder-enterprise/envbuilder`) is an Iron Bank-published sidecar injector. The init container approach works, but the init container itself must go through Iron Bank hardening too.

2. **Do we need SSH at all?** If VS Code REH provides everything users need (terminal, files, git), we can eliminate SSH entirely. This simplifies the image and avoids SSH STIG findings.

3. **What VS Code commit to pin?** We need to pick a stable VS Code version for the init container. Using `latest` defeats the purpose. Should align with the VS Code version most users have installed.

4. **CUDA version:** Iron Bank has CUDA 12.6 and 13.0.0. Which version do our target users need? Start with 12.6 (broader compatibility) or 13.0.0 (newer)?
