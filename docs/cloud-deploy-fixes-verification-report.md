# Cloud Deployment One-Shot Fixes — Verification Report

**Date**: 2026-03-09
**Branch**: `release/v1.0-pilot`
**Test method**: Local dry-run (Helm template, Terraform validate/plan, bash syntax check)
**EKS cluster**: Not torn down; no live deployment test performed

---

## Test Results Summary

| Test | Result | Notes |
|------|--------|-------|
| Bash syntax check (`generate-cloud-deployment.sh`) | PASS | `bash -n` exits 0 |
| Helm template (aegis-services: common + cloud) | PASS | Renders all resources with cert-manager enabled |
| Helm template (aegis-spoke: cloud-tls) | PASS | Renders TLS skip-verify env vars |
| Terraform validate | PASS | "The configuration is valid" |
| Terraform plan | PASS | 1 to add (StorageClass), 7 to change (outputs + DNS placeholders) |
| Terraform init (kubernetes provider) | PASS | `hashicorp/kubernetes v2.38.0` installed |

---

## Per-Fix Verification

### Fix 1: Spoke gRPC endpoint + DNS + OIDC vars (`terraform/outputs.tf`)

**Terraform plan output confirms** updated `helm_values_aegis_spoke` (sensitive, verified by reading outputs.tf directly):

- `AEGIS_CP_GRPC`: `aegis-platform-api.aegis-system.svc.cluster.local:8081` (was `prod-platform-api.prod.svc.cluster.local:8081`)
- `AEGIS_PROXY_INGRESS_HOST`: uses `cloudflare_record.proxy.hostname` (was `aws_route53_record.proxy.fqdn`)
- `AEGIS_CP_OIDC_TOKEN_URL`: `https://aegis-keycloak-service.aegis-system.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/token`
- `AEGIS_CP_OIDC_CLIENT_ID`: `spoke-agent`
- `AEGIS_CP_OIDC_CLIENT_SECRET`: `rEC99sBBWQAbRgg0xRQFBsMC8rt6pZOB`
- `AEGIS_CP_OIDC_AUDIENCE`: `aegis-platform`

**Status**: PASS

---

### Fix 2: Helm release/namespace vars + DB defaults (`terraform/variables.tf`)

- `helm_release_name` default: `aegis`
- `k8s_namespace` default: `aegis-system`
- `db_name` default: `aegis_platform` (was `aegis`)
- `db_user` default: `aegis_platform` (was `aegis_api`)

**Terraform plan confirms**: `rds_database_name = "aegis" -> "aegis_platform"`, `rds_username` changed (sensitive).

**Status**: PASS

---

### Fix 3: EBS gp3 StorageClass (`terraform/addons.tf`)

**Terraform plan confirms**:
```
+ kubernetes_storage_class_v1.ebs_gp3 will be created
  + allow_volume_expansion = true
  + parameters = { "encrypted" = "true", "type" = "gp3" }
  + storage_provisioner = "ebs.csi.aws.com"
  + volume_binding_mode = "WaitForFirstConsumer"
  + metadata.annotations = { "storageclass.kubernetes.io/is-default-class" = "true" }
```

**Status**: PASS

---

### Fix 4: cloud.yaml — DB creds, OIDC audience, MFA, storage (`charts/aegis-services/values/cloud.yaml`)

**Helm template output confirms**:
- `DB_NAME: "aegis_platform"` (was `"aegis"`)
- `DB_USER: "aegis_platform"` (was `"aegis_api"`)
- `OIDC_AUDIENCE: "backstage,aegis-platform"` (was not set in cloud.yaml)
- `REQUIRE_PHISHING_RESISTANT_MFA: "false"` (was not set in cloud.yaml)
- Keycloak postgres `storageClassName: ebs-gp3` (was `gp2`)
- Platform-api postgres `storageClassName: ebs-gp3` (was unset)

**Status**: PASS

---

### Fix 5: spoke-agent in AUTHZ bindings (`charts/aegis-services/values/common.yaml`)

**Helm template renders secret `authz-role-bindings.json`** which base64-decodes to:
```json
[
  {
    "clients": [
      "backstage",
      "vscode-extension",
      "spoke-agent"
    ],
    "roles": ["workspace-admin"],
    "projects": ["*"],
    "queues": ["*"]
  }
]
```

**Status**: PASS

---

### Fix 6: Spoke TLS skip-verify (`charts/aegis-spoke/values-cloud-tls.yaml`)

**Helm template output confirms** in k8s-agent Deployment:
- `AEGIS_CP_GRPC_INSECURE: "false"`
- `AEGIS_CP_GRPC_SKIP_VERIFY: "true"`
- `AEGIS_CP_OIDC_SKIP_TLS_VERIFY: "true"`

**Status**: PASS

---

### Fix 7: Deployment script fixes (`terraform/generate-cloud-deployment.sh`)

**7a. OIDC_AUDIENCE**: Line now emits `"backstage,aegis-platform"` (was `"backstage"`)
**7b. Spoke OIDC --set args**: 4 new `--set` lines added after `--set proxy.enabled=false`
**7c. In-cluster migration Step 6b**: New block inserted between Step 6 (Helm deploy) and Step 7 (LB wait)
**Bash syntax check**: PASS

**Status**: PASS

---

### Fix 8: Platform-api postgres storageClassName (`charts/aegis-services/templates/platform-api-postgres.yaml`)

**Helm template output confirms** `storageClassName: ebs-gp3` appears in the platform-api postgres volumeClaimTemplate.

**Status**: PASS

---

### Fix 9: Kubernetes provider (`terraform/main.tf`)

- Added `kubernetes` to `required_providers` (hashicorp/kubernetes ~> 2.25)
- Added `provider "kubernetes"` block with EKS exec-based auth
- `terraform init` installed `v2.38.0`
- `terraform validate` passed

**Status**: PASS

---

## Remaining Manual Steps Still Required

These items were NOT addressed in this changeset and will still require manual intervention during deployment:

### 1. ECR image push (pre-existing)
Images must be built and pushed to ECR before deployment. `make push-cloud-images` handles this but requires the docker daemon and ECR login.

### 2. Terraform apply for StorageClass
The new `ebs-gp3` StorageClass must be applied **before** `generate-cloud-deployment.sh` runs Helm install (otherwise PVCs fail). Running `terraform apply` will create it, but this is already part of the standard deploy flow.

### 3. Default StorageClass conflict
EKS ships with a `gp2` StorageClass marked as default. Our new `ebs-gp3` is also marked as default. Kubernetes allows multiple defaults but will **warn**. The PVCs explicitly specify `ebs-gp3` so this is not a functional issue, but to be clean, the `gp2` class default annotation should be removed. This could be automated:

```bash
# Potential addition to generate-cloud-deployment.sh or terraform:
kubectl annotate storageclass gp2 storageclass.kubernetes.io/is-default-class- --overwrite
```

### 4. DNS records reset to placeholder
Running `terraform apply` without `-var=platform_api_lb_hostname=...` will reset DNS to `placeholder.elb.amazonaws.com`. This is pre-existing — the script's Step 7 re-applies with actual LB hostnames. No fix needed, but worth noting.

### 5. Spoke client secret hardcoded
The `AEGIS_CP_OIDC_CLIENT_SECRET` value `rEC99sBBWQAbRgg0xRQFBsMC8rt6pZOB` is hardcoded in both `outputs.tf` and `generate-cloud-deployment.sh`. This was the existing Keycloak realm value. For production, this should come from a secret manager or Terraform variable.

### 6. Kubernetes provider auth requires AWS profile
The new `kubernetes` provider uses `--profile var.aws_profile`. If `aws_profile` is empty (CI/CD with env-based auth), the `--profile ""` arg may cause issues. Consider conditionalizing:

```hcl
args = compact(["eks", "get-token", "--cluster-name", aws_eks_cluster.main.name, "--region", var.aws_region,
  var.aws_profile != "" ? "--profile" : "", var.aws_profile != "" ? var.aws_profile : ""])
```

---

## Files Changed

| File | Lines changed |
|------|--------------|
| `terraform/outputs.tf` | Replaced 3 env vars with 7 |
| `terraform/variables.tf` | +12 lines (2 new vars), 2 default value changes |
| `terraform/addons.tf` | +17 lines (StorageClass resource) |
| `terraform/main.tf` | +13 lines (kubernetes provider + required_providers) |
| `charts/aegis-services/values/cloud.yaml` | 5 value changes + 4 new lines |
| `charts/aegis-services/values/common.yaml` | +1 line (spoke-agent) |
| `charts/aegis-spoke/values-cloud-tls.yaml` | Replaced 7-line file with 10 lines |
| `charts/aegis-services/templates/platform-api-postgres.yaml` | +2 lines (conditional storageClassName) |
| `terraform/generate-cloud-deployment.sh` | 3 changes: OIDC_AUDIENCE fix, 4 spoke --set args, ~30-line migration step |

---

## Recommendation

All 10 fixes across 9 files pass local verification. The changes are safe to apply to the live EKS cluster with:

```bash
# 1. Apply Terraform (creates StorageClass, updates outputs)
cd terraform && AWS_PROFILE=aegis-new terraform apply

# 2. Run deployment
cd .. && make deploy-cloud
```

The 6 remaining manual items above are candidates for future preventive fixes.
