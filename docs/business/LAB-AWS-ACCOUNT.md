# Lab AWS Account (smoke / demo)

**Last updated:** 2026-07-23

| Field | Value |
|-------|--------|
| **Account ID** | `471147325433` |
| **Region** | `us-east-1` |
| **CLI profile** | `aegis-lab` (also mirrored as `aegis-new` for older scripts) |
| **EKS cluster (lab)** | `aegis-hub-lab` |
| **ECR registry** | `471147325433.dkr.ecr.us-east-1.amazonaws.com` |
| **Terraform state** | **Local** — `terraform/terraform.tfstate` (not the old S3 bucket) |
| **Cloudflare DNS** | Disabled for lab smoke (`enable_cloudflare = false`) |
| **GPU nodes** | Off for smoke (CPU node group only) |

## Supersedes (do not use for new deploys)

| Old | Value |
|-----|--------|
| Account | `195714074609` |
| ECR | `195714074609.dkr.ecr.us-east-1.amazonaws.com` |
| TF state bucket | `aegis-tf-state-195714074609` |
| Profile example | historical `aegis-new` keys for the **old** account |

## Credentials

- Store access keys **only** in `~/.aws/credentials` under `[aegis-lab]`.
- **Never commit** keys or tokens to git.
- If keys were pasted into chat, **rotate them** in IAM after the session.

```bash
export AWS_PROFILE=aegis-lab
aws sts get-caller-identity   # must show Account 471147325433
```

## Deploy / destroy (lab)

```bash
cd ~/code/aegis-platform/terraform
export AWS_PROFILE=aegis-lab
terraform init -reconfigure   # local backend
terraform apply               # or use saved plan
eval "$(terraform output -raw kubectl_config_command)"

cd ..
export AWS_PROFILE=aegis-lab
export AWS_ECR_REGISTRY=471147325433.dkr.ecr.us-east-1.amazonaws.com
# Preferred path: images from GitLab CI (docs/CI-GITLAB-IMAGES.md), then:
export CLOUD_IMAGE_TAG=<CI_COMMIT_SHORT_SHA>
./scripts/hub-eks.sh up
# (never builds images on the laptop by default)
# Rare local build: ./scripts/hub-eks.sh images   # needs Docker
```

**Teardown** (after helm uninstall if apps created LBs):

```bash
# If helm installed:
helm uninstall aegis -n aegis-system 2>/dev/null || true
# Wait for NLBs to delete, then:
cd terraform && AWS_PROFILE=aegis-lab terraform destroy -auto-approve
```

See also `terraform/DESTROY_CHECKLIST.md` (update profile names to `aegis-lab`).

## Makefile defaults

- `AWS_PROFILE` default: `aegis-lab`
- `AWS_ECR_REGISTRY` default: `471147325433.dkr.ecr.us-east-1.amazonaws.com`
- `SKIP_DNS_UPDATE` default: `1` (private lab)
