# Aegis Incident Response Runbook

**Version:** 1.0.0-DRAFT
**Last Updated:** [DATE]
**Classification:** Customer Confidential

## Overview

This runbook provides procedures for detecting, responding to, and recovering from security incidents in Aegis Platform deployments. Customers can incorporate these procedures into their Incident Response Plan.

## Incident Classification

### Severity Levels

| Severity | Definition | Response Time | Examples |
|----------|-----------|---------------|----------|
| **Critical (P1)** | Active breach, data exfiltration, system compromise | Immediate (< 15 min) | Unauthorized admin access, ransomware, data leak |
| **High (P2)** | Potential breach, significant vulnerability exploitation | < 1 hour | Brute force attack succeeding, privilege escalation |
| **Medium (P3)** | Security control failure, policy violation | < 4 hours | Failed MFA bypass attempts, config drift |
| **Low (P4)** | Minor security event, informational | < 24 hours | Repeated failed logins, scan activity |

### Incident Categories

| Category | Description | Example Events |
|----------|-------------|----------------|
| **Authentication** | Unauthorized access attempts | Brute force, credential stuffing, session hijacking |
| **Authorization** | Privilege violations | Unauthorized API calls, RBAC bypass |
| **Data** | Data confidentiality/integrity | Unauthorized data access, data modification |
| **Availability** | Service disruption | DDoS, resource exhaustion, system crash |
| **Configuration** | Security misconfigurations | Disabled controls, exposed secrets |
| **Malware** | Malicious code execution | Container compromise, cryptomining |

## Detection

### 1. Authentication Anomalies

#### Indicators

```json
// Failed authentication spike
{
  "event": "auth_failure",
  "count": "> 10 in 5 minutes from same source",
  "severity": "medium"
}

// Successful auth after multiple failures
{
  "event": "auth_success",
  "previous_failures": "> 5",
  "severity": "high"
}

// Auth from unusual location
{
  "event": "auth_success",
  "geo_anomaly": true,
  "severity": "medium"
}
```

#### Detection Query (Splunk)

```spl
index=aegis_audit event=auth_failure
| stats count by src_ip, user
| where count > 10
| sort -count
```

#### Detection Query (Prometheus Alert)

```yaml
groups:
  - name: aegis-security
    rules:
      - alert: AuthenticationBruteForce
        expr: |
          sum(rate(aegis_auth_failures_total[5m])) by (src_ip) > 0.5
        for: 2m
        labels:
          severity: high
        annotations:
          summary: "Brute force attack detected from {{ $labels.src_ip }}"
```

### 2. Authorization Violations

#### Indicators

```json
// Repeated RBAC denials
{
  "event": "authz_denied",
  "count": "> 5 for same user/resource",
  "severity": "medium"
}

// Admin API access by non-admin
{
  "event": "authz_denied",
  "resource": "admin/*",
  "severity": "high"
}
```

#### Detection Query

```spl
index=aegis_audit event=authz_denied
| stats count by user, resource
| where count > 5
```

### 3. Data Access Anomalies

#### Indicators

```json
// Large data export
{
  "event": "data_export",
  "size": "> normal baseline",
  "severity": "medium"
}

// Access to multiple projects
{
  "event": "project_access",
  "distinct_projects": "> 10 in 1 hour",
  "severity": "medium"
}
```

### 4. System Anomalies

#### Indicators

```json
// Unusual container execution
{
  "event": "container_exec",
  "command": "contains(shell, wget, curl, nc)",
  "severity": "high"
}

// Resource exhaustion
{
  "event": "resource_limit",
  "type": "GPU memory exhausted",
  "severity": "medium"
}
```

## Response Procedures

### Procedure 1: Authentication Breach Response

**Trigger:** Confirmed unauthorized access via compromised credentials

**Severity:** Critical (P1)

#### Immediate Actions (0-15 minutes)

1. **Isolate affected accounts**
   ```bash
   # Disable user in Keycloak
   kubectl exec -it deployment/keycloak -n aegis-system -- \
     /opt/keycloak/bin/kcadm.sh update users/<USER_ID> \
     -r aegis -s enabled=false

   # Revoke all active sessions
   kubectl exec -it deployment/keycloak -n aegis-system -- \
     /opt/keycloak/bin/kcadm.sh delete users/<USER_ID>/sessions \
     -r aegis
   ```

2. **Terminate active connection sessions**
   ```bash
   # List active sessions for user
   grpcurl -d '{"subject": "compromised-user@example.com"}' \
     platform-api:8443 aegis.v1.PlatformAPI/ListConnectionSessions

   # Revoke specific session
   grpcurl -d '{"session_id": "session-xxx"}' \
     platform-api:8443 aegis.v1.PlatformAPI/RevokeConnectionSession
   ```

3. **Preserve evidence**
   ```bash
   # Export audit logs for affected period
   kubectl logs deployment/platform-api -n aegis-system \
     --since=2h > /evidence/$(date +%Y%m%d_%H%M%S)_platform_api.log

   # Export Keycloak events
   kubectl exec -it deployment/keycloak -n aegis-system -- \
     /opt/keycloak/bin/kcadm.sh get events \
     -r aegis --fields time,type,userId,ipAddress \
     > /evidence/$(date +%Y%m%d_%H%M%S)_keycloak_events.json
   ```

#### Short-term Actions (15 min - 4 hours)

4. **Scope the incident**
   - Identify all resources accessed by compromised account
   - Determine lateral movement
   - Check for privilege escalation

5. **Notify stakeholders**
   - Security team
   - Affected project owners
   - Legal/compliance (if data breach)

6. **Implement additional controls**
   ```bash
   # Enable MFA enforcement if not already
   kubectl set env deployment/platform-api -n aegis-system \
     REQUIRE_PHISHING_RESISTANT_MFA=true
   ```

#### Recovery Actions (4-24 hours)

7. **Reset credentials**
   - Force password reset for affected user
   - Rotate any API tokens or service accounts

8. **Restore from known-good state**
   - If data modified, restore from backup
   - Verify integrity of restored data

9. **Document and report**
   - Complete incident report
   - Update risk register
   - Initiate post-incident review

### Procedure 2: Workload Compromise Response

**Trigger:** Malicious activity detected in customer workload

**Severity:** High (P2)

#### Immediate Actions

1. **Isolate the workload**
   ```bash
   # Terminate the workload immediately
   grpcurl -d '{"workload_id": "workload-xxx"}' \
     platform-api:8443 aegis.v1.PlatformAPI/TerminateWorkload

   # Apply network isolation
   kubectl apply -f - <<EOF
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: isolate-workload-xxx
     namespace: aegis-workloads
   spec:
     podSelector:
       matchLabels:
         workload-id: workload-xxx
     policyTypes:
       - Ingress
       - Egress
   EOF
   ```

2. **Preserve container state**
   ```bash
   # Checkpoint container for forensics (if supported)
   crictl checkpoint <container-id> --export=/evidence/container-xxx.tar

   # Or save container logs
   kubectl logs pod/workload-xxx-0 -n aegis-workloads --all-containers \
     > /evidence/workload-xxx-logs.txt
   ```

3. **Check for lateral movement**
   ```bash
   # Review network connections from compromised workload
   kubectl exec -it pod/workload-xxx-0 -n aegis-workloads -- \
     netstat -tuln > /evidence/workload-xxx-netstat.txt

   # Check for connections to other workloads or platform components
   ```

#### Containment Actions

4. **Review similar workloads**
   - Check other workloads from same project
   - Verify container image integrity
   - Scan for same indicators of compromise

5. **Update detection rules**
   - Add new IOCs to detection system
   - Block malicious IPs/domains

### Procedure 3: API Key/Token Leak Response

**Trigger:** Aegis API credentials found in public repository or logs

**Severity:** Critical (P1)

#### Immediate Actions

1. **Identify leaked credentials**
   - Determine credential type (JWT, service account, API key)
   - Identify scope/permissions
   - Determine exposure timeline

2. **Revoke immediately**
   ```bash
   # Revoke JWT (if applicable)
   # Force session invalidation in Keycloak

   # Rotate service account
   kubectl delete secret service-account-key -n aegis-system
   kubectl create secret generic service-account-key \
     --from-literal=key=$(openssl rand -hex 32) -n aegis-system

   # Restart affected services
   kubectl rollout restart deployment/platform-api -n aegis-system
   ```

3. **Audit usage**
   ```bash
   # Search for usage of leaked credentials
   # (Query depends on credential type)
   ```

### Procedure 4: DDoS/Availability Attack Response

**Trigger:** Service degradation due to attack

**Severity:** High (P2)

#### Immediate Actions

1. **Confirm attack vs. legitimate traffic**
   ```bash
   # Check request rates
   kubectl top pods -n aegis-system

   # Review traffic patterns
   # (WAF/load balancer logs)
   ```

2. **Enable rate limiting**
   ```yaml
   # Apply rate limiting via ingress
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: aegis-ingress
     annotations:
       nginx.ingress.kubernetes.io/limit-rps: "10"
       nginx.ingress.kubernetes.io/limit-connections: "5"
   ```

3. **Block attacking sources**
   ```bash
   # Add to WAF blocklist
   # (Specific to your WAF provider)
   ```

## Evidence Collection

### Required Evidence for All Incidents

| Evidence Type | Collection Method | Storage Location |
|--------------|-------------------|------------------|
| Audit logs | `kubectl logs` export | S3/secure storage |
| Keycloak events | Admin API export | S3/secure storage |
| Network captures | tcpdump/Wireshark | Encrypted volume |
| Container state | Checkpoint or logs | Encrypted volume |
| Configuration | Helm/kubectl export | Git (private repo) |

### Evidence Preservation Script

```bash
#!/bin/bash
# evidence-collection.sh

INCIDENT_ID=${1:-$(date +%Y%m%d_%H%M%S)}
EVIDENCE_DIR="/evidence/${INCIDENT_ID}"

mkdir -p ${EVIDENCE_DIR}

echo "Collecting evidence for incident: ${INCIDENT_ID}"

# Platform API logs
kubectl logs deployment/platform-api -n aegis-system --since=24h \
  > ${EVIDENCE_DIR}/platform-api.log

# Keycloak logs
kubectl logs deployment/keycloak -n aegis-system --since=24h \
  > ${EVIDENCE_DIR}/keycloak.log

# Current configuration
kubectl get all -n aegis-system -o yaml \
  > ${EVIDENCE_DIR}/cluster-state.yaml

# Network policies
kubectl get networkpolicies -A -o yaml \
  > ${EVIDENCE_DIR}/network-policies.yaml

# Secrets list (not contents)
kubectl get secrets -A -o custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name \
  > ${EVIDENCE_DIR}/secrets-inventory.txt

# Calculate hashes
find ${EVIDENCE_DIR} -type f -exec sha256sum {} \; \
  > ${EVIDENCE_DIR}/evidence-hashes.txt

echo "Evidence collected in ${EVIDENCE_DIR}"
echo "Hash file: ${EVIDENCE_DIR}/evidence-hashes.txt"
```

## Escalation Contacts

| Role | Contact | Escalation Trigger |
|------|---------|-------------------|
| On-call Engineer | [PAGE] | All P1/P2 incidents |
| Security Lead | [EMAIL] | Confirmed breaches |
| Legal/Privacy | [EMAIL] | Data breaches |
| Customer Success | [EMAIL] | Customer-impacting incidents |
| Executive | [PAGE] | Public disclosure required |

## Post-Incident Review

### Review Checklist

- [ ] Timeline of events documented
- [ ] Root cause identified
- [ ] All affected systems identified
- [ ] Evidence preserved with chain of custody
- [ ] Remediation actions completed
- [ ] Detection improvements identified
- [ ] Runbook updates needed
- [ ] Training needs identified
- [ ] Customer notification completed (if required)
- [ ] Regulatory notification completed (if required)

### Report Template

```markdown
# Incident Report: [INCIDENT_ID]

## Executive Summary
[2-3 sentence summary]

## Timeline
| Time (UTC) | Event |
|-----------|-------|
| YYYY-MM-DD HH:MM | Initial detection |
| ... | ... |

## Impact
- Systems affected:
- Data affected:
- Users affected:
- Duration:

## Root Cause
[Description of root cause]

## Response Actions
1. [Action taken]
2. [Action taken]

## Remediation
- [ ] Short-term fix applied
- [ ] Long-term fix planned
- [ ] Detection improved

## Lessons Learned
1. [Lesson]
2. [Lesson]

## Recommendations
1. [Recommendation]
2. [Recommendation]
```

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | TBD | TBD | Initial draft |
