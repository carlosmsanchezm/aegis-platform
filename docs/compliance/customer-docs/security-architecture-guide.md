# Aegis Security Architecture Guide

**Version:** 1.0.0-DRAFT
**Last Updated:** [DATE]
**Classification:** Customer Confidential

## 1. Executive Summary

This document describes the security architecture of the Aegis Platform, providing the technical details customers need to include Aegis in their System Security Plan (SSP) and Authority to Operate (ATO) package.

## 2. System Overview

### 2.1 Purpose

Aegis Platform provides [BRIEF DESCRIPTION OF WHAT AEGIS DOES - e.g., GPU workload orchestration across Kubernetes clusters].

### 2.2 Deployment Models

| Model | Description | Authorization Boundary |
|-------|-------------|----------------------|
| Model B (Self-Hosted) | Customer deploys Aegis in their environment | Customer's ATO boundary |
| Model C (Managed) | Aegis deploys in customer's cloud account | Customer's ATO boundary |

## 3. System Architecture

### 3.1 High-Level Architecture (C4 - Context)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          CUSTOMER AUTHORIZATION BOUNDARY                      │
│                                                                               │
│   ┌─────────────┐         ┌─────────────────────────────────────┐           │
│   │             │         │         Aegis Platform               │           │
│   │    Users    │◄───────►│  ┌───────────┐    ┌───────────────┐ │           │
│   │   (Browser) │  HTTPS  │  │ Backstage │    │  Platform API │ │           │
│   │             │         │  │    UI     │◄──►│    (gRPC)     │ │           │
│   └─────────────┘         │  └───────────┘    └───────────────┘ │           │
│                           │         │                │          │           │
│   ┌─────────────┐         │         ▼                ▼          │           │
│   │   Keycloak  │◄────────│  ┌───────────────────────────────┐  │           │
│   │    (IdP)    │  OIDC   │  │     Spoke Clusters (K8s)      │  │           │
│   └─────────────┘         │  │  ┌─────────┐  ┌─────────────┐ │  │           │
│                           │  │  │K8s Agent│  │ GPU Workloads│ │  │           │
│                           │  │  └─────────┘  └─────────────┘ │  │           │
│                           │  └───────────────────────────────┘  │           │
│                           └─────────────────────────────────────┘           │
│                                                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Component Architecture (C4 - Container)

| Component | Technology | Purpose | Ports |
|-----------|-----------|---------|-------|
| Backstage UI | React/Node.js | User interface, catalog | 443 (HTTPS) |
| Platform API | Go/gRPC | Workload orchestration | 443 (gRPC-TLS) |
| Keycloak | Java | Identity provider | 443 (HTTPS) |
| PostgreSQL | PostgreSQL | Data persistence | 5432 (internal) |
| K8s Agent | Go | Cluster management | 443 (outbound) |
| Spoke Proxy | Go | Secure tunnel to hub | 443 (outbound) |

### 3.3 Component Details

#### 3.3.1 Backstage UI

**Purpose:** Web-based user interface for workload management

**Security Features:**
- OIDC authentication via Keycloak
- PKCE for token exchange
- Session management with configurable timeout
- CSP headers enforced

**Data Handled:**
- User session tokens (in-memory)
- Workload metadata (display only)

#### 3.3.2 Platform API

**Purpose:** Central orchestration service

**Security Features:**
- gRPC with mutual TLS
- JWT validation for all requests
- RBAC policy enforcement
- Structured audit logging

**Data Handled:**
- Workload definitions
- Cluster state
- Budget allocations
- Audit logs

#### 3.3.3 K8s Agent

**Purpose:** Runs in each spoke cluster, manages local workloads

**Security Features:**
- mTLS connection to hub
- Service account with minimal RBAC
- No inbound connections (outbound only)

#### 3.3.4 Spoke Proxy

**Purpose:** Secure tunnel for user connections to workloads

**Security Features:**
- TLS 1.2+ with FIPS cipher suites
- Short-lived connection tokens (5 min max)
- One-time use tokens available

## 4. Data Flow Diagrams

### 4.1 User Authentication Flow

```
┌──────┐     ┌───────────┐     ┌──────────┐     ┌─────────────┐
│ User │────►│ Backstage │────►│ Keycloak │────►│ IdP/LDAP    │
└──────┘     └───────────┘     └──────────┘     │ (Customer)  │
   │              │                  │          └─────────────┘
   │              │                  │
   │  1. Access   │  2. OIDC         │  3. Authenticate
   │     UI       │     Redirect     │     (MFA if configured)
   │              │                  │
   │              │◄─────────────────│
   │              │  4. ID Token     │
   │◄─────────────│     (JWT)        │
   │  5. Session  │                  │
   │     Cookie   │                  │
```

### 4.2 Workload Submission Flow

```
┌──────┐     ┌───────────┐     ┌─────────────┐     ┌──────────────┐
│ User │────►│ Backstage │────►│ Platform API│────►│ Spoke Cluster│
└──────┘     └───────────┘     └─────────────┘     └──────────────┘
                                     │
   1. Submit workload                │
   2. Validate JWT, check RBAC       │
   3. Check budget                   │
   4. Queue workload                 │
                                     │
                              5. K8s Agent polls ──────────────────►
                              6. Create Pod, attach GPU
                              7. Report status
```

### 4.3 Data at Rest

| Data Type | Storage Location | Encryption | Retention |
|-----------|-----------------|------------|-----------|
| User credentials | Keycloak DB | AES-256 | Per policy |
| Workload metadata | Platform API DB | AES-256 | 90 days |
| Audit logs | Platform API | TLS in transit | 90+ days |
| Container images | Customer registry | Customer controlled | Customer policy |

### 4.4 Data in Transit

| Flow | Protocol | Encryption | Authentication |
|------|----------|------------|----------------|
| User → Backstage | HTTPS | TLS 1.2+ | Session cookie |
| Backstage → Keycloak | HTTPS | TLS 1.2+ | OIDC |
| Backstage → Platform API | gRPC-TLS | TLS 1.2+ | JWT |
| Platform API → Spoke | mTLS | TLS 1.2+ | Client cert |
| User → Workload | SSH over TLS | TLS 1.2+ | Short-lived token |

## 5. Network Architecture

### 5.1 Network Boundaries

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CUSTOMER VPC                                    │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         PUBLIC SUBNET                                 │   │
│  │   ┌─────────────┐                                                     │   │
│  │   │    ALB      │  (HTTPS 443 only)                                  │   │
│  │   └──────┬──────┘                                                     │   │
│  └──────────┼────────────────────────────────────────────────────────────┘   │
│             │                                                                │
│  ┌──────────┼────────────────────────────────────────────────────────────┐   │
│  │          │              PRIVATE SUBNET (Hub)                          │   │
│  │   ┌──────▼──────┐    ┌─────────────┐    ┌─────────────┐              │   │
│  │   │  Backstage  │    │ Platform API│    │  Keycloak   │              │   │
│  │   └─────────────┘    └─────────────┘    └─────────────┘              │   │
│  │                             │                                         │   │
│  │   ┌─────────────────────────┴─────────────────────────┐              │   │
│  │   │                    PostgreSQL                      │              │   │
│  │   └────────────────────────────────────────────────────┘              │   │
│  └───────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐   │
│  │                     PRIVATE SUBNET (Spoke)                            │   │
│  │   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐              │   │
│  │   │  K8s Agent  │───►│ Spoke Proxy │    │ GPU Nodes   │              │   │
│  │   └─────────────┘    └─────────────┘    └─────────────┘              │   │
│  └───────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.2 Required Ingress Rules

| Source | Destination | Port | Protocol | Purpose |
|--------|-------------|------|----------|---------|
| Internet | ALB | 443 | HTTPS | User access |
| ALB | Backstage | 3000 | HTTP | Internal routing |
| ALB | Keycloak | 8443 | HTTPS | Auth endpoints |

### 5.3 Required Egress Rules

| Source | Destination | Port | Protocol | Purpose |
|--------|-------------|------|----------|---------|
| K8s Agent | Platform API | 443 | gRPC-TLS | Cluster registration |
| Spoke Proxy | Hub | 443 | TLS | Tunnel |
| All pods | Internet | 443 | HTTPS | Container registry |

## 6. Trust Boundaries

### 6.1 External Trust Boundary

- All traffic from internet terminates at ALB
- TLS termination with customer-provided certificates
- WAF rules applied (optional)

### 6.2 Internal Trust Boundaries

| Boundary | Controls |
|----------|----------|
| Namespace isolation | Kubernetes NetworkPolicy |
| Service-to-service | mTLS, service mesh (optional) |
| Database access | Network ACL, credential rotation |

## 7. Identity and Access Management

### 7.1 User Authentication

| Method | Description | Control Reference |
|--------|-------------|-------------------|
| OIDC/SAML | Primary authentication via Keycloak | IA-2 |
| MFA | Optional, configurable | IA-2(1) |
| CAC/PIV | Optional, via Keycloak adapter | IA-2(12) |

### 7.2 Service Authentication

| Service | Method | Credential |
|---------|--------|------------|
| K8s Agent | mTLS | Client certificate |
| Spoke Proxy | mTLS | Client certificate |
| Platform API | JWT | Keycloak-issued token |

### 7.3 Authorization Model

- RBAC enforced at Platform API level
- Roles: Admin, Project Owner, Developer, Viewer
- Permissions mapped to gRPC methods
- Audit log of all authorization decisions

## 8. Cryptographic Controls

### 8.1 Cryptographic Modules

| Use Case | Algorithm | Key Size | FIPS Mode |
|----------|-----------|----------|-----------|
| TLS | AES-GCM | 256-bit | Available |
| JWT Signing | RS256/ES256 | 2048/256 | Available |
| Password Hash | bcrypt | - | N/A |
| Disk Encryption | Customer controlled | Customer controlled | Customer choice |

### 8.2 FIPS 140-2/3 Compliance

Aegis supports FIPS mode when:
- Node.js built with FIPS OpenSSL
- Go compiled with BoringCrypto
- Container base image: UBI9 with FIPS enabled

See [Configuration Hardening Guide](./configuration-hardening-guide.md) for enablement.

## 9. Audit and Logging

### 9.1 Log Types

| Log Type | Content | Format | Retention |
|----------|---------|--------|-----------|
| Authentication | Login/logout, MFA events | JSON | 90+ days |
| Authorization | Access decisions, denials | JSON | 90+ days |
| Workload | Create, update, delete | JSON | 90+ days |
| System | Errors, performance | JSON | 30 days |

### 9.2 Log Fields (AU-3 Compliance)

All audit logs include:
- Timestamp (UTC, ISO 8601)
- Event type
- Subject identity
- Source IP
- Action performed
- Resource affected
- Outcome (success/failure)

### 9.3 SIEM Integration

Logs can be exported to customer SIEM via:
- Fluentd/Fluent Bit sidecar
- Direct Splunk/ELK integration
- CloudWatch Logs (AWS)
- Stackdriver (GCP)

## 10. Component Inventory

### 10.1 Container Images

| Image | Base | Version | Registry |
|-------|------|---------|----------|
| backstage | node:20-alpine | See SBOM | Customer ECR |
| platform-api | ubi9-minimal | See SBOM | Customer ECR |
| keycloak | ubi9-openjdk-17 | See SBOM | Customer ECR |
| k8s-agent | ubi9-minimal | See SBOM | Customer ECR |
| spoke-proxy | ubi9-minimal | See SBOM | Customer ECR |

### 10.2 Third-Party Dependencies

See [SBOM](../sbom/README.md) for complete dependency listing.

## 11. Appendices

### 11.1 Glossary

| Term | Definition |
|------|------------|
| Hub | Central control plane cluster |
| Spoke | Remote GPU cluster |
| Workload | User-submitted compute job |
| ATO | Authority to Operate |
| SSP | System Security Plan |

### 11.2 References

- NIST SP 800-53 Rev 5
- NIST SP 800-171 Rev 2
- FedRAMP Moderate Baseline
- DoD Cloud Computing SRG

### 11.3 Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | TBD | TBD | Initial draft |
