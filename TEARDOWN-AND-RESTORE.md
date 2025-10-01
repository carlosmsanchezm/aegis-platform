# Aegis Test Environment - Teardown & Restore Guide

## Current State (Before Teardown)

**Date:** October 1, 2025

**AWS Resources Running:**
- **EKS Cluster:** `aegis-spoke-prod` (us-east-1)
- **Worker Nodes:** 1x t3.medium (cpu-workers)
- **RDS Database:** `aegis-spoke-prod-db` (db.t3.micro, PostgreSQL)
- **Network Load Balancer:** `a9604fc12441d48e6bedf773e3f63996-98644661.us-east-1.elb.amazonaws.com`
- **VPC/Subnets:** Created by Terraform

**Estimated Monthly Cost:** ~$166-200/month (if left running)

**RDS Snapshot:** `aegis-spoke-prod-db-final-20251001` (created before teardown)

---

## What's Been Tested & Working

### ✅ Successfully Validated
1. **Event-driven architecture** - platform-api → k8s-agent → workload creation
2. **Cloud proxy** - aegis-auth-proxy exposed via NLB, JWT auth working
3. **SSH connections** - aegis-connect → cloud proxy → workspace pods
4. **VS Code integration** - SSH key auth, Remote-SSH extension
5. **Database migrations** - PostgreSQL schema with workloads, connection_sessions tables
6. **GPU workload creation** - k8s-agent correctly handles gpu_count=0

### 📝 Documentation Created
- `/Users/carlossanchez/code/aegis/VSCODE-SSH-SETUP.md`
- `/Users/carlossanchez/code/aegis/CLOUD-PROXY-SETUP.md`
- `/Users/carlossanchez/code/aegis/MVP-UX-ANALYSIS.md`
- `/Users/carlossanchez/code/aegis/VSCODE-EXTENSION-DESIGN-REVIEW.md`
- `/Users/carlossanchez/code/aegis/BACKSTAGE-VSCODE-INTEGRATION.md`

### 🚀 Ready for Next Phase
- VS Code extension implementation (RemoteAuthorityResolver approach)
- Backstage UI integration
- Production-ready authentication

---

## Teardown Procedure

### Option A: Quick Teardown (Recommended for Testing)

Just destroy the expensive resources, keep Terraform state:

```bash
# From aegis/terraform directory
cd /Users/carlossanchez/code/aegis/terraform

# Destroy everything
terraform destroy --auto-approve

# This will remove:
# - EKS cluster (including worker nodes)
# - RDS database (snapshot preserved)
# - Network Load Balancer
# - VPC, subnets, security groups
# - IAM roles/policies
```

**Cost after teardown:** $0.05-0.10/day for snapshots only (~$1.50-3/month)

### Option B: Selective Teardown (Keep Some Resources)

If you want to keep certain resources (e.g., VPC for faster restoration):

```bash
# Target specific resources for destruction
terraform destroy -target=module.eks --auto-approve
terraform destroy -target=aws_db_instance.aegis_spoke_db --auto-approve
terraform destroy -target=kubernetes_service.proxy --auto-approve
```

**Not recommended** - complexity isn't worth the small savings.

---

## Important: What Gets Preserved

### ✅ Automatically Preserved (No Action Needed)
1. **RDS Snapshots** - Stored in S3, available for restore
   - `aegis-spoke-prod-db-final-20251001` (manual)
   - Automatic daily snapshots (if enabled)
2. **Docker Images** - On Docker Hub
   - `carlosmsanchez/aegis-k8s-agent:latest`
   - `carlosmsanchez/aegis-platform-api:latest`
   - `carlosmsanchez/aegis-auth-proxy:latest`
   - `carlosmsanchez/aegis-workspace-vscode:latest`
3. **Terraform State** - Local file at `terraform/terraform.tfstate`
4. **Git Repositories** - All code in local git repos

### ⚠️ Will Be Lost (Need Manual Backup)
1. **Kubernetes ConfigMaps/Secrets** - Export before destroying
2. **Running workspace pods** - Any active workspaces will be terminated
3. **Logs** - CloudWatch logs will be deleted (unless retention set)

---

## Pre-Teardown: Export Important Configs

### 1. Export Kubernetes Resources

```bash
# Export all resources from aegis-services namespace
kubectl get all,cm,secrets -n aegis-services -o yaml > /tmp/aegis-services-backup.yaml

# Export platform-api migrations ConfigMap (important!)
kubectl get configmap platform-api-migrations -n aegis-services -o yaml > /tmp/platform-api-migrations.yaml

# Export any secrets (encrypted)
kubectl get secrets -n aegis-services -o yaml > /tmp/aegis-secrets.yaml

# Save to repo
cp /tmp/aegis-services-backup.yaml /Users/carlossanchez/code/aegis/backups/
cp /tmp/platform-api-migrations.yaml /Users/carlossanchez/code/aegis/charts/aegis-services/files/platform-api/migrations-backup.yaml
```

### 2. Export Database Schema (Already in Git)

Your migrations are already in:
```
charts/aegis-services/files/platform-api/migrations/
  000001_init.up.sql
  0002_clusters.up.sql
```

### 3. Export Helm Values (Already in Git)

```
charts/aegis-services/values-test-cloud.yaml
charts/aegis-spoke/values-test-cloud.yaml
```

### 4. Copy Terraform State to Safe Location

```bash
cp /Users/carlossanchez/code/aegis/terraform/terraform.tfstate \
   /Users/carlossanchez/code/aegis/backups/terraform.tfstate.$(date +%Y%m%d)
```

---

## Restoration Procedure (When Ready to Test Again)

### Step 1: Restore AWS Infrastructure (~10-15 minutes)

```bash
cd /Users/carlossanchez/code/aegis/terraform

# Re-apply Terraform (will recreate everything)
terraform apply --auto-approve

# Terraform will:
# - Create new EKS cluster
# - Create new RDS instance FROM SNAPSHOT (faster than fresh DB)
# - Create new NLB
# - Recreate VPC, subnets, security groups
```

**Important:** Terraform will automatically restore RDS from the latest snapshot if configured:

```hcl
# terraform/rds.tf (check this is set)
resource "aws_db_instance" "aegis_spoke_db" {
  # ...
  snapshot_identifier = "aegis-spoke-prod-db-final-20251001"  # uses this if set
}
```

### Step 2: Configure kubectl (~2 minutes)

```bash
# Update kubeconfig to point to new cluster
aws eks update-kubeconfig --region us-east-1 --name aegis-spoke-prod --profile myclaude

# Verify access
kubectl get nodes
```

### Step 3: Deploy Platform Services (~5 minutes)

```bash
# Deploy platform-api + k8s-agent + proxy
cd /Users/carlossanchez/code/aegis
helm upgrade --install aegis-services ./charts/aegis-services \
  -n aegis-services --create-namespace \
  -f charts/aegis-services/values-test-cloud.yaml

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=platform-api -n aegis-services --timeout=300s
```

### Step 4: Verify Database Migration (~1 minute)

```bash
# Port-forward to platform-api
kubectl port-forward -n aegis-services svc/aegis-services-aegis-services-platform-api 8081:8081 &

# Check database has correct schema
grpcurl -plaintext localhost:8081 list

# Should see:
# aegis.v1.AegisPlatform
```

### Step 5: Test End-to-End (~5 minutes)

```bash
# Create test workspace
grpcurl -plaintext -H "x-aegis-user: testuser@test.com" \
  -d '{"project_id":"p-demo","workspace":{"flavor":"nano","image":"ubuntu:22.04","interactive":true}}' \
  localhost:8081 aegis.v1.AegisPlatform/CreateWorkload

# Get new proxy URL (NLB DNS will be different)
NEW_NLB=$(kubectl get svc -n aegis-services aegis-services-aegis-services-proxy -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
echo "New NLB: $NEW_NLB"

# Create connection session
grpcurl -plaintext -H "x-aegis-user: testuser@test.com" \
  -d '{"workload_id":"<WORKSPACE_ID>","client":"vscode"}' \
  localhost:8081 aegis.v1.AegisPlatform/CreateConnectionSession

# Update SSH config with new NLB URL and test
aegis-refresh-ssh
ssh aegis-w-<WORKSPACE_ID> 'echo "Connection works!"'
```

---

## Cost Optimization Tips

### During Development (Not Actively Testing)
**Tear down AWS resources** - Restore when needed (this guide)
- **Savings:** ~$160/month → ~$2/month (snapshots only)

### During Active Testing
**Use spot instances for worker nodes:**
```hcl
# terraform/eks.tf
resource "aws_eks_node_group" "cpu_workers" {
  # ...
  capacity_type = "SPOT"  # 60-70% cheaper than on-demand
}
```
- **Savings:** ~$30/month → ~$10/month for t3.medium nodes

### Production (Future)
**Use autoscaling and scheduled shutdown:**
- Scale to zero workers during off-hours
- Use Karpenter for aggressive scale-down
- Schedule RDS to stop overnight (for dev/staging only)

---

## Quick Reference Commands

### Teardown
```bash
cd /Users/carlossanchez/code/aegis/terraform
terraform destroy --auto-approve
```

### Restore
```bash
cd /Users/carlossanchez/code/aegis/terraform
terraform apply --auto-approve
aws eks update-kubeconfig --region us-east-1 --name aegis-spoke-prod --profile myclaude
helm upgrade --install aegis-services ./charts/aegis-services -n aegis-services --create-namespace -f charts/aegis-services/values-test-cloud.yaml
```

### Check Costs
```bash
# Get AWS cost estimate for last 30 days
aws ce get-cost-and-usage --profile myclaude \
  --time-period Start=2025-09-01,End=2025-10-01 \
  --granularity MONTHLY \
  --metrics "UnblendedCost" \
  --query 'ResultsByTime[*].[TimePeriod.Start,Total.UnblendedCost.Amount]' \
  --output table
```

---

## Checklist Before Teardown

- [ ] **RDS snapshot created** (manual backup)
- [ ] **Terraform state backed up** to `backups/` directory
- [ ] **Kubernetes configs exported** (ConfigMaps, Secrets)
- [ ] **All code committed to git** (no uncommitted changes)
- [ ] **Docker images pushed** to Docker Hub
- [ ] **Documentation complete** (all .md files saved)
- [ ] **No active user workspaces** (no one is actively using the cluster)

---

## Estimated Restoration Time

| Step | Time | Notes |
|------|------|-------|
| Terraform apply | 10-15 min | EKS cluster creation is slowest |
| kubectl config | 2 min | Update kubeconfig |
| Helm install | 5 min | Platform services deployment |
| Database verify | 1 min | Check schema, run migrations |
| End-to-end test | 5 min | Create workspace, test SSH |
| **Total** | **~25 minutes** | From `terraform apply` to working system |

---

## Troubleshooting Common Restoration Issues

### Issue: RDS won't restore from snapshot
**Cause:** Snapshot ID not set in Terraform or snapshot deleted
**Fix:**
```bash
# List available snapshots
aws rds describe-db-snapshots --profile myclaude --region us-east-1 \
  --query 'DBSnapshots[?starts_with(DBSnapshotIdentifier,`aegis-spoke-prod-db`)].[DBSnapshotIdentifier,Status]' \
  --output table

# Update terraform/rds.tf with correct snapshot ID
```

### Issue: NLB DNS changed, SSH connections fail
**Cause:** New NLB gets new DNS name after recreation
**Fix:**
```bash
# Get new NLB DNS
kubectl get svc -n aegis-services aegis-services-aegis-services-proxy -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'

# Update AEGIS_PROXY_BASE_URL in platform-api deployment
kubectl set env deployment/platform-api -n aegis-services \
  AEGIS_PROXY_BASE_URL="https://<NEW_NLB_DNS>:8080"
```

### Issue: Helm install fails (CRD already exists)
**Cause:** Old CRDs from previous install not cleaned up
**Fix:**
```bash
# Delete old CRDs
kubectl delete crd aegisworkloads.aegis.yourorg.dev

# Retry Helm install
helm upgrade --install aegis-services ./charts/aegis-services ...
```

### Issue: Platform-API can't connect to RDS
**Cause:** Security group rules not applied yet or RDS not ready
**Fix:**
```bash
# Check RDS status
aws rds describe-db-instances --profile myclaude --region us-east-1 \
  --db-instance-identifier aegis-spoke-prod-db \
  --query 'DBInstances[0].[DBInstanceStatus,Endpoint.Address]'

# Wait for "available" status, then restart platform-api pod
kubectl rollout restart deployment/platform-api -n aegis-services
```

---

## What You Can Do During Teardown Period (No AWS Costs)

1. **Implement VS Code extension** - All local development (Phase 0-1)
2. **Build Backstage UI** - Run locally with `yarn dev`
3. **Improve k8s-agent** - Test locally with kind/minikube
4. **Write more tests** - Unit tests, integration tests
5. **Refine documentation** - Update guides based on learnings

**You can test locally and restore AWS when you need to test cloud integration.**

---

## Final Decision Matrix

| Scenario | Recommendation | Cost Impact |
|----------|---------------|-------------|
| **Not actively testing for >1 week** | ✅ Tear down | Save ~$160/month |
| **Testing in 1-3 days** | Maybe keep running | Costs ~$5-15 for a few days |
| **Need to demo immediately** | Keep running | Accept the cost |
| **Working on local dev (extension/backstage)** | ✅ Tear down | No cloud needed |
| **Team is actively using** | Keep running | Worth the cost |

**Your situation:** Working on VS Code extension design (local dev) → **Tear down AWS resources**

---

## Post-Teardown Cost: ~$2-3/month

**What you'll pay for:**
- RDS snapshots in S3: ~$0.095/GB/month (assuming ~20GB = ~$2/month)
- Terraform state (negligible, stored locally)

**What you won't pay for:**
- ❌ EC2 instances ($0)
- ❌ EKS control plane ($0)
- ❌ Load balancers ($0)
- ❌ RDS database ($0)
- ❌ Data transfer ($0)

---

## Ready to Proceed?

Run this command when you're ready to tear down:

```bash
cd /Users/carlossanchez/code/aegis/terraform
terraform destroy --auto-approve
```

**Duration:** ~10 minutes to destroy everything

**You can restore in ~25 minutes when needed using this guide.**
