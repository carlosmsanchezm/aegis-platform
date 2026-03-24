# Dev Mode vs Production Mode

This document explains the differences between Aegis dev mode (`AEGIS_DEV_MODE=true`) and production mode, and how to switch between them.

This is an application-mode document, not a deployment workflow document:
- Local and hybrid hub development should keep using the local/tunnel workflow.
- Full-cloud deployment should keep using `AGENT_DEPLOYMENT_GUIDE.md`.

---

## Mode Comparison

| Behavior | Dev Mode (`AEGIS_DEV_MODE=true`) | Production Mode (`AEGIS_DEV_MODE=false`) |
|----------|----------------------------------|------------------------------------------|
| AWS role assumption | Skipped -- ambient credentials used directly | Enforced -- STS AssumeRole with ExternalID |
| Project AWS fields required | Account ID only | Account ID + Role ARN + External ID |
| Multi-tenancy isolation | None -- all projects share one AWS identity | Full -- each project gets its own IAM role in its own account |
| Pulumi credential source | Pod's IRSA / env vars / `~/.aws/credentials` | AssumeRole into `aegis-project-{id}` in spoke account |
| CloudTrail attribution | All actions logged under hub identity | Actions logged under project-specific assumed role |
| Helm value | `AEGIS_DEV_MODE: "true"` | `AEGIS_DEV_MODE: "false"` (or omit) |
| Hardening profile | Typically `dev` | Typically `standard` |

---

## Switching from Dev to Production

### Prerequisites

- Hub account: IRSA role for platform-api already exists via `terraform/modules/platform-api-irsa/`
- Spoke/project account: needs the `spoke-project-role` IAM role (see below)

### Step-by-step Checklist

1. **Create the spoke project IAM role**

   Option A -- same account (pilot/dev):
   ```bash
   cd terraform/
   terraform apply \
     -var="create_spoke_project_role=true" \
     -var="spoke_project_id=e2e-pilot-test" \
     -var="hub_platform_api_role_arn=arn:aws:iam::195714074609:role/aegis-platform-api" \
     -var="spoke_external_id=aegis-pilot-2026"
   ```

   Option B -- separate spoke account:
   ```hcl
   module "spoke_project_role" {
     source = "git::https://github.com/yourorg/aegis-platform.git//terraform/modules/spoke-project-role"

     project_id                = "my-project"
     hub_platform_api_role_arn = "arn:aws:iam::195714074609:role/aegis-platform-api"
     external_id               = "aegis-my-project-2026"
   }
   ```

2. **Set `AEGIS_DEV_MODE: "false"`** in `charts/aegis-services/values/cloud.yaml`

3. **Optionally set AWS defaults** in Helm values (pre-fills the UI):
   ```yaml
   AEGIS_DEFAULT_AWS_ACCOUNT_ID: "123456789012"
   AEGIS_DEFAULT_AWS_ROLE_ARN: "arn:aws:iam::123456789012:role/aegis-project-e2e-pilot-test"
   AEGIS_DEFAULT_AWS_EXTERNAL_ID: "aegis-pilot-2026"
   ```

4. **Build/push first, then redeploy**
   ```bash
   make push-cloud-images   # if using ECR
   make deploy-cloud        # terraform apply (idempotent) + deploy
   ```

5. **Create project in UI** with the role ARN from terraform output + your chosen external ID

6. **Verify** in CloudTrail that Pulumi actions appear under the assumed role (`aegis-project-{id}`)

---

## AWS IAM Setup Reference

### Hub Account (platform-api IRSA)

Already managed by `terraform/modules/platform-api-irsa/`. Key permissions:

- `sts:AssumeRole` on `arn:aws:iam::*:role/aegis-project-*` -- allows assuming any project role
- `sts:GetCallerIdentity` -- required for EKS token generation
- `eks:DescribeCluster` -- dynamic endpoint + CA discovery

### Spoke/Project Account (spoke-project-role)

Managed by `terraform/modules/spoke-project-role/`. Creates:

- **IAM Role**: `aegis-project-{projectId}`
- **Trust Policy**: only the hub's platform-api IRSA role can assume it, with ExternalID condition
- **Permissions**: EKS, EC2, IAM, STS, CloudFormation, KMS, CloudWatch Logs

Trust policy template:
```json
{
  "Effect": "Allow",
  "Principal": { "AWS": "<hub-platform-api-role-arn>" },
  "Action": "sts:AssumeRole",
  "Condition": {
    "StringEquals": {
      "sts:ExternalId": "<external-id>"
    }
  }
}
```

### Production Trust Chain

```
Hub Account (platform-api pod via IRSA)
    |
    | sts:AssumeRole + ExternalID
    v
Spoke Account (aegis-project-{id} role)
    |
    | EKS, EC2, IAM, CloudFormation...
    v
Project AWS Resources (EKS cluster, node groups, etc.)
```

---

## Troubleshooting

### "AccessDenied when assuming role"

The trust policy principal doesn't match the hub's platform-api IRSA role ARN. Check:
```bash
aws iam get-role --role-name aegis-project-<id> --query 'Role.AssumeRolePolicyDocument'
```
Ensure the Principal ARN matches the output of:
```bash
terraform -chdir=terraform/ output -raw platform_api_irsa_role_arn
```

### "InvalidExternalId"

The external ID in the project's AWS credentials doesn't match the IAM role trust condition. Verify the external ID set during project creation matches the one in the trust policy.

### "Pulumi uses wrong account"

Check that `AEGIS_DEV_MODE` is explicitly `"false"`, not an empty string. The `IsDevMode()` function checks for `true`, `1`, or `yes` -- anything else (including empty) means production mode.

### "UI requires Role ARN but I'm in dev mode"

The backend `GetPlatformConfig` endpoint may not be returning `devMode: true`. Check:
1. `AEGIS_DEV_MODE` is set in the pod's environment
2. The `/api/v1/platform/config` endpoint is reachable through the Backstage proxy
3. Browser console for network errors fetching the config

### "Config endpoint returns 404"

The `GetPlatformConfig` handler is registered on the HTTP gateway at `/api/v1/platform/config`. Ensure:
1. Platform-api is running the latest version with the handler
2. The Backstage proxy is configured to forward `/api/v1/` paths
