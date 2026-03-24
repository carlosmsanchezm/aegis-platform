# Platform One and Iron Bank: Complete Guide for Aegis Platform

**Last Updated:** 2026-03-08
**Author:** Carlos Sanchez
**Status:** Research document -- actionable intelligence for DoD market entry
**Audience:** Aegis Technologies founder; 1-person company evaluating DoD ecosystem entry

---

## Table of Contents

1. [Executive Summary and Recommendation](#1-executive-summary-and-recommendation)
2. [Platform One Overview](#2-platform-one-overview)
3. [Iron Bank (Hardened Container Registry)](#3-iron-bank-hardened-container-registry)
4. [Big Bang Integration](#4-big-bang-integration)
5. [Party Bus / Software Factory](#5-party-bus--software-factory)
6. [DISA STIG Requirements](#6-disa-stig-requirements)
7. [FIPS 140-2/140-3 Requirements](#7-fips-140-2140-3-requirements)
8. [cATO (Continuous Authority to Operate)](#8-cato-continuous-authority-to-operate)
9. [Contact Points and Submission Portals](#9-contact-points-and-submission-portals)
10. [Success Stories](#10-success-stories)
11. [Estimated Timeline](#11-estimated-timeline)
12. [Aegis-Specific Action Plan](#12-aegis-specific-action-plan)

---

## 1. Executive Summary and Recommendation

**Bottom line:** Getting Aegis Platform into Iron Bank and Platform One is achievable for a 1-person company, but it is a 6-12 month process that requires significant technical hardening work before any formal submission. The good news: there is no minimum company size requirement, no cost to submit, and the DoD actively wants commercial innovation on Platform One.

**Recommended path for Aegis:**

1. Harden container images to Iron Bank standards (2-3 months of dev work)
2. Submit images to Iron Bank via the Container Hardening process (2-4 months review cycle)
3. Package Aegis as a Big Bang addon (1-2 months, can overlap with Iron Bank review)
4. Pursue cATO through a DoD sponsor/customer (ongoing, parallel track)

**Key constraint:** You need a DoD customer or sponsor who wants Aegis. Iron Bank accepts vendor-submitted images, but getting traction in the Platform One ecosystem is dramatically easier with an internal champion -- a program office, combatant command, or service branch that needs multi-cluster GPU orchestration.

---

## 2. Platform One Overview

### 2.1 What Is Platform One

Platform One (P1) is the **DoD Enterprise DevSecOps Initiative**, operated by the US Air Force's Chief Software Office (now under the Department of the Air Force Chief Digital and AI Officer, or CDAO). It provides:

- A standardized DevSecOps platform for all DoD branches
- Pre-approved, hardened software components
- CI/CD pipelines, container registries, and collaboration tools
- A path to continuous Authority to Operate (cATO)

Platform One is not a single product but an ecosystem of services:

| Service | What It Is | URL |
|---------|-----------|-----|
| **Iron Bank** | DoD's hardened container image registry | https://ironbank.dso.mil |
| **Big Bang** | Helm-based deployment framework for K8s | https://repo1.dso.mil/big-bang/bigbang |
| **Repo One** | DoD's GitLab instance (hosts all P1 code) | https://repo1.dso.mil |
| **Party Bus** | Program to evaluate and onboard new tools | https://p1.dso.mil |
| **IL2 DevSecOps** | Unclassified development environment | Available to .mil/.gov |
| **IL4/IL5 environments** | CUI and mission-critical environments | Requires higher clearance |

### 2.2 Why Platform One Matters

If software is on Platform One, it can be deployed by **any DoD organization** without each one doing a separate security review. This means:

- One hardening effort enables access to the entire DoD market (~3.4M personnel, ~$850B annual budget)
- Programs using Platform One get cATO, dramatically reducing their ATO timeline
- Platform One is the default DevSecOps platform for new DoD programs

### 2.3 How Third-Party Tools Get Into Platform One

There are three entry paths:

| Path | Description | Requirements | Best For |
|------|------------|--------------|----------|
| **Iron Bank image submission** | Get your container images into the hardened registry | Image hardening, STIG compliance, vulnerability remediation | Any containerized software |
| **Big Bang addon** | Package your tool as a Big Bang-compatible Helm chart | Iron Bank images + Helm chart + Istio compatibility | Kubernetes-native tools |
| **Party Bus evaluation** | Formal evaluation program for new capabilities | Sponsoring DoD organization + technical evaluation | Tools filling a gap in the P1 ecosystem |

**For Aegis, the recommended path is: Iron Bank images first, then Big Bang addon.**

### 2.4 Repository Access

| Resource | URL | Access |
|----------|-----|--------|
| Repo One (main Git server) | https://repo1.dso.mil | Mostly public; some resources need P1 account |
| Big Bang source | https://repo1.dso.mil/big-bang/bigbang | Public |
| Iron Bank | https://ironbank.dso.mil | Public browsing; pull requires credentials |
| Big Bang Docs | https://docs-bigbang.dso.mil | Public |
| P1 Mattermost | https://chat.il2.dso.mil | Requires P1 account |

---

## 3. Iron Bank (Hardened Container Registry)

### 3.1 What Is Iron Bank

Iron Bank is the DoD's centralized, hardened container image repository. It hosts container images that have been:

- Scanned for vulnerabilities (CVEs)
- Hardened according to DISA STIGs
- Rebuilt on approved base images (typically Red Hat UBI or Chainguard)
- Continuously monitored for new vulnerabilities
- Signed and attested for supply chain integrity

Iron Bank is hosted at **https://ironbank.dso.mil** and backed by a GitLab-based pipeline on Repo One (https://repo1.dso.mil).

As of 2025-2026, Iron Bank contains 1,500+ hardened container images from both open-source projects and commercial vendors.

### 3.2 Iron Bank Image Categories

| Category | Description | Examples |
|----------|------------|---------|
| **Community Images** | Open-source software hardened by P1 team or community | PostgreSQL, Redis, NGINX |
| **Vendor Images** | Commercial software hardened by the vendor | GitLab, Anchore, Twistlock |
| **DoD Images** | Images built by DoD programs | Mission-specific applications |

**Aegis would submit as a Vendor Image.** The vendor (you) maintains the image and is responsible for ongoing hardening and vulnerability remediation.

### 3.3 How to Get Container Images Approved

#### Step 1: Create a Repo One Account

- Go to https://login.dso.mil and register
- You need a CAC (Common Access Card), ECA (External Certificate Authority) certificate, or a Platform One SSO account
- **For non-DoD personnel:** You can get an ECA certificate from vendors like IdenTrust (~$100/year) or use Platform One SSO if sponsored
- Register at https://repo1.dso.mil

#### Step 2: Submit a Container Hardening Request

- Navigate to https://repo1.dso.mil/dsop (DoD Secure Open-source Project)
- Open an issue using the **Container Hardening** issue template
- Provide:
  - Container image name and version
  - Upstream source repository
  - Dockerfile or build instructions
  - Justification for inclusion (who needs it, what gap it fills)
  - Point of contact information

#### Step 3: Create the Hardened Dockerfile

Iron Bank images must follow strict requirements:

```dockerfile
# REQUIRED: Use an Iron Bank-approved base image
ARG BASE_REGISTRY=registry1.dso.mil
ARG BASE_IMAGE=ironbank/redhat/ubi/ubi9-minimal
ARG BASE_TAG=9.5

FROM ${BASE_REGISTRY}/${BASE_IMAGE}:${BASE_TAG}

# REQUIRED: Add standard labels
LABEL maintainer="carlos@aegis.dev"
LABEL name="aegis/platform-api"
LABEL version="1.0.0"
LABEL vendor="Aegis Technologies"
LABEL summary="Aegis Platform API - Multi-cluster GPU orchestration"
LABEL description="Control plane API for Aegis multi-cluster platform"
LABEL io.k8s.display-name="Aegis Platform API"

# REQUIRED: Run as non-root user
USER 1001:1001

# REQUIRED: No package manager in final image (multi-stage builds)
# REQUIRED: No secrets baked into the image
# REQUIRED: Health check endpoint
HEALTHCHECK --interval=30s --timeout=3s CMD ["/healthcheck"]

ENTRYPOINT ["/platform-api"]
```

#### Step 4: Image Hardening Requirements

| Requirement | Description | Aegis Status |
|-------------|------------|--------------|
| **Approved base image** | Must use Iron Bank base (UBI, Chainguard, Distroless) | Already using UBI9 minimal -- good |
| **Non-root execution** | Container must not run as root (UID 0) | Needs verification |
| **No package managers** | Final image must not contain yum/apt/apk | Multi-stage build handles this |
| **No embedded secrets** | No passwords, keys, tokens in image layers | Verify with secret scanning |
| **Minimal packages** | Only what is needed to run the application | Needs audit (current image has aws-cli, kubectl, jq, pulumi) |
| **CVE remediation** | Zero Critical, zero High CVEs (or justified waivers) | Needs scanning |
| **Read-only root filesystem** | Container filesystem should be read-only where possible | Needs implementation |
| **SBOM** | Software Bill of Materials (CycloneDX or SPDX format) | SBOM generation already set up |
| **Digital signature** | Image must be signed (cosign) | Not yet implemented |
| **Documentation** | README, license, changelog in repo | Partially done |
| **Health checks** | Must expose health/readiness endpoints | Already implemented |
| **Security contexts** | Must define SecurityContext with all restrictions | Partially implemented |

#### Step 5: Pipeline Integration

Iron Bank uses a CI pipeline on Repo One that automatically:

1. Rebuilds your image from source (reproducible builds)
2. Scans with multiple tools: Anchore/Grype, Twistlock/Prisma Cloud, OpenSCAP
3. Evaluates against STIG checklist
4. Generates findings report
5. Requires remediation or justification for all findings

#### Step 6: Review and Approval

- The Iron Bank team reviews your submission
- They may request changes to your Dockerfile, dependencies, or configuration
- Once approved, your image is published to `registry1.dso.mil/ironbank/your-org/your-image`
- You are responsible for updating the image when new vulnerabilities are found

### 3.4 Can a 1-Person Company Submit Images?

**Yes.** There is no minimum company size requirement. Iron Bank accepts submissions from:

- DoD programs
- Commercial vendors of any size
- Open-source maintainers
- Individual contributors (with sponsorship)

**Practical considerations for a solo founder:**

| Factor | Impact | Mitigation |
|--------|--------|-----------|
| Ongoing maintenance burden | Must respond to CVE findings within SLA | Automate scanning and rebuilds in CI |
| Repo One account access | Need ECA cert or P1 SSO | IdenTrust ECA cert ~$100/year |
| Review cycle communication | Must be responsive to P1 team questions | Set up notifications, check daily during review |
| No dedicated security team | Iron Bank expects security POC | You are the security POC |

### 3.5 Vulnerability Remediation SLAs

Once your image is in Iron Bank, you must maintain it:

| Severity | Remediation SLA | What Happens If Missed |
|----------|----------------|----------------------|
| Critical | 30 days | Image marked as non-compliant, may be delisted |
| High | 90 days | Image marked as non-compliant |
| Medium | 180 days | Tracked, expected remediation |
| Low | Best effort | Tracked |

You can request **justifications** (waivers) for findings that are false positives or not applicable, but these must be individually documented and approved.

---

## 4. Big Bang Integration

### 4.1 What Is Big Bang

Big Bang is Platform One's **declarative deployment framework** for Kubernetes. It is a Helm umbrella chart (plus Flux GitOps) that deploys a complete DevSecOps stack:

**Core components (always deployed):**
- Istio (service mesh, mTLS, traffic management)
- Kiali (service mesh observability)
- Monitoring (Prometheus, Grafana)
- Logging (EFK or PLG stack: Elasticsearch/Loki, Fluentbit/Promtail, Kibana/Grafana)
- Jaeger (distributed tracing)
- Policy enforcement (OPA Gatekeeper, Kyverno)
- Runtime security (Twistlock/Prisma Cloud or NeuVector)

**Addons (optional, modular):**
- GitLab, Nexus, SonarQube, Mattermost, Keycloak, Anchore, ArgoCD, MinIO, Vault, and many more

Source: https://repo1.dso.mil/big-bang/bigbang

### 4.2 What It Means to Be a Big Bang Addon

A Big Bang addon is a Helm chart that:

1. Uses only Iron Bank container images
2. Is compatible with the Istio service mesh (sidecar injection)
3. Integrates with Big Bang's monitoring and logging stack
4. Follows Big Bang's values interface conventions
5. Is tested within the Big Bang CI pipeline
6. Lives in the Big Bang repository (or is referenced as a third-party chart)

### 4.3 Technical Requirements for Big Bang Compatibility

#### 4.3.1 Istio Service Mesh Compatibility

This is the most impactful technical requirement. All traffic between pods goes through Istio Envoy sidecars.

| Requirement | Description | Aegis Impact |
|-------------|------------|-------------|
| **Sidecar injection** | Pods must tolerate Istio sidecar injection | Test all pods with istio-proxy sidecar |
| **mTLS** | Istio enforces mTLS between all services | Must not conflict with Aegis's own mTLS (step-ca) |
| **Port naming** | Service ports must follow Istio naming (`http-*`, `grpc-*`, `tcp-*`) | Rename service ports in Helm chart |
| **Protocol detection** | Istio needs to detect protocols correctly | gRPC ports need `grpc-` prefix |
| **Init container ordering** | Istio sidecar must start before app | Verify startup ordering |
| **Liveness/readiness probes** | Must work through Istio proxy | HTTP probes usually work; gRPC probes need config |
| **VirtualService / Gateway** | Ingress traffic managed by Istio | Add Istio VirtualService templates |

**Critical for Aegis:** The hub's gRPC communication with spoke agents must work through Istio. The spoke-proxy reverse proxy and WebSocket connections must also be compatible. Test thoroughly.

#### 4.3.2 Monitoring Integration

```yaml
# ServiceMonitor for Prometheus Operator
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: aegis-platform-api
spec:
  selector:
    matchLabels:
      app: platform-api
  endpoints:
  - port: metrics
    path: /metrics
```

Additionally:
- Expose Prometheus metrics endpoint (already noted in Aegis OBSERVABILITY.md)
- Provide Grafana dashboard JSON (optional but strongly recommended)
- Pod annotations for scraping:

```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9090"
    prometheus.io/path: "/metrics"
```

#### 4.3.3 Logging Integration

- Logs must go to stdout/stderr (12-factor app pattern)
- Structured JSON logging preferred
- Include standard fields: timestamp, level, message, component
- Fluentbit/Promtail will collect and forward to Elasticsearch/Loki

#### 4.3.4 Network Policies

Big Bang deploys restrictive NetworkPolicies by default. Your chart must:

- Declare all required ingress/egress in NetworkPolicy resources
- Work with the default-deny posture
- Allow Istio sidecar traffic
- Allow traffic from Istio ingress gateway
- Document any external connectivity requirements

**Note:** Aegis already has NetworkPolicy templates (`networkpolicy-egress-default-deny.yaml`, `networkpolicy-egress-platform-api.yaml`, `networkpolicy-egress-proxy.yaml`). These are a strong foundation.

#### 4.3.5 Security Context

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1001
  runAsGroup: 1001
  fsGroup: 1001
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  seccompProfile:
    type: RuntimeDefault
```

All containers must run as non-root with dropped capabilities.

#### 4.3.6 SSO Integration

- Accept OIDC tokens from Big Bang's Keycloak instance
- Configure via environment variables (issuer URL, client ID, etc.)
- Aegis already uses Keycloak -- just needs to be configurable to point at Big Bang's instance

### 4.4 How to Package a Helm Chart for Big Bang

#### 4.4.1 Values Interface

Big Bang addons must expose a standardized values interface:

```yaml
# Standard Big Bang addon values structure
domain: bigbang.dev

istio:
  enabled: true
  injection: enabled
  hardened:
    enabled: true
    customAuthorizationPolicies: []

monitoring:
  enabled: true

networkPolicies:
  enabled: true
  controlPlaneCidr: "0.0.0.0/0"

openshift: false

# SSO integration (Big Bang uses Keycloak by default)
sso:
  enabled: false
  client_id: ""
  client_secret: ""
  # Keycloak URL injected by Big Bang

# Image pull configuration pointing to Iron Bank
image:
  repository: registry1.dso.mil/ironbank/aegis/platform-api
  tag: "1.0.0"
  pullPolicy: IfNotPresent

imagePullSecrets:
  - name: private-registry
```

#### 4.4.2 Chart Structure

```
chart/
  Chart.yaml
  values.yaml
  templates/
    _helpers.tpl
    deployment.yaml
    service.yaml
    networkpolicy.yaml           # Required
    istio-virtualservice.yaml    # Istio ingress
    istio-sidecar.yaml           # Istio PeerAuthentication
    servicemonitor.yaml          # Prometheus integration
    tests/
      test-connection.yaml
  docs/
    overview.md
    keycloak.md                  # SSO integration docs
  CHANGELOG.md
```

#### 4.4.3 Testing Requirements

| Test Type | Description | Tool |
|-----------|------------|------|
| **Helm lint** | Chart passes `helm lint` | Helm CLI |
| **Template rendering** | Templates render correctly with default values | `helm template` |
| **Kubernetes validation** | Resources are valid K8s manifests | `kubeval` or `kubeconform` |
| **Big Bang integration** | Deploys successfully within Big Bang stack | Big Bang CI pipeline |
| **Istio compatibility** | Works with Istio sidecar injection | Manual or CI test |
| **RKE2 testing** | Big Bang prefers RKE2 as its K8s distribution | Test on RKE2 cluster |
| **Cypress/functional** | Application works end-to-end in Big Bang env | Optional but recommended |

### 4.5 Submitting a Big Bang Addon

1. Fork the Big Bang repo on Repo One (or create a new project in the `big-bang/product` namespace)
2. Add your addon chart following the structure above
3. Add your addon to the Big Bang umbrella chart's `addons` section
4. Submit a merge request
5. The Big Bang team reviews for compliance with their standards
6. CI pipeline runs full integration tests
7. Once merged, your addon is available to all Big Bang deployments

---

## 5. Party Bus / Software Factory

### 5.1 What Is Party Bus

Party Bus is Platform One's program for **evaluating and onboarding new software tools** into the P1 ecosystem. It acts as the intake and assessment process for tools that want to be part of Platform One's supported catalog.

The name comes from the idea that tools are "getting on the bus" -- joining the Platform One ecosystem.

### 5.2 How It Works

| Phase | Activities | Duration |
|-------|-----------|----------|
| **1. Nomination** | A DoD organization nominates a tool for evaluation | 1-2 weeks |
| **2. Intake** | P1 team reviews the nomination, assesses fit | 2-4 weeks |
| **3. Technical Evaluation** | Security scanning, architecture review, integration testing | 4-8 weeks |
| **4. Hardening** | Container images hardened for Iron Bank | 4-12 weeks |
| **5. Integration** | Big Bang addon packaging and testing | 4-8 weeks |
| **6. Approval** | Final review and addition to P1 catalog | 2-4 weeks |

### 5.3 Requirements for Party Bus

| Requirement | Description |
|-------------|------------|
| **DoD sponsor** | A DoD organization must nominate/request the tool |
| **Gap justification** | Tool must fill a capability gap not met by existing P1 tools |
| **Containerized** | Must be deployable as containers on Kubernetes |
| **Vendor cooperation** | Vendor must participate in hardening and provide source/build info |
| **Licensing** | Licensing model must be compatible with DoD-wide deployment |

### 5.4 How to Get Into Party Bus

**You cannot self-nominate.** A DoD organization must request your tool. However, you can:

1. **Build relationships with DoD programs** that need GPU orchestration
2. **Present at DoD tech events** (Platform One Demo Days, AFWERX, DIU events)
3. **Engage the Platform One team directly** via Mattermost (see Section 9)
4. **Publish your Iron Bank images first** -- this demonstrates commitment and maturity
5. **Get into the SBIR/STTR pipeline** -- DoD small business innovation programs

### 5.5 Alternative: Bypass Party Bus

You do not strictly need Party Bus to be useful in the DoD ecosystem. If you:

1. Get your images into Iron Bank
2. Package a Big Bang-compatible addon
3. Have a DoD customer who wants to deploy it

...then that customer can deploy Aegis on their Big Bang cluster directly. Party Bus is about being part of the *officially supported catalog*, but DoD programs can deploy any Iron Bank image.

---

## 6. DISA STIG Requirements

### 6.1 What STIGs Apply

DISA STIGs (Security Technical Implementation Guides) are the DoD's configuration standards. For Aegis, these STIGs apply:

| STIG | ID | Version (as of 2025) | Applies To |
|------|----|---------------------|-----------|
| **Container Platform SRG** | V-257523 series | v2r2+ | Kubernetes cluster configuration |
| **Kubernetes STIG** | V-242376 series | v2r2+ | K8s API server, kubelet, etcd, controller-manager |
| **Container Image STIG** | (Part of Container Platform) | v2r2+ | Your container images |
| **Application Security STIG** | V-222384 series | v5r3+ | The Aegis application itself |
| **Web Server STIG** | V-241785 series (if applicable) | Latest | Proxy, Keycloak if exposed via web |
| **RHEL 8/9 STIG** | (if using UBI base) | Latest | Base image configuration |

All STIGs are available at https://public.cyber.mil/stigs/downloads/

### 6.2 Kubernetes STIG (V-242376 Series)

This applies to the Kubernetes cluster Aegis runs on. For Aegis as a vendor, you need to ensure your deployment works on a STIG-compliant cluster.

Key requirements that affect Aegis:

| STIG ID | Requirement | Impact on Aegis |
|---------|------------|----------------|
| V-242376 | Pods must run as non-root | Implement in SecurityContext |
| V-242377 | Pods must not allow privilege escalation | `allowPrivilegeEscalation: false` |
| V-242378 | Pods must have resource limits | Set CPU/memory limits in Helm values |
| V-242379 | Pods must use read-only root filesystem | Where feasible |
| V-242380 | Container images from approved registry | Iron Bank images |
| V-242381 | RBAC enabled, no default SA | Use dedicated ServiceAccounts (already done) |
| V-242382 | Network policies must restrict traffic | Already have NetworkPolicies |
| V-242383 | Pods must not run as privileged | Verify all pods have `privileged: false` |
| V-242384 | Host namespaces not shared | No `hostNetwork`, `hostPID`, `hostIPC` |
| V-242385 | TLS 1.2+ for all communications | Already implemented |
| V-242395 | Secrets encrypted at rest | Cluster-level config (customer responsibility) |
| V-242397 | Network policies enforced | Already have NetworkPolicies |
| V-242414 | Pod Security Standards enforced | Must pass `restricted` profile |
| V-242415 | Namespaces with resource quotas | Document resource requirements |

### 6.3 Container Image STIG

| Requirement | Description | Aegis Action |
|-------------|------------|-------------|
| No root user | Container process must not run as UID 0 | Verify all Dockerfiles |
| No SETUID/SETGID binaries | Remove or document any SUID/SGID files | Audit base image |
| Minimal packages | Only necessary packages installed | Already using multi-stage builds |
| Health checks | Container must expose health endpoints | Already implemented |
| No embedded credentials | No secrets in image layers | Run secret scanner on images |
| Image signing | Images must be digitally signed | Implement cosign signing |
| Vulnerability scanning | Zero Critical/High unmitigated | Run Trivy/Grype regularly |
| Base image currency | Must use supported, patched base images | Already on UBI9 |

### 6.4 Application Security STIG Checklist

The Application Security and Development STIG (ASD STIG) is the most comprehensive and subjective. Key categories:

| Category | Key Requirements | Aegis Relevance |
|----------|-----------------|----------------|
| **Authentication (IA)** | MFA support, session management, account lockout, password policy | Keycloak handles this; document configuration |
| **Access Control (AC)** | RBAC, least privilege, separation of duties | Platform API RBAC; document roles |
| **Audit (AU)** | Log all security-relevant events, protect log integrity, retention policy | Structured logging; audit trail |
| **Encryption (SC)** | TLS 1.2+, FIPS-validated crypto, AES-256 at rest, key management | FIPS Go build + TLS config |
| **Input Validation (SI)** | Sanitize all inputs, prevent injection, parameterized queries | gRPC protobuf handles most; audit HTTP handlers |
| **Session Management** | Timeout, termination, invalidation, secure tokens | JWT with short TTL |
| **Error Handling** | No sensitive data in error messages | Audit error responses |
| **Configuration** | Secure defaults, no unnecessary services | Document secure configuration |
| **Software Updates** | Patch management, vulnerability response | Automated CI/CD pipeline |

### 6.5 How STIGs Are Evaluated

Iron Bank's pipeline runs **OpenSCAP** against your images using STIG profiles. Results are categorized:

| Category | Meaning | Action Required |
|----------|---------|----------------|
| **CAT I** | Critical vulnerability | Must fix before approval |
| **CAT II** | Medium risk | Must fix or provide justification |
| **CAT III** | Low risk | Should fix, can justify deferral |
| **Not Applicable** | Control does not apply | Document why |

---

## 7. FIPS 140-2/140-3 Requirements

### 7.1 What FIPS Compliance Means

FIPS 140-2 (and its successor FIPS 140-3) is a US government standard for cryptographic modules. For DoD systems, **all cryptographic operations must use FIPS-validated modules**. This is not optional.

What this means practically:

| Layer | Requirement | Who Is Responsible |
|-------|------------|-------------------|
| TLS termination | Must use FIPS-validated TLS library | Aegis (application) |
| Data at rest encryption | Must use FIPS-validated encryption | Customer (infrastructure) |
| Key generation | Keys generated using FIPS DRBG | Aegis (application) |
| Hashing | Must use FIPS-approved algorithms (SHA-256+) | Aegis (application) |
| Token signing | JWT signing must use FIPS-validated crypto | Keycloak configuration |

### 7.2 FIPS for Go Applications (BoringCrypto / Go FIPS)

Go 1.24+ (which Aegis uses) has native FIPS 140-3 support:

#### Option A: GODEBUG=fips140=only (Go 1.24+, Recommended)

Starting with Go 1.24 (released Feb 2025), Go has a built-in FIPS 140-3 module:

```bash
# At runtime, set environment variable
GODEBUG=fips140=only ./platform-api
```

This causes:
- All crypto operations use Go's FIPS 140-3 validated module
- Non-FIPS algorithms (like MD5, RC4) return errors if called
- TLS negotiation only accepts FIPS-approved cipher suites
- No rebuild required -- same binary, different runtime behavior

**Go FIPS module status (as of early 2026):**
- Go's cryptographic module is in the CMVP (Cryptographic Module Validation Program) queue
- Google submitted Go's BoringCrypto-based module; NIST validation is pending
- Many DoD programs accept "FIPS-consistent" implementations while awaiting formal validation
- The Go team has stated the module is designed to meet FIPS 140-3 Level 1

#### Option B: BoringCrypto (Legacy, Still Valid)

The older approach using `GOEXPERIMENT=boringcrypto`:

```bash
# Build with BoringCrypto
GOEXPERIMENT=boringcrypto CGO_ENABLED=1 go build -o platform-api .
```

This links against Google's BoringSSL FIPS module, which **has a FIPS 140-2 validation certificate** (Certificate #4407 and others). Requires CGO and a C compiler in the build environment.

#### Recommendation for Aegis

**Use `GODEBUG=fips140=only` (Option A).** Reasons:

1. Already referenced in Aegis spoke templates (`charts/aegis-spoke/templates/proxy-deployment.yaml`, `k8s-agent-deployment.yaml`)
2. No CGO dependency -- cleaner builds
3. Same binary for dev and production (FIPS is a runtime flag)
4. Go team's forward direction; BoringCrypto experiment may be deprecated

**What Aegis already has:**
- Dockerfile uses UBI9 minimal base with crypto-policies package installed
- Helm charts set `GODEBUG=fips140=only` on spoke components
- `AGENT_DEPLOYMENT_GUIDE.md` documents FIPS on/off for dev vs production

**What still needs work:**
- Enable `GODEBUG=fips140=only` on ALL services (hub too, not just spoke)
- Test that all TLS connections work in FIPS mode
- Verify no non-FIPS crypto libraries are used (some JWT libraries may be affected)
- Fix the Pulumi SHA-1 conflict (see Section 12.3)
- Document the specific FIPS module version and CMVP certificate number

### 7.3 FIPS for Container Images

Beyond the application, the base OS image should support FIPS:

| Component | Requirement | Aegis Status |
|-----------|------------|-------------|
| **Base image** | Use a FIPS-enabled base (UBI9 with crypto-policies) | Already on UBI9 -- set `FIPS` crypto policy |
| **OpenSSL** | Must be FIPS-validated OpenSSL | UBI9 ships FIPS-validated OpenSSL |
| **SSH** | If used, must use FIPS mode | Not applicable (no SSH in service containers) |
| **System crypto policy** | Set to FIPS | `update-crypto-policies --set FIPS` |

For Iron Bank images specifically:

```dockerfile
# In the final stage
RUN update-crypto-policies --set FIPS
```

**Caveat:** The current Aegis Dockerfile intentionally does NOT bake FIPS in because "Pulumi uses SHA-1 internally for resource URNs." For Iron Bank, you would need a separate image build that does not include Pulumi, or handle FIPS enablement differently.

### 7.4 FIPS for Keycloak

- Keycloak 20+ supports FIPS mode
- Requires FIPS-enabled JDK (Red Hat build of OpenJDK with FIPS)
- Configure via `KC_FIPS_MODE=strict` or `KC_FIPS_MODE=non-strict`
- Iron Bank has a hardened Keycloak image that may already be FIPS-configured

### 7.5 FIPS Validation vs. FIPS Compliance

This distinction is critical:

| Term | Meaning | Cost | Timeline |
|------|---------|------|----------|
| **FIPS Validated** | Cryptographic module has been tested and certified by NIST/CMVP accredited lab | $50K-$500K+ | 12-24 months |
| **FIPS Compliant** | Uses FIPS-validated modules (someone else did the validation) | $0 | Immediate |
| **FIPS Consistent** | Implements the same algorithms/modes as FIPS but not formally validated | $0 | Immediate |

**You do NOT need to get your own FIPS validation.** You need to **use** FIPS-validated modules:

- Go's BoringCrypto has an existing FIPS 140-2 certificate (via Google, CMVP #4407)
- Go 1.24+ native FIPS module is in the CMVP queue (pending validation)
- UBI9's OpenSSL is FIPS-validated (via Red Hat)
- AWS infrastructure services are FIPS-validated (via AWS)

**You cannot self-attest FIPS.** The module must be validated by an accredited lab. But since you are using already-validated modules (BoringCrypto, OpenSSL from UBI9, AWS KMS), you document which validated modules you use and their certificate numbers.

**For a 1-person company:** Use FIPS-validated modules and document them. Do NOT pursue formal FIPS validation of your product -- that is for crypto module vendors, not application developers.

### 7.6 FIPS Documentation Requirements

For Iron Bank and DoD deployments, you need:

1. **Cryptographic inventory** -- List every place crypto is used
2. **Module mapping** -- Map each crypto use to a FIPS-validated module
3. **Certificate references** -- CMVP certificate numbers for each module
4. **Configuration evidence** -- Proof that FIPS mode is enabled

Example inventory for Aegis:

| Crypto Use | Algorithm | Module | CMVP Cert |
|-----------|-----------|--------|-----------|
| TLS (platform-api) | TLS 1.2/1.3, AES-256-GCM, ECDHE | Go BoringCrypto / Go FIPS | #4407 (pending for Go native) |
| TLS (proxy) | TLS 1.2/1.3 | Go BoringCrypto / Go FIPS | Same |
| JWT signing (Keycloak) | RS256 | Red Hat OpenJDK FIPS | Check Red Hat cert |
| mTLS certificates | X.509, ECDSA P-256 | Go BoringCrypto / Go FIPS | Same |
| TLS (Keycloak) | TLS 1.2/1.3 | Red Hat OpenSSL | Red Hat FIPS cert |

---

## 8. cATO (Continuous Authority to Operate)

### 8.1 What Is cATO

cATO (continuous Authority to Operate) is the DoD's evolution of the traditional ATO process. Instead of a point-in-time security assessment every 3 years, cATO provides:

- **Continuous monitoring** of security posture
- **Automated compliance checking** via DevSecOps pipelines
- **Real-time risk dashboards** instead of static documents
- **Faster deployment** of new features (no re-authorization for every change)

cATO was formalized by the DoD CIO memo in February 2022 and is now the preferred authorization model for DoD software.

### 8.2 How cATO Relates to Platform One

Platform One is the primary enabler of cATO for DoD programs:

| Traditional ATO | cATO via Platform One |
|----------------|----------------------|
| 12-18 month process | Initial assessment, then continuous |
| Paper-based evidence | Automated evidence collection |
| Point-in-time scan | Continuous scanning |
| Re-assess every 3 years | Ongoing authorization |
| Per-system assessment | Inherit Platform One's assessment |

**The key benefit:** If a DoD program deploys Aegis on a Platform One (Big Bang) cluster that already has a cATO, Aegis inherits many of the infrastructure-level controls. The program only needs to assess Aegis-specific application controls.

### 8.3 How Self-Hosted Software Fits In

Aegis as self-hosted software (Model B) fits the cATO model as follows:

| Control Layer | Responsibility | cATO Mechanism |
|---------------|---------------|----------------|
| Infrastructure (AWS GovCloud) | Customer/Cloud Provider | AWS FedRAMP authorization |
| Kubernetes platform | Customer (Big Bang) | Platform One cATO |
| Container images | Aegis (Iron Bank) | Iron Bank continuous scanning |
| Application configuration | Customer (with Aegis guidance) | Customer's cATO pipeline |
| Application code | Aegis | Vendor responsibility (SBOMs, patching) |

**What Aegis must provide for customer cATO:**

| Document | Purpose |
|----------|---------|
| **Iron Bank images** | Continuously scanned, maintained |
| **SBOM** | Updated with each release |
| **Hardening guide** | How to configure Aegis securely (already drafted) |
| **Customer Responsibility Matrix** | What you provide vs. what customer configures (already exists) |
| **STIG checklist** | Completed STIG self-assessment for your application |
| **Vulnerability scan results** | Current CVE scan of your container images |
| **Architecture diagrams** | Network flows, data flows, authentication flows |
| **Vulnerability disclosure process** | How you handle and communicate CVEs |
| **Patch SLAs** | Committed timelines for security patches |
| **Incident response procedures** | How to handle security incidents involving your software |

### 8.4 OSCAL (Open Security Controls Assessment Language)

OSCAL is a **machine-readable format** for security documentation (NIST standard). Instead of Word documents, your security controls are expressed in JSON/XML/YAML.

Why it matters:
- FedRAMP is transitioning to OSCAL-based submissions
- DoD is adopting OSCAL for cATO pipelines
- Automated compliance tools consume OSCAL
- Shows technical maturity to government evaluators

Aegis already has OSCAL catalog files (`docs/oscal/`). The next step is creating an OSCAL System Security Plan (SSP) component definition that maps Aegis capabilities to specific controls.

### 8.5 cATO Requirements Summary

To support a customer's cATO, Aegis needs:

| Requirement | Status | Priority |
|-------------|--------|----------|
| Iron Bank images with continuous scanning | Not started | High |
| Automated SBOM generation per release | Tooling exists, not automated in CI | High |
| STIG compliance documentation | Not started | High |
| Vulnerability disclosure policy | Not formalized | Medium |
| Patch SLA commitment | Not formalized | Medium |
| Hardening guide | Draft exists | Medium |
| Customer Responsibility Matrix | Exists | Low (update) |
| OSCAL component definition | Catalogs exist, SSP component needed | Medium |
| Prometheus metrics for security monitoring | Partially implemented | Medium |

---

## 9. Contact Points and Submission Portals

### 9.1 Primary Contact Channels

| Channel | URL/Contact | Purpose |
|---------|------------|---------|
| **Platform One Website** | https://p1.dso.mil | Overview, program info |
| **Repo One (GitLab)** | https://repo1.dso.mil | Code repos, Iron Bank submissions, issues |
| **Iron Bank Portal** | https://ironbank.dso.mil | Browse hardened images, documentation |
| **P1 Mattermost** | https://chat.il2.dso.mil | Real-time chat with P1 team and community |
| **DSO Login** | https://login.dso.mil | SSO registration for P1 services |
| **Big Bang Docs** | https://docs-bigbang.dso.mil | Big Bang deployment documentation |
| **DISA STIGs** | https://public.cyber.mil/stigs/ | STIG downloads |
| **NIST OSCAL** | https://pages.nist.gov/OSCAL/ | OSCAL standard documentation |
| **NIST CMVP** | https://csrc.nist.gov/projects/cryptographic-module-validation-program | FIPS certificate lookup |

### 9.2 Mattermost Channels (Primary Communication)

Platform One's Mattermost is the most active communication channel. Key channels:

| Channel | Purpose |
|---------|---------|
| `#iron-bank` | Iron Bank questions, submission help |
| `#iron-bank-help` | Technical support for image hardening |
| `#big-bang` | Big Bang deployment and integration |
| `#big-bang-dev` | Big Bang development discussions |
| `#party-bus` | Party Bus program discussions |
| `#general` | General Platform One questions |
| `#newbie` | New to Platform One -- good starting point |

**To join:** Register at https://login.dso.mil, then access Mattermost at https://chat.il2.dso.mil

### 9.3 Key Teams and Roles

| Team | Responsibility | How to Reach |
|------|---------------|-------------|
| **Iron Bank Team** | Image hardening review and approval | `#iron-bank` on Mattermost, or issues on Repo One |
| **Big Bang Team** | Big Bang framework, addon reviews | `#big-bang` on Mattermost |
| **Party Bus Team** | Tool evaluation and onboarding | `#party-bus` on Mattermost |
| **Platform One DevSecOps** | General platform questions | `#general` on Mattermost |

### 9.4 Events and Networking

| Event | Description | Frequency |
|-------|-----------|-----------|
| **Platform One Demo Days** | Vendors demo tools to DoD audience | Quarterly |
| **AFWERX** | Air Force innovation hub | Ongoing |
| **DIU (Defense Innovation Unit)** | DoD tech scouting | Ongoing |
| **CDAO/JAIC events** | DoD AI/ML focused | Periodic |
| **SBIR/STTR** | Small business innovation research | Annual solicitations |

### 9.5 Getting an Account Without a CAC

If you do not have a Common Access Card (military/DoD civilian):

1. **ECA Certificate** -- Get from IdenTrust (https://www.identrust.com/certificates/eca), costs ~$100/year. This is the most common path for commercial vendors.
2. **Platform One SSO** -- Some P1 services accept SSO registration without a CAC. Check https://login.dso.mil.
3. **DoD Sponsor** -- A DoD employee can sponsor your access to certain P1 resources.

---

## 10. Success Stories

### 10.1 Small/Mid-Size Companies in Iron Bank and Platform One

While Platform One does not prominently advertise company sizes, these examples demonstrate that non-enterprise companies have succeeded:

| Company | Product | Size at Entry | Path | Notes |
|---------|---------|--------------|------|-------|
| **Anchore** | Container security scanning | ~50-100 employees | Iron Bank + Big Bang core | Now a core Big Bang component; started as a small startup |
| **Mattermost** | Team messaging | ~100-200 employees | Iron Bank + Big Bang addon | Replaced Slack for DoD; open-source core helped |
| **Gitea** | Git hosting | Open-source project (small team) | Iron Bank image | Community-contributed hardened image |
| **NeuVector** | Container runtime security | ~50 employees (pre-SUSE acquisition) | Iron Bank + Big Bang addon | Entered P1 as a small company, later acquired by SUSE |
| **Loki/Grafana** | Observability | Open-source (Grafana Labs ~500) | Iron Bank + Big Bang core | Open-source projects can be submitted by community |
| **MinIO** | Object storage | ~100-200 employees | Iron Bank + Big Bang addon | S3-compatible storage for air-gapped environments |

### 10.2 Key Patterns from Successful Entries

1. **Open-source advantage:** OSS projects have an easier path because the P1 community can contribute hardening work. Aegis could benefit from open-sourcing the core platform.
2. **DoD champion:** Every successful entry had at least one DoD program that wanted the tool. The champion provides justification and tests the deployment.
3. **Self-service hardening:** Companies that did their own Iron Bank hardening work (rather than waiting for P1 team) moved faster.
4. **Fills a gap:** Tools that provide capabilities not already available in the P1 ecosystem get prioritized. "Multi-cluster GPU orchestration" is a gap worth highlighting.
5. **Responsive maintenance:** Companies that quickly responded to CVE findings built trust with the Iron Bank team.

### 10.3 Relevance to Aegis

Aegis has a potentially strong position because:

- **GPU/AI workload orchestration is in high demand** across DoD (CDAO, service AI centers, national labs)
- **Multi-cluster management for Kubernetes** is a recognized gap in Big Bang
- **The hub-spoke architecture** aligns with DoD's multi-classification-level deployment model (IL2 hub managing IL4/IL5 spokes)
- **Self-hosted model** reduces the authorization boundary (customer controls infrastructure)
- **Existing Keycloak integration** aligns with Big Bang's SSO approach

---

## 11. Estimated Timeline

### 11.1 Realistic Timeline: Start to Iron Bank Images

| Phase | Duration | Activities |
|-------|----------|-----------|
| **Month 1-2: Preparation** | 8 weeks | Get Repo One account, study Iron Bank requirements, audit current images, begin hardening |
| **Month 3-4: Image Hardening** | 8 weeks | Rebuild on Iron Bank base images, fix CVEs, implement FIPS, non-root, minimal packages, SBOM |
| **Month 5: Submission** | 2-4 weeks | Submit hardening request, provide Dockerfile and docs, respond to initial feedback |
| **Month 6-8: Review Cycle** | 8-12 weeks | Iron Bank team reviews, automated scanning, back-and-forth on findings |
| **Month 9: Approval** | 2-4 weeks | Final review, image published to registry1.dso.mil |

**Total: 6-9 months** from start to published Iron Bank images.

### 11.2 Realistic Timeline: Big Bang Addon (Overlaps with Iron Bank)

| Phase | Duration | Activities |
|-------|----------|-----------|
| **Month 4-5: Chart Development** | 4-6 weeks | Refactor Helm chart for Big Bang compatibility, Istio integration |
| **Month 6-7: Testing** | 4-6 weeks | Test on Big Bang cluster (ideally RKE2), fix integration issues |
| **Month 8-9: Submission** | 4-6 weeks | Submit to Big Bang repo, review cycle |

**Total: 4-6 months**, but overlaps with Iron Bank work (start at month 4).

### 11.3 End-to-End Timeline

```
Month:  1    2    3    4    5    6    7    8    9    10   11   12
        |----|----|----|----|----|----|----|----|----|----|----|----|
Prep    [========]
Harden       [==============]
IB Submit              [====]
IB Review                   [==================]
IB Approved                                     [====]
BB Chart          [==============]
BB Test                     [===========]
BB Submit                             [===========]
BB Merged                                        [========]
cATO work   [================================================------>
```

**Month 9-10:** Iron Bank images available
**Month 10-12:** Big Bang addon merged
**Ongoing:** cATO support and vulnerability maintenance

### 11.4 Accelerators

| Action | Time Saved | How |
|--------|-----------|-----|
| Already using UBI9 base image | 2-4 weeks | Less base image rework needed |
| Already have NetworkPolicies | 1-2 weeks | Less Big Bang adaptation |
| Already have SBOM tooling | 1-2 weeks | Faster compliance documentation |
| Already have FIPS flags in spoke charts | 1-2 weeks | Less FIPS integration work |
| Get Iron Bank team feedback early | 2-4 weeks | Iterate before formal submission |
| Have a DoD customer champion | Significant | Prioritized review, clear justification |

---

## 12. Aegis-Specific Action Plan

### 12.1 Immediate Actions (This Month)

| Action | Effort | Notes |
|--------|--------|-------|
| Register for Repo One account (ECA cert from IdenTrust) | 1 day + cert processing time | ~$100/year for ECA cert |
| Join Platform One Mattermost, introduce yourself in `#iron-bank` and `#newbie` | 1 hour | Ask about current vendor submission process |
| Audit all container images for root user, SUID binaries, unnecessary packages | 1 day | Document findings |
| Run Trivy/Grype on all 4 images, document all Critical/High CVEs | 2 hours | `trivy image carlosmsanchez/aegis-platform-api:dev` |

### 12.2 Image Hardening (Month 1-3)

| Image | Current Base | Iron Bank Target Base | Key Changes |
|-------|-------------|----------------------|-------------|
| platform-api | `ubi9-minimal:9.5` (from registry.access.redhat.com) | `registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal` | Remove Pulumi from IB image (separate concern), add non-root USER, sign image |
| proxy | Check current | `registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal` | Same hardening pattern |
| k8s-agent | Check current | `registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal` | Same hardening pattern |
| workspace-vscode | Check current | Needs evaluation -- VS Code server has many deps | Most complex; may need extensive work |

**Critical decision for platform-api:** The current Dockerfile bundles Pulumi CLI, AWS CLI, kubectl, and jq. For Iron Bank, you likely need to split this into:

1. A lean `platform-api` image (just the Go binary + minimal deps)
2. A separate `platform-api-provisioner` image (with Pulumi, AWS CLI, kubectl)

Iron Bank will flag the large attack surface of the combined image.

### 12.3 FIPS Implementation (Month 2-3)

| Task | Description | Effort |
|------|------------|--------|
| Enable `GODEBUG=fips140=only` in ALL Go service deployments | Add env var to Helm values (hub + spoke) | 1 hour |
| Test all services with FIPS enabled | Run integration tests, check for SHA-1 usage that breaks | 1-2 days |
| Fix Pulumi SHA-1 issue | Pulumi uses SHA-1 for URNs; separate provisioner image or workaround | 1-2 days |
| Enable FIPS crypto policy in UBI9 base | `update-crypto-policies --set FIPS` in Dockerfile | 1 hour |
| Document cryptographic inventory | Table of all crypto usage with CMVP cert references | 1 day |
| Configure Keycloak FIPS mode | `KC_FIPS_MODE=strict` with FIPS-enabled JDK | 1 day |

### 12.4 Big Bang Compatibility (Month 3-5)

| Task | Description | Effort |
|------|------------|--------|
| Add Istio-compatible port names | Rename ports to `grpc-api`, `http-metrics`, etc. | 2 hours |
| Test with Istio sidecar injection | Deploy on a cluster with Istio, verify all traffic works | 2-3 days |
| Resolve mTLS conflict | Aegis has its own mTLS (step-ca); document how it coexists with Istio mTLS or defer to Istio | 1-2 days |
| Add ServiceMonitor for Prometheus | Helm template for Prometheus Operator | 2 hours |
| Create Grafana dashboard JSON | Dashboard for Aegis platform metrics | 1 day |
| Implement Big Bang values interface | `istio.enabled`, `monitoring.enabled`, `networkPolicies.enabled`, `sso.enabled` | 1-2 days |
| Add Istio VirtualService templates | Ingress configuration for Istio | 4 hours |
| Add Big Bang SSO integration | Configure Aegis to use Big Bang's Keycloak instance | 2-3 days |
| Test on RKE2 | Big Bang's preferred K8s distribution | 1-2 days |
| Write Big Bang-specific tests | Helm lint, template validation, integration tests | 1-2 days |

### 12.5 Documentation (Month 4-5)

| Task | Description | Effort |
|------|------------|--------|
| Complete STIG self-assessment checklist | Map each applicable STIG control to Aegis config | 2-3 days |
| Create OSCAL component definition | Machine-readable control implementation | 1-2 days |
| Finalize Customer Responsibility Matrix | Update existing CRM for DoD context | 1 day |
| Create DoD-specific hardening guide | Adapt existing guide for Big Bang deployment | 1-2 days |
| Document architecture diagrams | Network flows, data flows, auth flows for DoD evaluators | 1 day |

### 12.6 Estimated Costs

| Item | Cost | When |
|------|------|------|
| ECA Certificate (IdenTrust) | ~$100/year | Month 1 |
| Development time (your time) | Opportunity cost only | Months 1-9 |
| Test infrastructure (Big Bang on RKE2/EKS) | $200-500/month | Months 4-8 |
| Image signing (cosign key management) | $0 (use Sigstore/Fulcio) | Month 3 |
| **Total out-of-pocket** | **~$1,500-$4,500** | Over 9 months |

The major cost is your time -- estimated at 40-60% of your work hours for 2-3 months during the hardening phase, then 10-20% ongoing for maintenance.

---

## Appendix A: Aegis Existing Compliance Assets That Apply

The following existing Aegis compliance work directly supports Platform One / Iron Bank entry:

| Asset | Location | Relevance |
|-------|----------|-----------|
| UBI9 base image | `services/platform-api/Dockerfile` | Already using Iron Bank-compatible base |
| NetworkPolicies | `charts/aegis-services/templates/networkpolicy-*.yaml` | Big Bang compatibility |
| SBOM generation | `docs/compliance/sbom/` | Iron Bank requirement |
| FIPS runtime flag (spoke) | `charts/aegis-spoke/templates/*.yaml` | GODEBUG=fips140=only already in spoke charts |
| Hardening guide | `docs/compliance/customer-docs/configuration-hardening-guide.md` | STIG mapping foundation |
| Customer Responsibility Matrix | `docs/compliance/customer-docs/customer-responsibility-matrix.md` | cATO documentation |
| FedRAMP gap analysis | `docs/compliance/fedramp/02-gap-analysis.md` | Overlapping control requirements |
| FedRAMP scope assessment | `docs/compliance/fedramp/00-scope-and-impact-assessment.md` | Authorization boundary defined |
| SOC 2 policies | `docs/compliance/soc2/policies/` | Security policy foundation |
| OSCAL catalogs | `docs/oscal/` | Machine-readable control references |
| Structured logging | Platform API observability | Audit trail for STIG AU controls |
| Control implementation statements | `docs/compliance/customer-docs/control-implementation-statements.md` | Maps controls to Aegis features |

## Appendix B: Glossary

| Term | Definition |
|------|-----------|
| **ATO** | Authority to Operate -- formal authorization to run a system |
| **cATO** | Continuous ATO -- ongoing authorization via automated monitoring |
| **CAC** | Common Access Card -- DoD smart card for identity/authentication |
| **CMVP** | Cryptographic Module Validation Program (NIST) |
| **DISA** | Defense Information Systems Agency |
| **DSOP** | DoD Secure Open-source Project (now largely merged into Iron Bank process) |
| **ECA** | External Certificate Authority -- cert for non-DoD personnel to access DoD systems |
| **IL2/IL4/IL5** | Impact Levels -- DoD data classification (2=public, 4=CUI, 5=mission critical) |
| **Iron Bank** | DoD's hardened container image registry |
| **OSCAL** | Open Security Controls Assessment Language (NIST machine-readable format) |
| **P1** | Platform One |
| **Repo One** | DoD's GitLab instance hosting P1 source code |
| **RKE2** | Rancher Kubernetes Engine 2 -- Big Bang's preferred K8s distribution |
| **STIG** | Security Technical Implementation Guide (DISA) |
| **UBI** | Universal Base Image (Red Hat) |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-08 | Carlos Sanchez | Comprehensive rewrite with Aegis-specific analysis |
