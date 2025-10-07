# Aegis Auth Proxy

The `aegis-auth-proxy` service provides secure, authenticated tunneling to backend workloads. It enforces JWT-based authentication, token replay prevention, session timeouts, and client certificate validation per FedRAMP compliance requirements (AC-17, IA-2, SC-10, SC-13).

## Overview

The proxy accepts incoming CONNECT and WebSocket connections from authenticated clients, validates JWT tokens with workload-specific claims, and establishes secure tunnels to backend services. All connection attempts and session lifecycle events are logged for audit purposes.

## Readiness Metrics

### Endpoint

The proxy exposes a readiness endpoint at `/ready` (or via the `ReadinessHandler` method) that performs individual health checks and returns:
- **HTTP 200** if all checks pass
- **HTTP 503** if any check fails

### Metrics

The proxy exports Prometheus metrics to track the health and performance of readiness checks.

#### `proxy.readiness.check_duration_seconds`

**Type:** Histogram  
**Labels:**
- `check_name` - The name of the individual readiness check (e.g., `jti_store`, `logger`, `verifier`)

**Purpose:**  
Tracks the duration of each readiness check execution in seconds. This metric helps operators monitor the health of critical proxy components and detect performance degradation.

**Buckets:**  
`[0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0]` seconds

### Individual Checks

The readiness endpoint validates the following components:

1. **`jti_store`** - Verifies the JWT ID (JTI) store is initialized and operational. This store prevents token replay attacks by tracking used tokens.

2. **`logger`** - Validates the structured logger (zap) is initialized and ready to accept log events.

3. **`verifier`** - Confirms the JWT verifier is initialized with the correct signing key and audience configuration.

Each check is executed sequentially, and both success and failure results are logged with structured fields including check name, duration, and status.

### Prometheus Queries

**Average check duration by check name:**
```promql
rate(proxy.readiness.check_duration_seconds_sum[5m]) / rate(proxy.readiness.check_duration_seconds_count[5m])
```

**95th percentile check duration:**
```promql
histogram_quantile(0.95, rate(proxy.readiness.check_duration_seconds_bucket[5m]))
```

**Check execution rate:**
```promql
rate(proxy.readiness.check_duration_seconds_count[5m])
```

### Alerting

**Example: Alert when readiness checks are slow**
```yaml
alert: ProxyReadinessCheckSlow
expr: histogram_quantile(0.95, rate(proxy.readiness.check_duration_seconds_bucket[5m])) > 0.5
for: 5m
labels:
  severity: warning
annotations:
  summary: "Proxy readiness checks are slow"
  description: "The 95th percentile check duration is {{ $value }}s, exceeding 0.5s threshold."
```

**Example: Alert when checks are failing (inferred from logs or custom metrics)**
```yaml
alert: ProxyReadinessCheckFailing
expr: up{job="aegis-auth-proxy"} == 0
for: 2m
labels:
  severity: critical
annotations:
  summary: "Proxy is not ready"
  description: "The aegis-auth-proxy readiness checks have been failing for 2 minutes."
```

## Configuration

The proxy is configured via environment variables and command-line flags. Key configuration includes:

- **JWT Secret:** Shared secret for validating JWT tokens
- **Expected Audience:** Required `aud` claim in JWT tokens
- **Destination Suffix:** Allowed suffix for backend destinations (e.g., `.cluster.local`)
- **Session Timeouts:** Inactivity timeouts for standard and privileged sessions
- **TLS Certificates:** Server certificate and key for mTLS

Refer to the `internal/server/config.go` file for the complete configuration structure.

## Deployment

The proxy should be deployed as a Kubernetes Deployment with:
- Readiness probe configured to call the readiness endpoint
- Liveness probe to detect unrecoverable failures
- Prometheus annotations for metric scraping
- Resource limits to prevent resource exhaustion

Example readiness probe:
```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8443
    scheme: HTTPS
  initialDelaySeconds: 5
  periodSeconds: 10
```

## Security Considerations

- All connections require valid JWT tokens
- JTI store prevents token replay within the configured TTL
- Session inactivity timeouts enforce automatic disconnection
- Client certificate SANs can be restricted via allow-list
- All session events are logged for audit trails
