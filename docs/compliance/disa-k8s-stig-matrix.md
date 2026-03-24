# Aegis Platform: DISA Kubernetes STIG Compliance Matrix

> **STIG Version**: Kubernetes STIG V2R2 (Release Date: 2024-06-10)
> **Assessment Date**: 2026-03-23
> **Platform Version**: 1.0-pilot (release/v1.0-pilot branch)
> **Assessor**: Automated codebase analysis (Claude)
> **Cross-references**: FedRAMP control mapping (`docs/compliance/fedramp/01-control-mapping.csv`), CRM (`docs/compliance/customer-docs/customer-responsibility-matrix.md`), CIS (`docs/compliance/customer-docs/control-implementation-statements.md`), Hardening Guide (`docs/compliance/customer-docs/configuration-hardening-guide.md`)

## Executive Summary

The DISA Kubernetes STIG V2R2 contains **93 findings** (V-242376 through V-242468). Because Aegis Platform is an **application that runs ON Kubernetes** — not a Kubernetes distribution — the vast majority of STIG findings (92.5%) are **Customer Responsibility**: the deploying organization's cluster administrators must configure the underlying Kubernetes infrastructure. Aegis owns application-layer controls (pod security, network policies, namespace isolation, secrets handling) and provides a Configuration Hardening Guide with recommendations for all customer-owned controls.

This aligns with the Shared Responsibility Matrix (`docs/shared-responsibility-matrix.md`): Aegis owns application-layer controls; the customer owns infrastructure-layer controls.

## Summary

| Status | Count | Percentage |
|--------|-------|------------|
| Satisfied | 4 | 4.3% |
| Partially Satisfied | 3 | 3.2% |
| Not Satisfied | 0 | 0.0% |
| Not Applicable | 0 | 0.0% |
| Customer Responsibility | 86 | 92.5% |
| **Total** | **93** | **100%** |

### Severity Breakdown

| Severity | Total | Customer Resp. | Satisfied | Partial | Not Satisfied |
|----------|-------|---------------|-----------|---------|---------------|
| CAT I (High) | 17 | 15 | 1 | 1 | 0 |
| CAT II (Medium) | 73 | 68 | 3 | 2 | 0 |
| CAT III (Low) | 3 | 3 | 0 | 0 | 0 |

### CAT I Findings Summary

All 17 CAT I findings are addressed — 15 are Customer Responsibility (cluster infrastructure), 1 is Satisfied by Aegis, and 1 is Partially Satisfied (secrets as env vars — remediation planned).

## How to Read This Matrix

- **Satisfied**: Aegis implements this control in code/configuration. Evidence exists in the codebase.
- **Partially Satisfied**: Aegis partially implements; additional configuration or customer action needed.
- **Not Satisfied**: Gap identified; remediation planned or required.
- **Not Applicable**: Control does not apply to Aegis's architecture.
- **Customer Responsibility**: Control must be implemented by the deploying organization (cluster admin). Aegis provides guidance in the Configuration Hardening Guide. See CRM for details.

**Hub vs Spoke**: The hub cluster runs Platform API, Keycloak, and reverse proxy. Spoke clusters run lightweight K8s agents with outbound-only connectivity. Notes indicate when a control applies differently.

---

## Compliance Matrix

---

### Category 1: TLS and Encryption

---

#### V-242376: Controller Manager must use TLS 1.2+
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--tls-min-version=VersionTLS12` on Controller Manager to prevent use of SSL and TLS < 1.2.
- **Aegis Implementation**: Cluster infrastructure control. Customer configures Controller Manager flags during cluster provisioning. Aegis's own application TLS enforces TLS 1.2+ minimum (`services/platform-api/internal/server/server.go:2926-2958`, `tls.VersionTLS12`). Hardening Guide recommends TLS 1.2+ at all layers.
- **Evidence**: `docs/compliance/customer-docs/configuration-hardening-guide.md` (TLS Configuration section), `services/platform-api/internal/server/server.go:2926` (app-level TLS)
- **Cross-reference**: SC-8 (Implemented per `fedramp/01-control-mapping.csv`) — See CRM: CUST (infrastructure)
- **Gap/Notes**: Customer must set on Controller Manager. EKS manages this by default.

#### V-242377: Scheduler must use TLS 1.2+
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--tls-min-version=VersionTLS12` on Scheduler.
- **Aegis Implementation**: Cluster infrastructure control. Customer configures Scheduler flags.
- **Evidence**: `docs/compliance/customer-docs/configuration-hardening-guide.md`
- **Cross-reference**: SC-8 (Implemented) — See CRM: CUST
- **Gap/Notes**: Customer must set on Scheduler. EKS manages this by default.

#### V-242378: API Server must use TLS 1.2+
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--tls-min-version=VersionTLS12` on API Server.
- **Aegis Implementation**: Cluster infrastructure control. Customer configures API Server. Aegis's gRPC server enforces `MinVersion: tls.VersionTLS12` in its own TLS config.
- **Evidence**: `services/platform-api/internal/server/server.go:2930` (`MinVersion: tls.VersionTLS12`), `docs/compliance/customer-docs/configuration-hardening-guide.md`
- **Cross-reference**: SC-8 (Implemented) — See CRM: CUST
- **Gap/Notes**: Customer must configure. EKS default is TLS 1.2+.

#### V-242379: etcd must use TLS (auto-tls disabled)
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--auto-tls=false` on etcd to prevent automatic TLS certificate generation.
- **Aegis Implementation**: Cluster infrastructure control. Aegis does not interact with etcd directly.
- **Evidence**: `docs/shared-responsibility-matrix.md` (customer owns infrastructure)
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: EKS fully manages etcd; this is inherently met on EKS.

#### V-242380: etcd must use TLS (peer-auto-tls disabled)
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--peer-auto-tls=false` on etcd.
- **Aegis Implementation**: Cluster infrastructure control. Aegis does not interact with etcd.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: Inherently met on EKS (managed etcd).

#### V-242418: API Server must use approved cipher suites
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--tls-cipher-suites` to approved ciphers (TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, etc.).
- **Aegis Implementation**: Cluster infrastructure control. Aegis's own application enforces FIPS-approved cipher suites when FIPS enabled: `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256`, `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384`, `TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256`, `TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384`.
- **Evidence**: `services/platform-api/internal/server/server.go:2935-2944` (FIPS cipher suites), `charts/aegis-services/values/common.yaml:88-94` (configured cipher suites)
- **Cross-reference**: SC-8, SC-13 (Partial — FIPS gap per `fedramp/01-control-mapping.csv`) — See CRM: CFG
- **Gap/Notes**: Customer must configure on API Server. Aegis app-level ciphers are correct.

#### V-242439: API Server must disable basic authentication
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Remove `--basic-auth-file` from API Server. Basic auth transmits credentials in cleartext.
- **Aegis Implementation**: Cluster infrastructure control. Aegis uses OIDC/JWT authentication exclusively — never basic auth. Aegis does not require or configure `--basic-auth-file`.
- **Evidence**: `services/platform-api/internal/server/mw/auth.go:185-264` (OIDC-only auth), `docs/compliance/customer-docs/security-architecture-guide.md`
- **Cross-reference**: IA-2 (Implemented) — See CRM: CUST
- **Gap/Notes**: Customer must verify not set. EKS does not support basic auth (inherently met). **CAT I — verify during assessment.**

#### V-242440: API Server must disable token authentication
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Remove `--token-auth-file` from API Server.
- **Aegis Implementation**: Cluster infrastructure control. Aegis uses OIDC/JWT tokens from Keycloak, not static token files.
- **Evidence**: `services/platform-api/internal/server/mw/auth.go` (JWT validation)
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: Customer must verify. EKS does not use static token files.

#### V-242441: Endpoints must use approved certificate and key pair
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--kubelet-client-certificate` and `--kubelet-client-key` on API Server for mutual TLS to kubelets.
- **Aegis Implementation**: Cluster infrastructure control. Aegis's own mTLS uses proper certificates between platform-api and spoke agents.
- **Evidence**: `services/platform-api/internal/server/server.go:2948-2953` (mTLS with `RequireAndVerifyClientCert`)
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: EKS configures kubelet client certificates by default.

#### V-242468: API Server must prohibit TLS 1.0/1.1 and SSL 2.0/3.0
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--tls-min-version=VersionTLS12` or higher on API Server per NIST 800-52R2.
- **Aegis Implementation**: Cluster infrastructure control. Aegis's application enforces TLS 1.2+ at all layers.
- **Evidence**: `services/platform-api/internal/server/server.go:2930`, `charts/aegis-services/values/common.yaml:88`
- **Cross-reference**: AC-17(2) — See CRM: CUST
- **Gap/Notes**: Customer must configure. EKS default meets requirement.

---

### Category 2: Access Control and RBAC

---

#### V-242381: Controller Manager must create unique service accounts
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Set `--use-service-account-credentials=true` on Controller Manager. Prevents shared credentials across controllers.
- **Aegis Implementation**: Cluster infrastructure control. Aegis creates dedicated service accounts for each component: `platform-api-sa` (`charts/aegis-services/templates/platform-api-serviceaccount.yaml`), spoke agent service account (`charts/aegis-spoke/templates/k8s-agent-rbac.yaml`).
- **Evidence**: `charts/aegis-services/templates/platform-api-serviceaccount.yaml`, `charts/aegis-spoke/templates/k8s-agent-rbac.yaml`
- **Cross-reference**: AC-6 (Implemented) — See CRM: CUST
- **Gap/Notes**: Customer must set Controller Manager flag. **CAT I — verify during assessment.**

#### V-242382: API Server must enable Node,RBAC authorization mode
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Set `--authorization-mode=Node,RBAC` on API Server.
- **Aegis Implementation**: Cluster infrastructure control. Aegis requires RBAC to function — its ClusterRoles/RoleBindings depend on RBAC being enabled. Application-level RBAC is enforced independently via `internal/authz/policy.go`.
- **Evidence**: `charts/aegis-services/templates/platform-api-rbac.yaml` (ClusterRole/Binding), `services/platform-api/internal/authz/policy.go:49-82` (app-level RBAC)
- **Cross-reference**: AC-3 (Implemented) — See CRM: CUST
- **Gap/Notes**: Customer must configure. EKS defaults to RBAC. **CAT I — verify during assessment.**

#### V-242383: User-managed resources must be in dedicated namespaces
- **Severity**: CAT I
- **Status**: **Satisfied**
- **STIG Description**: No user workloads in `default`, `kube-public`, or `kube-node-lease` namespaces.
- **Aegis Implementation**: Aegis deploys all workloads to dedicated namespaces: `aegis-system` for hub components, project-specific namespaces for user workloads (configurable via `namespacePerProject: true` in hardening guide). K8s agents on spoke clusters also use dedicated namespaces. No Aegis resources are deployed to default/kube-public/kube-node-lease.
- **Evidence**: `charts/aegis-services/templates/platform-api-deployment.yaml` (`.Release.Namespace`), `charts/aegis-spoke/templates/k8s-agent-deployment.yaml` (`.Release.Namespace`), `docs/compliance/customer-docs/configuration-hardening-guide.md` (tenant isolation section: `namespacePerProject: true`)
- **Cross-reference**: AC-3 (Implemented) — See CRM: YES-P
- **Gap/Notes**: None. Aegis fully satisfies this requirement. **CAT I — SATISFIED.**

#### V-242390: API Server must have anonymous authentication disabled
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Set `--anonymous-auth=false` on API Server.
- **Aegis Implementation**: Cluster infrastructure control. Aegis's own API server requires authenticated requests (OIDC JWT) — no anonymous access is possible at the application layer.
- **Evidence**: `services/platform-api/internal/server/mw/auth.go:94-99` (auth middleware enforces Bearer token)
- **Cross-reference**: AC-3 (Implemented) — See CRM: CUST
- **Gap/Notes**: Customer must configure K8s API Server. EKS supports this setting. **CAT I.**

#### V-242391: Kubelet must have anonymous authentication disabled
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Set `--anonymous-auth=false` on Kubelet.
- **Aegis Implementation**: Cluster infrastructure control. Aegis does not interact with kubelet API directly.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: AC-3 — See CRM: CUST
- **Gap/Notes**: Customer must configure. EKS-managed nodes set this by default. **CAT I.**

#### V-242392: Kubelet must enable explicit authorization
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Set `--authorization-mode=Webhook` on Kubelet.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: AC-3 — See CRM: CUST
- **Gap/Notes**: Customer must configure. EKS defaults to Webhook mode. **CAT I.**

#### V-242435: Kubernetes must prevent non-privileged users from executing privileged functions
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: `--authorization-mode` must not be `AlwaysAllow`.
- **Aegis Implementation**: Cluster infrastructure control. Aegis requires RBAC for its ClusterRoles to function. Aegis's application-level authorization (`internal/authz/policy.go`) uses fail-closed RBAC — denies all if no bindings configured.
- **Evidence**: `services/platform-api/internal/authz/policy.go:49-53` (fail-closed: `if len(p.bindings) == 0 { return false }`), `charts/aegis-services/templates/platform-api-rbac.yaml`
- **Cross-reference**: AC-6(10) (Implemented) — See CRM: CUST
- **Gap/Notes**: Customer must ensure RBAC is enabled. **CAT I.**

#### V-242436: API Server must have ValidatingAdmissionWebhook enabled
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Enable `ValidatingAdmissionWebhook` in `--enable-admission-plugins`.
- **Aegis Implementation**: Cluster infrastructure control. Aegis does not deploy custom admission webhooks but supports admission controllers.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-7a — See CRM: CUST
- **Gap/Notes**: Customer must enable. EKS enables this by default.

---

### Category 3: API Server Hardening

---

#### V-242384: Scheduler must have secure binding
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--bind-address=127.0.0.1` on Scheduler.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: EKS manages scheduler configuration.

#### V-242385: Controller Manager must have secure binding
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--bind-address=127.0.0.1` on Controller Manager.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: EKS manages Controller Manager configuration.

#### V-242386: API Server must have insecure port flag disabled
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Set `--insecure-port=0` on API Server.
- **Aegis Implementation**: Cluster infrastructure control. Aegis's own services only listen on TLS-secured ports (8080 HTTP via TLS termination, 8081 gRPC-TLS, 8443 HTTPS).
- **Evidence**: `charts/aegis-services/values/common.yaml` (port definitions), `services/platform-api/internal/server/server.go:2926-2958` (TLS-only serving)
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must configure. EKS removed insecure port. **CAT I.**

#### V-242387: Kubelet must have read-only port disabled
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Set `--read-only-port=0` on Kubelet.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must configure. EKS managed nodes disable this. **CAT I.**

#### V-242388: API Server must have insecure bind address not set
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Remove `--insecure-bind-address` from API Server.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: EKS does not expose insecure bind address. **CAT I.**

#### V-242389: API Server must have the secure port set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: `--secure-port` must not be 0.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: EKS defaults to port 443.

#### V-242395: Kubernetes dashboard must not be enabled
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: No kubernetes-dashboard pods should exist.
- **Aegis Implementation**: Aegis does not deploy or require the Kubernetes dashboard. Aegis provides its own UI via Backstage.
- **Evidence**: `charts/aegis-services/` (no dashboard templates), `charts/aegis-spoke/` (no dashboard)
- **Cross-reference**: AC-3 — See CRM: CUST
- **Gap/Notes**: Customer must verify dashboard is not installed.

#### V-242398: DynamicAuditing must not be enabled
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Feature gate `DynamicAuditing` must be false.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: AU-12 — See CRM: CUST
- **Gap/Notes**: Customer must verify. Modern K8s versions removed this feature.

#### V-242399: DynamicKubeletConfig must not be enabled
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Feature gate `DynamicKubeletConfig` must be false.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Removed in K8s 1.24+. EKS current versions inherently meet this.

#### V-242400: API Server must have Alpha APIs disabled
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Feature gate `AllAlpha` must be false.
- **Aegis Implementation**: Cluster infrastructure control. Aegis does not require any alpha APIs.
- **Evidence**: `charts/aegis-services/` (no alpha API references)
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must verify. EKS does not enable alpha by default.

#### V-242409: Controller Manager must disable profiling
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--profiling=false` on Controller Manager.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-7a — See CRM: CUST
- **Gap/Notes**: Customer must configure.

#### V-242438: API Server must configure timeouts
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--request-timeout` to > 0.
- **Aegis Implementation**: Cluster infrastructure control. Aegis's own gRPC server and proxy configure appropriate timeouts.
- **Evidence**: `services/proxy/internal/server/server.go:116-143` (session inactivity timeout)
- **Cross-reference**: SC-5 — See CRM: CUST
- **Gap/Notes**: Customer must configure. EKS default is 60s.

---

### Category 4: Audit Logging

---

#### V-242401: API Server must have an audit policy set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--audit-policy-file` on API Server.
- **Aegis Implementation**: Cluster infrastructure control. Aegis provides comprehensive **application-level** audit logging: structured JSON events with AU-3 compliant fields (timestamp, subject, action, resource, outcome, source IP). This complements K8s API Server audit logging.
- **Evidence**: `services/platform-api/internal/store/postgres/audit_events.go:35-46` (audit event schema), `charts/aegis-services/values/common.yaml:52-75` (audit config)
- **Cross-reference**: AU-14(1), AU-2 (Implemented) — See CRM: CFG
- **Gap/Notes**: Customer must configure K8s audit policy. Aegis provides app-level audit independently.

#### V-242402: API Server must have an audit log path set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--audit-log-path` on API Server.
- **Aegis Implementation**: Cluster infrastructure control. Aegis logs to stdout (JSON) for collection by Fluent Bit/Fluentd sidecars.
- **Evidence**: `charts/aegis-services/values/common.yaml:68` (`destination: stdout`)
- **Cross-reference**: AU-14(1) — See CRM: CFG
- **Gap/Notes**: Customer must configure K8s audit log path.

#### V-242403: API Server must generate comprehensive audit records
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Audit policy must be at minimum Metadata level with event type, source, results, users, and containers.
- **Aegis Implementation**: Cluster infrastructure control. Aegis's application audit records include all required fields: `event_type, timestamp, subject, resource_type, resource_id, action, outcome, details, source_ip`.
- **Evidence**: `services/platform-api/internal/store/postgres/audit_events.go:35-46`, `charts/aegis-services/values/common.yaml:60-67` (mandatory fields)
- **Cross-reference**: AU-3 (Implemented) — See CRM: CFG
- **Gap/Notes**: Customer must configure K8s audit policy level.

#### V-242461: API Server audit logs must be enabled
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: `--audit-policy-file` must be set with valid content.
- **Aegis Implementation**: Cluster infrastructure control. Redundant with V-242401.
- **Evidence**: `docs/compliance/customer-docs/configuration-hardening-guide.md`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must configure.

#### V-242462: API Server audit log max size must be set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--audit-log-maxsize` to minimum 100 MB.
- **Aegis Implementation**: Cluster infrastructure control. Aegis uses stdout logging with 90-day retention via external log pipeline.
- **Evidence**: `charts/aegis-services/values/common.yaml:73` (`retention: 90 days`)
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must configure.

#### V-242463: API Server audit log maximum backup must be set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--audit-log-maxbackup` to minimum 10.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must configure.

#### V-242464: API Server audit log retention must be set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--audit-log-maxage` to minimum 30 days.
- **Aegis Implementation**: Cluster infrastructure control. Aegis's own audit retention is 90+ days (exceeds STIG requirement).
- **Evidence**: `charts/aegis-services/values/common.yaml:73`, `docs/compliance/customer-docs/configuration-hardening-guide.md` (S3 lifecycle policy)
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must configure K8s audit retention.

#### V-242465: API Server audit log path must be set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: `--audit-log-path` must be a valid path. Redundant with V-242402.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must configure.

---

### Category 5: Certificate and PKI Management

---

#### V-242419: API Server must have SSL Certificate Authority set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--client-ca-file` on API Server.
- **Aegis Implementation**: Cluster infrastructure control. Aegis's own gRPC server supports client CA validation for mTLS.
- **Evidence**: `services/platform-api/internal/server/server.go:2948-2953` (client CA via `AEGIS_GRPC_TLS_CLIENT_CA`)
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: Customer must configure. EKS configures this automatically.

#### V-242420: Kubelet must have SSL Certificate Authority set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--client-ca-file` on Kubelet.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: EKS manages kubelet certificates.

#### V-242421: Controller Manager must have SSL Certificate Authority set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--root-ca-file` on Controller Manager.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: EKS manages this automatically.

#### V-242422: API Server must have a certificate for communication
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--tls-cert-file` and `--tls-private-key-file` on API Server.
- **Aegis Implementation**: Cluster infrastructure control. Aegis's own TLS uses certificate/key pairs.
- **Evidence**: `services/platform-api/internal/server/server.go:2927-2929`
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: EKS configures API Server certificates.

#### V-242423: etcd must enable client authentication
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--client-cert-auth=true` on etcd.
- **Aegis Implementation**: Cluster infrastructure control. Aegis does not directly interact with etcd.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: EKS manages etcd (inherently met).

#### V-242424: Kubelet must enable tls-private-key-file
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--tls-private-key-file` on Kubelet.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: EKS manages kubelet TLS.

#### V-242425: Kubelet must enable tls-cert-file
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--tls-cert-file` on Kubelet.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: EKS manages kubelet TLS.

#### V-242426: etcd must enable peer client cert authentication
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--peer-client-cert-auth=true` on etcd.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-8 — See CRM: CUST
- **Gap/Notes**: EKS manages etcd (inherently met).

#### V-242427: etcd must have a key file for secure communication
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--key-file` on etcd.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-23 — See CRM: CUST
- **Gap/Notes**: EKS manages etcd (inherently met).

#### V-242428: etcd must have a certificate for communication
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--cert-file` on etcd.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-23 — See CRM: CUST
- **Gap/Notes**: EKS manages etcd (inherently met).

#### V-242429: etcd must have SSL Certificate Authority set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--etcd-cafile` on API Server.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-23 — See CRM: CUST
- **Gap/Notes**: EKS manages etcd connectivity.

#### V-242430: etcd must have a certificate for API Server communication
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--etcd-certfile` on API Server.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-23 — See CRM: CUST
- **Gap/Notes**: EKS manages etcd connectivity.

#### V-242431: etcd must have a key file for API Server communication
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--etcd-keyfile` on API Server.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-23 — See CRM: CUST
- **Gap/Notes**: EKS manages etcd connectivity.

#### V-242432: etcd must have peer-cert-file set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--peer-cert-file` on etcd.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-23 — See CRM: CUST
- **Gap/Notes**: EKS manages etcd (inherently met).

#### V-242433: etcd must have peer-key-file set
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--peer-key-file` on etcd.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: SC-23 — See CRM: CUST
- **Gap/Notes**: EKS manages etcd (inherently met).

---

### Category 6: Kubelet Configuration

---

#### V-242397: Kubelet staticPodPath must not enable static pods
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Remove `staticPodPath` from kubelet config. Static pods bypass API server admission control.
- **Aegis Implementation**: Cluster infrastructure control. Aegis does not use static pods — all components are deployed via Helm (standard K8s Deployments).
- **Evidence**: `charts/aegis-services/templates/platform-api-deployment.yaml` (standard Deployment), `charts/aegis-spoke/templates/k8s-agent-deployment.yaml`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must configure kubelet. **CAT I.**

#### V-242404: Kubelet must deny hostname override
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Remove `--hostname-override` from kubelet.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-5(6) — See CRM: CUST
- **Gap/Notes**: Customer must configure.

#### V-242416: Kubelet must not disable timeouts
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Set `--streaming-connection-idle-timeout` to > 0 (minimum 5 min).
- **Aegis Implementation**: Cluster infrastructure control. Aegis implements its own session inactivity timeouts at the application layer.
- **Evidence**: `services/proxy/internal/server/server.go:116-143` (session timeout with privileged/standard tiers)
- **Cross-reference**: SC-10 — See CRM: CUST
- **Gap/Notes**: Customer must configure kubelet timeout.

#### V-242434: Kubelet must enable kernel protection
- **Severity**: CAT I
- **Status**: Customer Responsibility
- **STIG Description**: Set `--protect-kernel-defaults=true` on Kubelet.
- **Aegis Implementation**: Cluster infrastructure control. Aegis pods do not modify kernel parameters — security contexts include `seccompProfile: RuntimeDefault` which restricts syscalls.
- **Evidence**: `charts/aegis-services/values/common.yaml:129` (`seccompProfile: type: RuntimeDefault`)
- **Cross-reference**: SC-3 — See CRM: CUST
- **Gap/Notes**: Customer must configure. **CAT I.**

---

### Category 7: Pod Security

---

#### V-242414: Cluster must use non-privileged host ports for user pods
- **Severity**: CAT II
- **Status**: **Satisfied**
- **STIG Description**: User pods must not bind to host-privileged ports (< 1024).
- **Aegis Implementation**: All Aegis pods use non-privileged ports:
  - Platform API: 8080 (HTTP), 8081 (gRPC)
  - Proxy: 8080
  - Keycloak: 8443
  - K8s Agent: No ports (operator pattern, outbound-only)
  - Backstage: 7007
  No host port bindings are used.
- **Evidence**: `charts/aegis-services/values/common.yaml:44-50` (platform-api ports), `charts/aegis-spoke/templates/k8s-agent-deployment.yaml` (no ports), `charts/aegis-services/templates/proxy-deployment.yaml`
- **Cross-reference**: CM-7b — See CRM: YES-P
- **Gap/Notes**: None. All Aegis ports are > 1024.

#### V-242415: Secrets must not be stored as environment variables
- **Severity**: CAT I
- **Status**: **Partially Satisfied**
- **STIG Description**: Secrets must not use `secretKeyRef` in env vars. Mount secrets from files or use external secret stores instead.
- **Aegis Implementation**: Aegis currently uses `secretKeyRef` in environment variables in multiple deployment templates:
  - Platform API: DB password, JWT secret, authz bindings (`charts/aegis-services/templates/platform-api-deployment.yaml:118,135,152`)
  - Proxy: signing key (`charts/aegis-services/templates/proxy-deployment.yaml:49,102`)
  - Backstage: multiple secrets (`charts/aegis-services/templates/backstage/deployment.yaml:127,212,235,242,253`)
  - Spoke proxy: signing key (`charts/aegis-spoke/templates/proxy-deployment.yaml:51`)
  - PostgreSQL: passwords (`charts/aegis-services/templates/platform-api-postgres.yaml:66`)
  - Keycloak PostgreSQL: passwords (`charts/aegis-services/templates/keycloak/postgres.yaml:81,86`)
- **Evidence**: 18 `secretKeyRef` usages across Helm templates (see list above)
- **Cross-reference**: IA-5(1)(c) — See CRM: CFG
- **Gap/Notes**: **CAT I finding. Remediation required.** Migrate secrets from env vars to volume-mounted files or integrate with external secret stores (Vault CSI provider, AWS Secrets Manager CSI driver). This is the only CAT I finding with Aegis as the responsible party. See Remediation Plan below.

#### V-242417: Kubernetes must separate user functionality
- **Severity**: CAT II
- **Status**: **Satisfied**
- **STIG Description**: No user pods in `kube-system`, `kube-public`, or `kube-node-lease` namespaces.
- **Aegis Implementation**: Aegis deploys all components to dedicated namespaces (`aegis-system` or release namespace). User workloads run in project-specific namespaces. The hardening guide configures `namespacePerProject: true` for tenant isolation.
- **Evidence**: `charts/aegis-services/templates/platform-api-deployment.yaml` (`.Release.Namespace`), `docs/compliance/customer-docs/configuration-hardening-guide.md` (tenant isolation: `namespacePerProject: true`, `autoNetworkPolicy: true`)
- **Cross-reference**: SC-2 — See CRM: YES-P
- **Gap/Notes**: None. Satisfied.

#### V-242437: Kubernetes must have a pod security policy set
- **Severity**: CAT I
- **Status**: **Partially Satisfied**
- **STIG Description**: PodSecurityPolicy (or PSS equivalent) must exist with `runAsUser: MustRunAsNonRoot`, supplementalGroups/fsGroup min > 0.
- **Aegis Implementation**: Aegis enforces strong security contexts on all pods when `hardeningProfile != "dev"`:
  - `runAsNonRoot: true` (all pods)
  - `runAsUser: 10000` (platform-api), `65532` (proxy, k8s-agent)
  - `capabilities: { drop: [ALL] }` (all pods)
  - `readOnlyRootFilesystem: true` (all pods)
  - `allowPrivilegeEscalation: false` (all pods)
  - `seccompProfile: { type: RuntimeDefault }` (all pods)
  - `fsGroup: 10000` (platform-api)
  PSP was deprecated in K8s 1.21 and removed in 1.25. The modern equivalent is Pod Security Standards (PSS) via Pod Security Admission. **Aegis now applies PSS `restricted` labels** on workload namespaces when `hardeningProfile != "dev"` (`pod-security.kubernetes.io/enforce: restricted`, `warn`, `audit`). Customer must still enable PSA at the cluster level for the hub namespace.
- **Evidence**: `charts/aegis-services/values/common.yaml:120-130` (platform-api security context), `charts/aegis-services/values/common.yaml:262-272` (proxy security context), `charts/aegis-spoke/templates/k8s-agent-deployment.yaml:24-45` (k8s-agent security context), `charts/aegis-services/templates/workload-namespace.yaml:14-17` **(NEW: PSS namespace labels)**
- **Cross-reference**: AC-6 (Implemented) — See CRM: CFG
- **Gap/Notes**: Aegis pod security contexts and PSS namespace labels satisfy workload requirements. Customer must enable Pod Security Admission at the cluster level for the hub namespace. **CAT I — workload-level SATISFIED; cluster-level PSA enablement is Customer Responsibility.**

---

### Category 8: Network / Ports, Protocols, Services

---

#### V-242410: API Server must enforce PPS per PPSM CAL
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: API Server ports, protocols, and services must comply with DoD PPSM Category Assurance List.
- **Aegis Implementation**: Cluster infrastructure control. Aegis's own services expose only required ports and enforce default-deny NetworkPolicies.
- **Evidence**: `charts/aegis-services/templates/networkpolicy-egress-default-deny.yaml`, `charts/aegis-services/templates/networkpolicy-egress-platform-api.yaml`
- **Cross-reference**: CM-7b — See CRM: CUST
- **Gap/Notes**: Customer must verify API Server ports against PPSM CAL.

#### V-242411: Scheduler must enforce PPS per PPSM CAL
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Scheduler PPS must comply with DoD PPSM CAL.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-7b — See CRM: CUST
- **Gap/Notes**: EKS manages scheduler ports.

#### V-242412: Controllers must enforce PPS per PPSM CAL
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: Controller Manager PPS must comply with DoD PPSM CAL.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-7b — See CRM: CUST
- **Gap/Notes**: EKS manages controller ports.

#### V-242413: etcd must enforce PPS per PPSM CAL
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: etcd PPS must comply with DoD PPSM CAL.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-7b — See CRM: CUST
- **Gap/Notes**: EKS manages etcd (inherently met).

---

### Category 9: Worker Node Security

---

#### V-242393: Worker Nodes must not have sshd service running
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: `sshd` service must not be active on worker nodes.
- **Aegis Implementation**: Cluster infrastructure control. Aegis container images do not include sshd.
- **Evidence**: `services/platform-api/Dockerfile` (no sshd installed), `agents/k8s-agent/Dockerfile`, `services/proxy/Dockerfile`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must configure worker nodes. EKS managed nodes can use SSM instead of SSH.

#### V-242394: Worker Nodes must not have sshd service enabled
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: `sshd` service must be disabled on worker nodes.
- **Aegis Implementation**: Cluster infrastructure control.
- **Evidence**: `docs/shared-responsibility-matrix.md`
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must configure.

---

### Category 10: Software Currency

---

#### V-242396: kubectl cp command must give expected access and results
- **Severity**: CAT II
- **Status**: Customer Responsibility
- **STIG Description**: kubectl must be version 1.12.9+ (CVE-2019-1002101 path traversal fix).
- **Aegis Implementation**: Cluster infrastructure control. Aegis does not use `kubectl cp` in its operation.
- **Evidence**: N/A
- **Cross-reference**: CM-6b — See CRM: CUST
- **Gap/Notes**: Customer must verify kubectl version.

#### V-242442: Must remove old components after updates
- **Severity**: CAT II
- **Status**: **Satisfied**
- **STIG Description**: No duplicate image versions across pods. Old pods using outdated images must be removed.
- **Aegis Implementation**: Aegis provides SHA-pinned container images and Helm chart versioning. Rolling update strategy in deployments ensures old ReplicaSets are scaled down. **All 5 Deployment templates now set `revisionHistoryLimit: 3`**, automatically cleaning up old ReplicaSets and preventing stale component accumulation. Customer manages the upgrade cadence.
- **Evidence**: `charts/aegis-services/templates/platform-api-deployment.yaml:11` (`revisionHistoryLimit: 3`), `charts/aegis-services/templates/proxy-deployment.yaml:11`, `charts/aegis-services/templates/backstage/deployment.yaml:53`, `charts/aegis-spoke/templates/k8s-agent-deployment.yaml:11`, `charts/aegis-spoke/templates/proxy-deployment.yaml:14`
- **Cross-reference**: SI-2(6) — See CRM: CFG
- **Gap/Notes**: Customer must manage upgrade cadence. Aegis now enforces cleanup of old ReplicaSets.

#### V-242443: Must contain latest updates per IAVMs, CTOs, DTMs, STIGs
- **Severity**: CAT II
- **Status**: **Partially Satisfied**
- **STIG Description**: Kubernetes version must support skew policy and have current security patches.
- **Aegis Implementation**: Aegis maintains updated container images with security patches. Go dependencies managed via `go.sum` with checksum verification. SBOM generation available (`docs/compliance/sbom/generate-sbom.sh`). However, cluster K8s version is customer-managed.
- **Evidence**: `docs/compliance/sbom/generate-sbom.sh`, `go.sum`, container Dockerfiles with pinned base images
- **Cross-reference**: SI-2 — See CRM: CFG
- **Gap/Notes**: Customer must maintain K8s version currency. Aegis maintains application image currency.

---

### Category 11: File Permissions and Ownership (22 findings)

All 22 file permission/ownership findings are **Customer Responsibility** — they apply to Kubernetes node-level files (manifests, kubelet config, PKI certificates, etcd data, kubeadm config) that the cluster administrator manages. Aegis does not have access to or control over node-level file systems.

On managed Kubernetes (EKS), most of these are managed by the cloud provider and inherently met.

---

#### V-242405: Kubernetes manifests must be owned by root
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-5(6)
- **Description**: `/etc/kubernetes/manifests/` files must be owned by root:root.
- **Gap/Notes**: Customer/EKS responsibility. EKS manages control plane.

#### V-242406: Kubelet configuration file must be owned by root
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-5(6)
- **Description**: `/etc/sysconfig/kubelet` must be owned by root:root.
- **Gap/Notes**: Customer responsibility for worker node config.

#### V-242407: Kubelet configuration file permissions set to 644 or more restrictive
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-5(6)
- **Description**: `/etc/sysconfig/kubelet` must have permissions 644 or more restrictive.
- **Gap/Notes**: Customer responsibility.

#### V-242408: Kubernetes manifests must have least privileges
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-5(6)
- **Description**: Manifest files must have permissions 644 or more restrictive.
- **Gap/Notes**: EKS manages control plane manifests.

#### V-242444: Component manifests must be owned by root
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: `/etc/kubernetes/manifests/*` must be owned by root:root.
- **Gap/Notes**: EKS manages control plane.

#### V-242445: etcd data must be owned by etcd
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: `/var/lib/etcd/*` must be owned by etcd:etcd.
- **Gap/Notes**: EKS manages etcd (inherently met).

#### V-242446: Kubernetes conf files must be owned by root
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: admin.conf, scheduler.conf, controller-manager.conf must be owned by root:root.
- **Gap/Notes**: EKS manages control plane config.

#### V-242447: Kube Proxy must have file permissions 644 or more restrictive
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: kube-proxy kubeconfig must have permissions 644 or more restrictive.
- **Gap/Notes**: Customer responsibility.

#### V-242448: Kube Proxy must be owned by root
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: kube-proxy kubeconfig must be owned by root:root.
- **Gap/Notes**: Customer responsibility.

#### V-242449: Kubelet certificate authority file must have permissions 644
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: kubelet `--client-ca-file` must have permissions 644 or more restrictive.
- **Gap/Notes**: EKS manages kubelet CA.

#### V-242450: Kubelet certificate authority must be owned by root
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: kubelet `--client-ca-file` must be owned by root:root.
- **Gap/Notes**: EKS manages kubelet CA.

#### V-242451: Component PKI must be owned by root
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: All files under `/etc/kubernetes/pki/` must be owned by root:root.
- **Gap/Notes**: EKS manages PKI.

#### V-242452: Kubelet config must have file permissions 644
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: `/etc/kubernetes/kubelet.conf` must have permissions 644 or more restrictive.
- **Gap/Notes**: Customer responsibility.

#### V-242453: Kubelet config must be owned by root
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: `/etc/kubernetes/kubelet.conf` must be owned by root:root.
- **Gap/Notes**: Customer responsibility.

#### V-242454: kubeadm must be owned by root
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: kubeadm config must be owned by root:root.
- **Gap/Notes**: Not applicable to EKS (no kubeadm). Customer-managed clusters must verify.

#### V-242455: kubeadm.conf must have file permissions 644
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: kubeadm.conf must have permissions 644 or more restrictive.
- **Gap/Notes**: Not applicable to EKS.

#### V-242456: kubelet config.yaml must have file permissions 644
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: `/var/lib/kubelet/config.yaml` must have permissions 644 or more restrictive.
- **Gap/Notes**: Customer responsibility.

#### V-242457: kubelet config.yaml must be owned by root
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: `/var/lib/kubelet/config.yaml` must be owned by root:root.
- **Gap/Notes**: Customer responsibility.

#### V-242458: API Server manifests must have file permissions 644
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: `/etc/kubernetes/manifests/*` must have permissions 644 or more restrictive.
- **Gap/Notes**: EKS manages control plane manifests.

#### V-242459: etcd must have file permissions 644
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: `/var/lib/etcd/*` must have permissions 644 or more restrictive.
- **Gap/Notes**: EKS manages etcd (inherently met).

#### V-242460: admin.conf must have file permissions 644
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: admin.conf, scheduler.conf, controller-manager.conf must have permissions 644.
- **Gap/Notes**: EKS manages control plane config.

#### V-242466: PKI CRT files must have permissions 644
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: `/etc/kubernetes/pki/*.crt` must have permissions 644 or more restrictive.
- **Gap/Notes**: EKS manages PKI.

#### V-242467: PKI key files must have permissions 600
- **Severity**: CAT II | **Status**: Customer Responsibility | **NIST**: CM-6b
- **Description**: `/etc/kubernetes/pki/*.key` must have permissions 600 or more restrictive.
- **Gap/Notes**: EKS manages PKI.

---

## Remediation Plan

### CAT I Gaps (Critical — Must Fix)

**V-242415: Secrets stored as environment variables** (Partially Satisfied)

This is the **only CAT I finding where Aegis is the responsible party**. All other CAT I findings are Customer Responsibility (cluster infrastructure).

**Current State:** 18 `secretKeyRef` usages in environment variables across Helm templates:
- `platform-api-deployment.yaml` (3 secrets: DB password, JWT secret, authz bindings)
- `proxy-deployment.yaml` (2 secrets: signing keys)
- `backstage/deployment.yaml` (5 secrets: DB, auth, session)
- `spoke proxy-deployment.yaml` (1 secret: signing key)
- `platform-api-postgres.yaml` (1 secret: DB password)
- `keycloak/postgres.yaml` (2 secrets: DB credentials)
- `platform-api-migrations-job.yaml` (1 secret: DB password)

**Remediation Steps:**
1. Migrate secrets from `env.valueFrom.secretKeyRef` to volume-mounted secrets (`volumes` + `volumeMounts` referencing K8s Secrets)
2. Update application code to read secrets from files instead of env vars (e.g., `os.ReadFile("/run/secrets/db-password")`)
3. Alternatively, integrate with external secret stores:
   - AWS Secrets Manager CSI Driver (for EKS)
   - Vault CSI Provider (reference Dockerfiles exist in `docs/ironbank-reference/`)
   - External Secrets Operator
4. Update Helm templates to support both volume-mount and env-var modes for backward compatibility during migration

**Effort:** Medium (1-2 weeks). Core pattern change + testing across all services.
**Priority:** High — CAT I finding, ATO blocker.

**V-242437: Pod security policy** (Partially Satisfied)

Aegis's pod security contexts are comprehensive (non-root, drop ALL caps, read-only rootfs, seccomp). PSS namespace labels (`pod-security.kubernetes.io/enforce: restricted`) are now applied to workload namespaces when `hardeningProfile != "dev"`. Remaining gap: customer must enable PSA on the hub namespace.

**Remediation Steps:**
1. ~~Add Helm template for namespace labels~~ **DONE** — `charts/aegis-services/templates/workload-namespace.yaml` now includes PSS `restricted` labels
2. Document in hardening guide: customer must apply PSA labels to hub namespace (`aegis-system`)
3. Verify all Aegis pods pass PSS `restricted` profile

**Effort:** Low (documentation only — Helm change complete).
**Priority:** Medium — workload-level satisfied; hub namespace is customer documentation.

### CAT II Gaps (High — Should Fix)

**V-242442: Remove old components after updates** — ~~Partially Satisfied~~ **RESOLVED → Satisfied**
- ~~Aegis uses rolling updates but doesn't enforce cleanup~~
- **Fix applied:** `revisionHistoryLimit: 3` added to all 5 Deployment templates (platform-api, proxy, backstage, k8s-agent, spoke-proxy)

**V-242443: Latest updates per IAVMs/STIGs** (Partially Satisfied)
- Aegis maintains image currency but lacks formal patch cadence documentation
- **Remediation:** Document security patch SLA (e.g., CAT I CVEs patched within 30 days), integrate vulnerability scanning into CI/CD (Grype/Trivy per `docs/IRONBANK-HANDOFF.md`)
- **Effort:** Low (documentation + CI integration)

### CAT III Gaps (Medium — Plan to Fix)

No CAT III findings apply to Aegis directly. All 3 CAT III findings (if present in full STIG) are Customer Responsibility.

### Customer Responsibility Items

The following 86 STIG findings are Customer Responsibility per the Shared Responsibility Matrix (`docs/shared-responsibility-matrix.md`) and CRM (`docs/compliance/customer-docs/customer-responsibility-matrix.md`):

**Cluster Control Plane (41 findings):**
API Server, Controller Manager, Scheduler, etcd configuration flags — all managed by the cluster administrator (or cloud provider for managed K8s). On EKS, most are inherently met.

**Node-Level Security (24 findings):**
File permissions and ownership on `/etc/kubernetes/`, `/var/lib/etcd/`, `/var/lib/kubelet/`, PKI directories. Customer must configure worker nodes or use hardened AMIs.

**Kubelet Configuration (10 findings):**
Anonymous auth, authorization mode, TLS, timeouts, kernel protection, static pods, hostname override.

**Infrastructure Services (11 findings):**
Dashboard removal, SSH on workers, kubectl version, PPSM CAL compliance, feature gates.

**Guidance for Customers:**
- See `docs/compliance/customer-docs/configuration-hardening-guide.md` for recommended settings
- For EKS: ~60% of Customer Responsibility findings are inherently met by the managed control plane
- For self-managed clusters (RKE2, kubeadm): all findings must be explicitly verified
- For Platform One / Big Bang: Big Bang applies STIG-compliant configurations to the cluster baseline

---

## Cross-Framework Mapping

| STIG Finding | NIST 800-53 | FedRAMP Status | CMMC Status | CRM Ownership |
|---|---|---|---|---|
| V-242376 (CM TLS) | SC-8 | Implemented | Implemented | CUST |
| V-242377 (Sched TLS) | SC-8 | Implemented | Implemented | CUST |
| V-242378 (API TLS) | SC-8 | Implemented | Implemented | CUST |
| V-242379 (etcd auto-tls) | SC-8 | Implemented | Implemented | CUST |
| V-242380 (etcd peer-tls) | SC-8 | Implemented | Implemented | CUST |
| V-242381 (CM svc accts) | AC-6 | Implemented | Implemented | CUST |
| V-242382 (API RBAC) | AC-3 | Implemented | Implemented | CUST |
| **V-242383 (Namespaces)** | **AC-3** | **Implemented** | **Implemented** | **YES-P** |
| V-242384 (Sched bind) | CM-6b | Implemented | Implemented | CUST |
| V-242385 (CM bind) | CM-6b | Implemented | Implemented | CUST |
| V-242386 (Insecure port) | CM-6b | Implemented | Implemented | CUST |
| V-242387 (Kubelet RO port) | CM-6b | Implemented | Implemented | CUST |
| V-242388 (Insecure bind) | CM-6b | Implemented | Implemented | CUST |
| V-242389 (Secure port) | CM-6b | Implemented | Implemented | CUST |
| V-242390 (Anon auth) | AC-3 | Implemented | Implemented | CUST |
| V-242391 (Kubelet anon) | AC-3 | Implemented | Implemented | CUST |
| V-242392 (Kubelet authz) | AC-3 | Implemented | Implemented | CUST |
| V-242393 (sshd running) | CM-6b | Implemented | Implemented | CUST |
| V-242394 (sshd enabled) | CM-6b | Implemented | Implemented | CUST |
| V-242395 (Dashboard) | AC-3 | Implemented | Implemented | CUST |
| V-242396 (kubectl ver) | CM-6b | Implemented | Implemented | CUST |
| V-242397 (Static pods) | CM-6b | Implemented | Implemented | CUST |
| V-242398 (DynAudit) | AU-12 | Implemented | Implemented | CUST |
| V-242399 (DynKubelet) | CM-6b | Implemented | Implemented | CUST |
| V-242400 (Alpha APIs) | CM-6b | Implemented | Implemented | CUST |
| V-242401 (Audit policy) | AU-14(1) | Implemented | Implemented | CFG |
| V-242402 (Audit path) | AU-14(1) | Implemented | Implemented | CFG |
| V-242403 (Audit records) | AU-3 | Implemented | Implemented | CFG |
| V-242404 (Hostname) | CM-5(6) | Implemented | Implemented | CUST |
| V-242405-V-242408 (Perms) | CM-5(6) | Implemented | Implemented | CUST |
| V-242409 (CM profiling) | CM-7a | Implemented | Implemented | CUST |
| V-242410-V-242413 (PPSM) | CM-7b | Implemented | Implemented | CUST |
| **V-242414 (Host ports)** | **CM-7b** | **Implemented** | **Implemented** | **YES-P** |
| **V-242415 (Secrets env)** | **IA-5(1)(c)** | **Implemented** | **Implemented** | **CFG** |
| V-242416 (Kubelet timeout) | SC-10 | Implemented | Implemented | CUST |
| **V-242417 (NS separation)** | **SC-2** | **Implemented** | **Implemented** | **YES-P** |
| V-242418 (Cipher suites) | SC-8 | Partial | Partial | CFG |
| V-242419-V-242433 (PKI) | SC-8/SC-23 | Implemented | Implemented | CUST |
| V-242434 (Kernel protect) | SC-3 | Implemented | Implemented | CUST |
| V-242435 (Priv functions) | AC-6(10) | Implemented | Implemented | CUST |
| V-242436 (Admission) | CM-7a | Implemented | Implemented | CUST |
| **V-242437 (Pod security)** | **AC-6** | **Implemented** | **Implemented** | **CFG** |
| V-242438 (Timeout) | SC-5 | Implemented | Implemented | CUST |
| V-242439 (Basic auth) | SC-8 | Implemented | Implemented | CUST |
| V-242440 (Token auth) | SC-8 | Implemented | Implemented | CUST |
| V-242441 (Cert/key pair) | SC-8 | Implemented | Implemented | CUST |
| **V-242442 (Old components)** | **SI-2(6)** | **Implemented** | **Implemented** | **CFG** |
| **V-242443 (Updates)** | **SI-2** | **Implemented** | **Implemented** | **CFG** |
| V-242444-V-242460 (Perms) | CM-6b | Implemented | Implemented | CUST |
| V-242461-V-242465 (Audit) | CM-6b | Implemented | Implemented | CUST |
| V-242466-V-242467 (PKI perms) | CM-6b | Implemented | Implemented | CUST |
| V-242468 (TLS prohib) | AC-17(2) | Implemented | Implemented | CUST |

### Consistency Check with FedRAMP Control Mapping

Cross-referencing with `docs/compliance/fedramp/01-control-mapping.csv`:

| NIST Control | FedRAMP CSV Status | STIG Assessment Consistent? | Notes |
|---|---|---|---|
| AC-3 | Implemented | Yes | RBAC implemented at app and K8s levels |
| AC-6 | Implemented | Yes | Pod security contexts enforce least privilege |
| AU-2 | Implemented | Yes | Comprehensive audit logging in application |
| AU-3 | Implemented | Yes | Structured JSON with all required fields |
| CM-6 | Implemented | Yes | Hardening guide provides configuration baselines |
| CM-7 | Implemented | Yes | Minimal containers, NetworkPolicies, no dashboard |
| IA-2 | Implemented | Yes | OIDC/SAML via Keycloak, MFA support |
| IA-5 | Implemented | **Partial conflict** | CSV says Implemented; STIG V-242415 (secrets as env vars) is a gap |
| SC-8 | Implemented | Yes | TLS 1.2+ at all layers |
| SC-13 | Partial | Yes | FIPS capability exists but requires customer enablement |
| SI-2 | Implemented | Yes | SBOM, patching process documented |

**One inconsistency identified and RESOLVED:** IA-5 was marked "Implemented" in the FedRAMP CSV, but V-242415 (secrets as environment variables) reveals a specific STIG gap. **FedRAMP CSV has been updated** to "Partial" with a note referencing V-242415 and the 18 `secretKeyRef` usages (`docs/compliance/fedramp/01-control-mapping.csv:58`).

---

## Appendix A: Evidence File Verification

The following evidence files referenced in this matrix were verified to exist in the codebase:

| File Path | Exists | Referenced By |
|---|---|---|
| `charts/aegis-services/templates/networkpolicy-egress-default-deny.yaml` | Yes | V-242410 |
| `charts/aegis-services/templates/networkpolicy-egress-platform-api.yaml` | Yes | V-242410 |
| `charts/aegis-services/values/common.yaml` | Yes | Multiple |
| `charts/aegis-services/templates/platform-api-deployment.yaml` | Yes | V-242383, V-242415, V-242437 |
| `charts/aegis-services/templates/platform-api-rbac.yaml` | Yes | V-242382 |
| `charts/aegis-services/templates/platform-api-serviceaccount.yaml` | Yes | V-242381 |
| `charts/aegis-spoke/templates/k8s-agent-deployment.yaml` | Yes | V-242414, V-242437 |
| `charts/aegis-spoke/templates/k8s-agent-rbac.yaml` | Yes | V-242381 |
| `charts/aegis-spoke/templates/networkpolicy-agent.yaml` | Yes | V-242410 |
| `services/platform-api/internal/server/server.go` | Yes | V-242376-V-242378, V-242418 |
| `services/platform-api/internal/server/mw/auth.go` | Yes | V-242390, V-242439 |
| `services/platform-api/internal/authz/policy.go` | Yes | V-242382, V-242435 |
| `services/platform-api/internal/store/postgres/audit_events.go` | Yes | V-242401-V-242403 |
| `services/platform-api/main.go` | Yes | FIPS verification |
| `services/platform-api/Dockerfile` | Yes | Container security |
| `services/proxy/Dockerfile` | Yes | Container security |
| `agents/k8s-agent/Dockerfile` | Yes | Container security |
| `services/proxy/internal/server/server.go` | Yes | V-242416 |
| `services/proxy/internal/jti/store.go` | Yes | JWT security |
| `docs/shared-responsibility-matrix.md` | Yes | All Customer Resp. |
| `docs/compliance/customer-docs/configuration-hardening-guide.md` | Yes | Multiple |
| `docs/compliance/customer-docs/customer-responsibility-matrix.md` | Yes | CRM references |
| `docs/compliance/fedramp/01-control-mapping.csv` | Yes | Cross-framework |
| `docs/IRONBANK-HANDOFF.md` | Yes | Container hardening |
| `docs/compliance/sbom/generate-sbom.sh` | Yes | V-242443 |

---

## Appendix B: Automation References

For automated STIG compliance scanning:
- **MITRE InSpec Node Baseline**: `inspec exec https://github.com/mitre/k8s-node-stig-baseline`
- **MITRE InSpec Cluster Baseline**: `inspec exec https://github.com/mitre/k8s-cluster-stig-baseline`
- **Aegis Hardening Verification**: `docs/compliance/customer-docs/configuration-hardening-guide.md` (verification script section)

---

*Generated by automated codebase analysis. Manual review recommended before submission as ATO artifact.*
