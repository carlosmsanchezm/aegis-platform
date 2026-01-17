# Aegis Configuration Hardening Guide

**Version:** 1.0.0-DRAFT
**Last Updated:** [DATE]
**Classification:** Customer Confidential

## Overview

This guide provides secure configuration instructions for deploying Aegis Platform in compliance-sensitive environments. Following these settings ensures the platform meets FedRAMP Moderate, DoD IL-4, and NIST 800-171 requirements.

## Quick Reference

| Setting | Default | Hardened | Control |
|---------|---------|----------|---------|
| TLS Version | 1.2+ | 1.2+ (no 1.0/1.1) | SC-8 |
| Session Timeout | 60 min | 15 min | AC-11 |
| MFA Required | false | true | IA-2(1) |
| FIPS Mode | false | true | SC-13 |
| Audit Log Level | INFO | AUDIT | AU-2 |
| Token TTL | 300s | 300s (max) | AC-12 |

## 1. Cryptographic Configuration

### 1.1 Enable FIPS Mode

FIPS 140-2/3 validated cryptography is required for federal deployments.

#### Platform API (Go)

```yaml
# Helm values.yaml
platformApi:
  image:
    # Use FIPS-enabled image
    repository: your-registry/aegis-platform-api-fips
    tag: "1.0.0-fips"
  env:
    # Force FIPS-only cipher suites
    GODEBUG: "fips140=only"
```

#### Backstage (Node.js)

```yaml
# Helm values.yaml
backstage:
  image:
    # Use FIPS-enabled Node.js image
    repository: your-registry/aegis-backstage-fips
    tag: "1.0.0-fips"
  env:
    # Enable FIPS mode
    OPENSSL_FIPS: "1"
    NODE_OPTIONS: "--openssl-legacy-provider"
```

#### Verification

```bash
# Verify Go FIPS mode
kubectl exec -it deployment/platform-api -- \
  go run -fips140 -c 'import "crypto/fips140"; print(fips140.Enabled())'

# Verify Node.js FIPS mode
kubectl exec -it deployment/backstage -- \
  node -e "console.log(require('crypto').getFips())"
```

### 1.2 TLS Configuration

#### Ingress Controller

```yaml
# Ingress annotations for NGINX
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    # Minimum TLS 1.2
    nginx.ingress.kubernetes.io/ssl-protocols: "TLSv1.2 TLSv1.3"
    # FIPS-approved cipher suites only
    nginx.ingress.kubernetes.io/ssl-ciphers: |
      ECDHE-RSA-AES256-GCM-SHA384:
      ECDHE-RSA-AES128-GCM-SHA256:
      DHE-RSA-AES256-GCM-SHA384:
      DHE-RSA-AES128-GCM-SHA256
    # HSTS
    nginx.ingress.kubernetes.io/hsts: "true"
    nginx.ingress.kubernetes.io/hsts-max-age: "31536000"
```

#### Spoke Proxy

```yaml
# Helm values.yaml
spokeProxy:
  tls:
    minVersion: "1.2"
    cipherSuites:
      - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
      - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```

## 2. Authentication Configuration

### 2.1 Enable MFA Enforcement

```yaml
# Helm values.yaml
platformApi:
  env:
    # Require phishing-resistant MFA
    REQUIRE_PHISHING_RESISTANT_MFA: "true"
    # Allowed AMR values (adjust per your IdP)
    ALLOWED_AMR_VALUES: "mfa,otp,hwk"
```

### 2.2 Keycloak Hardening

```yaml
# Keycloak realm settings (apply via Keycloak API or console)
realm:
  # Password policy
  passwordPolicy: |
    length(12) and
    upperCase(1) and
    lowerCase(1) and
    digits(1) and
    specialChars(1) and
    notUsername and
    passwordHistory(12)

  # Brute force protection
  bruteForceProtected: true
  permanentLockout: false
  maxFailureWaitSeconds: 900
  minimumQuickLoginWaitSeconds: 60
  waitIncrementSeconds: 60
  quickLoginCheckMilliSeconds: 1000
  maxDeltaTimeSeconds: 43200
  failureFactor: 5

  # Session settings
  ssoSessionIdleTimeout: 900        # 15 minutes
  ssoSessionMaxLifespan: 36000      # 10 hours
  accessTokenLifespan: 300          # 5 minutes
  accessTokenLifespanForImplicitFlow: 300
```

### 2.3 CAC/PIV Configuration

```yaml
# Keycloak X.509 authentication
authentication:
  flows:
    browser:
      executions:
        - authenticator: auth-x509-client-username-form
          requirement: ALTERNATIVE
          config:
            x509-cert-auth.mapping-source-selection: "Subject's Common Name"
            x509-cert-auth.user-attribute-mapper: "userprincipalname"
            x509-cert-auth.crl-checking-enabled: "true"
            x509-cert-auth.ocsp-checking-enabled: "true"
```

## 3. Session Management

### 3.1 Session Timeout Configuration

```yaml
# Helm values.yaml
platformApi:
  env:
    # Maximum session idle time (seconds)
    SESSION_IDLE_TIMEOUT: "900"  # 15 minutes

    # Connection token TTL (max 300 seconds)
    AEGIS_PROXY_TOKEN_TTL_SECONDS: "300"

    # Enable one-time tokens
    AEGIS_PROXY_ONE_TIME_TOKENS: "true"

backstage:
  env:
    # Session configuration
    SESSION_SECRET: "${SESSION_SECRET}"  # From Kubernetes secret
    SESSION_MAX_AGE: "900000"  # 15 minutes in ms
```

### 3.2 Graceful Session Handling

```yaml
# Enable SUSPENDED state instead of FAILED on timeout
platformApi:
  env:
    # Workloads go to SUSPENDED not FAILED on idle timeout
    WORKLOAD_IDLE_STATE: "SUSPENDED"
    # Time before SUSPENDED workload is cleaned up
    SUSPENDED_CLEANUP_HOURS: "24"
```

## 4. Audit Logging Configuration

### 4.1 Enable Full Audit Logging

```yaml
# Helm values.yaml
platformApi:
  env:
    # Log level for security events
    LOG_LEVEL: "info"
    AUDIT_LOG_LEVEL: "debug"

    # Structured JSON logging
    LOG_FORMAT: "json"

    # Include all required fields
    AUDIT_INCLUDE_REQUEST_BODY: "false"  # PII consideration
    AUDIT_INCLUDE_RESPONSE_BODY: "false"

    # Log destinations
    AUDIT_LOG_STDOUT: "true"

  # Sidecar for log forwarding
  sidecars:
    - name: fluent-bit
      image: fluent/fluent-bit:latest
      volumeMounts:
        - name: varlog
          mountPath: /var/log
```

### 4.2 Log Retention Configuration

```yaml
# Fluent Bit configuration for 90+ day retention
fluent-bit:
  outputs:
    - name: s3
      match: audit.*
      bucket: aegis-audit-logs
      region: us-gov-west-1
      s3_key_format: /audit/$TAG/%Y/%m/%d/%H_%M_%S
      total_file_size: 100M
      upload_timeout: 60s
      # Lifecycle policy handles 90-day retention
```

### 4.3 SIEM Integration

```yaml
# Splunk HEC integration
fluent-bit:
  outputs:
    - name: splunk
      match: "*"
      host: splunk-hec.your-domain.com
      port: 8088
      splunk_token: "${SPLUNK_HEC_TOKEN}"
      tls: on
      tls.verify: on
```

## 5. Network Security

### 5.1 Network Policies

```yaml
# Default deny all ingress
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-ingress
  namespace: aegis-system
spec:
  podSelector: {}
  policyTypes:
    - Ingress

---
# Allow only required traffic
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-platform-api
  namespace: aegis-system
spec:
  podSelector:
    matchLabels:
      app: platform-api
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: backstage
        - podSelector:
            matchLabels:
              app: k8s-agent
      ports:
        - protocol: TCP
          port: 8443
```

### 5.2 Egress Restrictions

```yaml
# Restrict egress to required destinations
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: restrict-egress
  namespace: aegis-workloads
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    # Allow DNS
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system
      ports:
        - protocol: UDP
          port: 53
    # Allow container registry
    - to:
        - ipBlock:
            cidr: 10.0.0.0/8  # Internal registry
      ports:
        - protocol: TCP
          port: 443
```

### 5.3 Tenant Isolation

```yaml
# Helm values.yaml
tenantIsolation:
  enabled: true
  # Each project gets its own namespace
  namespacePerProject: true
  # Network policies auto-created
  autoNetworkPolicy: true
  # Resource quotas per tenant
  defaultResourceQuota:
    requests.cpu: "10"
    requests.memory: "32Gi"
    limits.nvidia.com/gpu: "4"
```

## 6. PKI Configuration

### 6.1 step-ca Integration

```yaml
# Helm values.yaml
pki:
  enabled: true
  provider: step-ca

stepCa:
  url: "https://step-ca.aegis-system.svc:9000"
  rootCA: "/etc/ssl/certs/step-root-ca.crt"
  provisioner:
    name: "aegis-provisioner"
    type: "JWK"
    # Key in Kubernetes secret
    keySecret: step-ca-provisioner-key
```

### 6.2 cert-manager Integration

```yaml
# ClusterIssuer for internal certificates
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: aegis-internal-ca
spec:
  ca:
    secretName: aegis-ca-key-pair

---
# Certificate for Platform API
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: platform-api-tls
  namespace: aegis-system
spec:
  secretName: platform-api-tls
  duration: 2160h    # 90 days
  renewBefore: 360h  # 15 days
  issuerRef:
    name: aegis-internal-ca
    kind: ClusterIssuer
  commonName: platform-api.aegis-system.svc
  dnsNames:
    - platform-api.aegis-system.svc
    - platform-api.aegis-system.svc.cluster.local
```

## 7. Complete Hardened values.yaml

```yaml
# Complete hardened deployment values
global:
  fipsEnabled: true
  tlsMinVersion: "1.2"

platformApi:
  image:
    repository: your-registry/aegis-platform-api-fips
    tag: "1.0.0-fips"
  env:
    GODEBUG: "fips140=only"
    REQUIRE_PHISHING_RESISTANT_MFA: "true"
    SESSION_IDLE_TIMEOUT: "900"
    AEGIS_PROXY_TOKEN_TTL_SECONDS: "300"
    AEGIS_PROXY_ONE_TIME_TOKENS: "true"
    LOG_FORMAT: "json"
    AUDIT_LOG_LEVEL: "debug"
    WORKLOAD_IDLE_STATE: "SUSPENDED"

backstage:
  image:
    repository: your-registry/aegis-backstage-fips
    tag: "1.0.0-fips"
  env:
    OPENSSL_FIPS: "1"
    SESSION_MAX_AGE: "900000"

keycloak:
  extraEnv:
    - name: KC_FIPS_MODE
      value: "strict"

spokeProxy:
  tls:
    minVersion: "1.2"
    cipherSuites:
      - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
      - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256

networkPolicies:
  enabled: true
  defaultDeny: true

tenantIsolation:
  enabled: true
  namespacePerProject: true
  autoNetworkPolicy: true

pki:
  enabled: true
  provider: step-ca
  certManager:
    enabled: true
    renewBefore: 360h

auditLogging:
  enabled: true
  format: json
  level: debug
  retention: 90d
  siemIntegration:
    enabled: true
    type: splunk
```

## 8. Verification Checklist

Run these checks after deployment:

```bash
#!/bin/bash
# hardening-verification.sh

echo "=== Aegis Hardening Verification ==="

# 1. TLS Version
echo "Checking TLS version..."
openssl s_client -connect aegis.your-domain.com:443 -tls1_1 2>&1 | \
  grep -q "handshake failure" && echo "✅ TLS 1.1 disabled" || echo "❌ TLS 1.1 enabled"

# 2. FIPS Mode
echo "Checking FIPS mode..."
kubectl exec -it deployment/platform-api -n aegis-system -- \
  env | grep -q "GODEBUG=fips140=only" && echo "✅ FIPS mode enabled" || echo "❌ FIPS mode disabled"

# 3. MFA Enforcement
echo "Checking MFA enforcement..."
kubectl get configmap platform-api-config -n aegis-system -o jsonpath='{.data.REQUIRE_PHISHING_RESISTANT_MFA}' | \
  grep -q "true" && echo "✅ MFA enforced" || echo "❌ MFA not enforced"

# 4. Network Policies
echo "Checking network policies..."
kubectl get networkpolicies -n aegis-system | \
  grep -q "default-deny" && echo "✅ Default deny policy exists" || echo "❌ No default deny policy"

# 5. Audit Logging
echo "Checking audit logging..."
kubectl logs deployment/platform-api -n aegis-system --tail=10 | \
  jq -e '.level' > /dev/null && echo "✅ JSON logging enabled" || echo "❌ JSON logging disabled"

echo "=== Verification Complete ==="
```

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | TBD | TBD | Initial draft |
