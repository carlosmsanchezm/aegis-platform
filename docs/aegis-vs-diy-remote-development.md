# Aegis vs. DIY Remote Development: Why Not Just SSH?

## The Question

> "I can configure VS Code Remote-SSH through a bastion host, tunnel into a VM in my GovCloud VPC, and get a thin client on my desktop looking at a VS Code server running on the remote machine. What's the fundamental difference between that and what Aegis does?"

This is a fair question. The answer is not that the connection mechanism is radically different — it's that **everything around the connection** is different.

---

## The Short Answer

VS Code Remote-SSH and Sovran (the Aegis VS Code extension) both use VS Code's remote development capability to run a Remote Extension Host on a remote machine and stream editor content to your local VS Code. The transport is architecturally similar.

**The difference is not how bits reach the editor. The difference is:**

1. What secures the session (single-use tokens vs. long-lived SSH sessions)
2. What protects data on the endpoint (RAM-disk + ephemeral tokens vs. nothing)
3. What the attacker gets if the connection is compromised (one isolated pod vs. a bastion with VPC-wide access)
4. What the platform team has to build and maintain (Helm chart vs. months of infrastructure)
5. What the data scientist experiences (click "Connect" vs. configure SSH keys and tunnels)

---

## Full Comparison

| Dimension | DIY: VS Code Remote-SSH + Bastion | Aegis + Sovran |
|---|---|---|
| **Setup per session** | Configure SSH keys, bastion host, tunnel, ProxyJump config. 10-15 min first time, 2-5 min after. | Click "Connect" in VS Code. 30 seconds. |
| **Authentication** | SSH keys (static, no expiry unless manually rotated). Key compromise = persistent access. | OIDC PKCE via Keycloak with MFA. Single-use JWT token, 5-minute TTL, JTI tracking. Token works exactly once, then it's dead. |
| **Session lifetime** | SSH session stays open as long as the connection lives. Hours, sometimes days. If the session is hijacked, the attacker has it until disconnect. | 5-minute token, auto-renewed by the extension. If intercepted, it's already expired or already consumed. Idle timeout (configurable, default 15 min) auto-disconnects inactive sessions. |
| **Data on endpoint after session** | VS Code caches files in `~/.vscode/`, temp directories, editor buffers on disk. CUI data persists on your laptop after you disconnect. | Secure mode: all VS Code runtime data on encrypted RAM-disk. `hotExit: off` (no disk caching of unsaved files). Clipboard wiped on disconnect. RAM-disk destroyed on exit. Nothing persists. |
| **Blast radius if compromised** | Bastion host has network access to the entire private VPC. Attacker can pivot to any service reachable from the bastion — databases, other VMs, control plane. | Workspace pod runs in an isolated Kubernetes namespace with NetworkPolicy. Attacker sees one pod. Cannot reach other projects, other pods, or other services. Non-root container, read-only root filesystem, dropped capabilities. |
| **Multi-tenancy** | None. Anyone with SSH access to the bastion can reach any VM on the VPC. Isolation is manual (separate VMs per team, separate security groups). | Per-project namespace isolation. RBAC per user. Budget enforcement per project. Project A cannot see Project B's workloads, data, or credentials. |
| **Audit trail** | SSH connection logs: IP address, timestamp, username. No application-level audit of what happened inside the session. | Structured JSON audit events: who connected, to which workspace, from which IP, when, every action logged. AU-3 compliant fields. SIEM-exportable. |
| **GPU scheduling** | Manual. You SSH to a VM that hopefully has a GPU free. If the GPU is in use, you wait, or you ask someone to provision another VM. | User picks a GPU flavor (A100, H100, etc.), clicks "Launch Workspace." Aegis places it on a cluster with available capacity. Kueue handles fair queuing. Budget enforcement prevents overspend. |
| **Operational overhead** | Platform team builds and maintains: bastion host (patching, hardening, MFA config), SSH key rotation policies, security groups, VM images with GPU drivers, user accounts, monitoring dashboards, compliance documentation. | Platform team installs one Helm chart. Aegis handles auth (Keycloak), scheduling (Kueue), isolation (namespaces + NetworkPolicy), audit logging, session management, and budget enforcement. |
| **Self-service** | Developer submits a ticket to get bastion access, VM provisioned, SSH key added, GPU allocated. Days to weeks. | Developer clicks "Launch Workspace" in the Backstage UI. Picks a GPU size. 90 seconds to a working environment. |
| **Compliance documentation** | You build it yourself. NIST control mapping, session management procedures, audit configuration, key rotation policies — all custom documentation per environment. | Built into the platform. Control mappings exist. Audit logging is in the code path. Session token lifecycle is enforced, not documented-and-hoped-for. |

---

## "What About Data Leakage?"

This is the sharpest question. VS Code Remote-SSH caches files locally — temp directories, `.vscode` folder, editor state. CUI data touches the local disk, even briefly. Auditors flag this under NIST 800-171 media protection controls.

**What Sovran secure mode does:**

| Protection | How it works |
|---|---|
| **RAM-disk sandbox** | All VS Code runtime data (user-data, extensions, cache, temp files) lives on an encrypted RAM-disk. Environment variables `TMPDIR`, `XDG_CACHE_HOME`, `XDG_CONFIG_HOME`, `XDG_DATA_HOME` all point to RAM-disk paths. |
| **Hot exit disabled** | VS Code normally caches unsaved file content to disk for crash recovery. Secure mode sets `files.hotExit: "off"`. Editor buffers exist only in process memory, never written to persistent storage. |
| **Ephemeral tokens** | Auth tokens stored in-memory only. No refresh tokens requested or stored. All secrets wiped on session end. |
| **Full disk encryption check** | Launcher verifies FileVault (macOS) or LUKS (Linux) is enabled before creating the RAM-disk. Even if data touches persistent storage, it's encrypted. |
| **Clipboard wipe** | On disconnect, clipboard is cleared (`vscode.env.clipboard.writeText('')`). CUI text copied during the session does not persist. |
| **Session cleanup** | On exit: all secrets wiped from VS Code SecretStorage, RAM-disk unmounted and destroyed. Terminal prints "No CUI persists on disk." |

**What Sovran secure mode does NOT do:**

| Gap | Why | Mitigation |
|---|---|---|
| Prevent clipboard copy during session | VS Code has no API to intercept OS clipboard writes | Requires MDM policy enforcement at OS level |
| Prevent screenshots during session | No VS Code API to block OS screenshots | Requires MDM / DLP software |
| Prevent process memory dumps | OS-level attack; editor buffers are in process RAM during session | FDE covers swap; RAM-disk covers VS Code cache |

**The honest comparison:** During an active session, both VS Code Remote-SSH and Sovran have file content in process memory — that's how any editor works. The difference is what happens *after* the session: with Remote-SSH, data lingers on disk. With Sovran secure mode, everything is destroyed.

---

## "What About Code-Server / Browser-Based VS Code?"

Browser-based VS Code (code-server, Gitpod, GitHub Codespaces) is a fundamentally different approach:

| | VS Code Remote-SSH | Sovran (Aegis) | Code-Server (Browser) |
|---|---|---|---|
| **Client** | Native VS Code | Native VS Code | Browser tab |
| **File handling** | Files synced to local machine | Files streamed to local VS Code (RAM-disk in secure mode) | Files stay server-side; browser sees rendered pixels |
| **Data leakage risk** | High (local disk cache) | Low (RAM-disk, ephemeral) | Lowest (pixels only) |
| **Performance** | Native IDE speed | Native IDE speed | Browser latency, no native keybindings, limited extension support |
| **Extension ecosystem** | Full VS Code marketplace | Full VS Code marketplace | Limited (Open VSX only, no Copilot, no Pylance) |
| **Debugging** | Full native debugger | Full native debugger | Limited browser debugging |
| **GPU workflows** | Works (if VM has GPU) | Works (pod has GPU) | Poor (browser can't handle GPU-heavy UIs) |

**Why Sovran chose native VS Code over browser:** Browser-based VS Code solves the data leakage problem completely (pixels only), but the developer experience is significantly worse — latency, limited extensions, no native debugging. For ML engineers who need GPU compute with full IDE capability (debugging PyTorch training loops, profiling CUDA kernels), browser-based VS Code is not viable.

Sovran's approach: use native VS Code for the IDE experience, and address data leakage through secure mode (RAM-disk + ephemeral tokens + session cleanup) rather than through pixels. The tradeoff is explicit: slightly more data exposure during the session (mitigated by RAM-disk), dramatically better developer experience.

---

## "Is SSH Itself the Problem?"

No. SSH with FIPS-validated OpenSSL is cryptographically sound. The encryption, key exchange, and integrity checks are fine for CUI in transit (NIST 800-53 SC-8, SC-13).

**The problem is everything SSH doesn't give you:**

| What SSH provides | What SSH does NOT provide |
|---|---|
| Encrypted transport | Session token lifecycle management |
| Server authentication (host keys) | Application-level RBAC per project |
| Client authentication (keys or password) | Multi-tenant isolation |
| | Budget enforcement per team |
| | Self-service workspace provisioning |
| | GPU scheduling and placement |
| | Structured audit logging (who did what, not just who connected) |
| | Session idle timeout with auto-disconnect |
| | Endpoint data protection (RAM-disk, ephemeral tokens) |
| | Compliance documentation (NIST control mappings) |
| | One-click setup (vs. SSH key management, bastion config, tunnel setup) |

SSH is a transport protocol. It moves bytes securely. Aegis is a platform that manages the entire lifecycle: who can access which GPU, for how long, at what cost, with what audit trail, with what data protection on the endpoint.

**Analogy:** Asking "why not just use SSH?" is like asking "why use Kubernetes when you can just run Docker?" Docker runs containers. Kubernetes manages the lifecycle — scheduling, scaling, health checks, networking, RBAC. SSH moves bits. Aegis manages the GPU workspace lifecycle.

---

## Compliance Controls Aegis Adds That SSH Alone Doesn't Cover

| NIST 800-53 Control | What it requires | SSH alone | Aegis |
|---|---|---|---|
| **AC-2** Account Management | Automated account lifecycle, role-based access | Manual (create SSH users, manage keys) | Keycloak OIDC with automated provisioning, role-based project access |
| **AC-11** Session Lock | Lock session after inactivity period | No built-in idle timeout | Configurable inactivity timeout (default 15 min), auto-disconnect in secure mode |
| **AC-12** Session Termination | Terminate session after defined conditions | SSH `ClientAliveInterval` (network-level only) | Single-use tokens (5-min TTL), JTI tracking prevents replay, forced disconnect on timeout |
| **AU-2/AU-3** Audit Events | Log security-relevant events with required fields | Connection logs only (IP, time, user) | Structured JSON: subject, action, resource, outcome, source IP, timestamp — every operation |
| **SC-8** Transmission Confidentiality | Encrypt CUI in transit | Yes (SSH encryption) | Yes (TLS 1.2+ on WebSocket, mTLS between services) |
| **SC-13** Cryptographic Protection | Use FIPS-validated crypto modules | Requires FIPS OpenSSL + hardened SSH config | Same requirement (BoringCrypto planned), plus session tokens use HS256/RS256 |
| **MP-7** Media Protection | Protect CUI on portable/removable media | No endpoint protection | Secure mode: RAM-disk, FDE verification, clipboard wipe, hotExit off |

---

## The One-Liner (For Your Next Interview)

> "SSH is a transport protocol — it moves bytes securely. Aegis is a GPU workspace platform that manages the entire lifecycle: who accesses which GPU, for how long, at what cost, with what audit trail, with what data protection on the endpoint, and with zero SSH key management. The connection mechanism is similar to VS Code Remote-SSH — the difference is that with a bastion, your platform team spends months building everything around that SSH tunnel. Aegis replaces all of it with a Helm chart."
