# Aegis Quick Reference

## 🚀 Cloud Deployment (From Scratch)

```bash
# 1. Deploy infrastructure
cd terraform/
terraform init && terraform apply

# 2. Generate Helm values
./generate-helm-values.sh

# 3. Update domains (replace 'yourdomain.com')
DOMAIN="your-actual-domain.com"
sed -i '' "s/yourdomain.com/${DOMAIN}/g" ../charts/aegis-services/values-cloud-generated.yaml
sed -i '' "s/yourdomain.com/${DOMAIN}/g" ../charts/aegis-spoke/values-cloud-generated.yaml

# 4. Configure kubectl
aws eks update-kubeconfig --region $(terraform output -raw aws_region) --name $(terraform output -raw cluster_name)

# 5. Create secrets
kubectl create secret generic aegis-platform-secrets \
  --from-literal=db-password="$(terraform output -raw db_password_secret_value)" \
  --from-literal=proxy-jwt-secret="$(terraform output -raw jwt_secret_value)" \
  --namespace aegis-system --create-namespace

# 6. Deploy
cd ../charts/
helm upgrade --install aegis-services ./aegis-services \
  -f ./aegis-services/values/common.yaml \
  -f ./aegis-services/values/cloud.yaml \
  -f ./aegis-services/values-cloud-generated.yaml \
  --namespace aegis-system --create-namespace

helm upgrade --install aegis-spoke ./aegis-spoke \
  -f ./aegis-spoke/values-cloud.yaml \
  -f ./aegis-spoke/values-cloud-generated.yaml \
  --namespace aegis-system
```

## 🏠 Local Development

### Start Local Cluster
```bash
# Port-forward services
PF_PLATFORM_HTTP_PORT=10080 PF_PLATFORM_GRPC_PORT=10081 \\
kubectl port-forward -n default svc/aegis-services-aegis-services-platform-api $PF_PLATFORM_HTTP_PORT:8080 $PF_PLATFORM_GRPC_PORT:8081 &
PF_PROXY_HTTP_PORT=10085 \\
kubectl port-forward -n default svc/aegis-services-aegis-services-proxy $PF_PROXY_HTTP_PORT:8085 &

> Adjust `PF_PLATFORM_HTTP_PORT`, `PF_PLATFORM_GRPC_PORT`, or `PF_PROXY_HTTP_PORT` if the defaults clash with other services.

# Start Backstage
cd aegis-platform/
yarn start  # Runs on http://localhost:7008
```

### VS Code Extension (Local)
```bash
# Launch VS Code with proposed APIs
/Applications/Visual\ Studio\ Code.app/Contents/Resources/app/bin/code \
  --enable-proposed-api aegis.aegis-remote

# Then in VS Code:
# 1. Cmd+Shift+P → "Aegis: Sign In"
#    - Username: dev-user@example.com
#    - Token: supersecret
# 2. Go to Backstage and click "Open in VS Code"
```

### Deploy Local Helm Charts
```bash
helm upgrade --install aegis-services ./charts/aegis-services \
  -f ./charts/aegis-services/values/common.yaml \
  -f ./charts/aegis-services/values/local.yaml \
  --namespace default

helm upgrade --install aegis-spoke ./charts/aegis-spoke \
  -f ./charts/aegis-spoke/values-local.yaml \
  --namespace default
```

## 📊 Terraform Outputs

### View All Outputs
```bash
cd terraform/
terraform output
```

### Specific Outputs
```bash
# Cluster info
terraform output cluster_name
terraform output cluster_endpoint
terraform output kubectl_config_command

# Database
terraform output rds_endpoint
terraform output rds_database_name

# ECR
terraform output ecr_registry_url

# Helm values (YAML)
terraform output helm_values_aegis_services
terraform output helm_values_aegis_spoke

# Secrets (sensitive)
terraform output -raw db_password_secret_value
terraform output -raw jwt_secret_value

# K8s secret commands
terraform output -raw k8s_secret_commands
```

## 🐛 Debugging

### Check Services
```bash
# Pods
kubectl get pods -n aegis-system

# Services
kubectl get svc -n aegis-system

# Ingress
kubectl get ingress -n aegis-system

# Secrets
kubectl get secrets -n aegis-system
```

### View Logs
```bash
# Platform API
kubectl logs -n aegis-system -l app.kubernetes.io/component=platform-api -f

# Proxy
kubectl logs -n aegis-system -l app.kubernetes.io/component=proxy -f

# k8s-Agent
kubectl logs -n aegis-system -l app.kubernetes.io/component=k8s-agent -f
```

### Test Endpoints
```bash
# Platform API health
curl http://localhost:8080/healthz  # local
curl https://platform-api.yourdomain.com/healthz  # cloud

# List workloads (requires grpcurl)
grpcurl -plaintext localhost:8081 aegis.v1.AegisPlatform/ListWorkloads  # local
grpcurl platform-api-grpc.yourdomain.com:443 aegis.v1.AegisPlatform/ListWorkloads  # cloud

# Proxy
curl -v http://localhost:10085  # local
curl -v https://proxy.yourdomain.com/proxy  # cloud
```

### VS Code Extension Debug
```bash
# View extension logs
# In VS Code: Cmd+Shift+P → "Aegis: Show Logs"

# Or check log file directly
tail -f ~/Library/Application\ Support/Code/logs/*/window1/exthost/output_logging_*/2-Aegis\ Remote.log
```

## 🔄 Update Workflow

### Update Infrastructure
```bash
cd terraform/
terraform apply
./generate-helm-values.sh
```

### Update Helm Deployment
```bash
cd charts/
helm upgrade aegis-services ./aegis-services \
  -f ./aegis-services/values/common.yaml \
  -f ./aegis-services/values/cloud.yaml \
  -f ./aegis-services/values-cloud-generated.yaml \
  --namespace aegis-system
```

### Update VS Code Extension
```bash
cd /path/to/aegis-vscode-remote/extension
npm run build
npx @vscode/vsce package --out aegis-remote.vsix
/Applications/Visual\ Studio\ Code.app/Contents/Resources/app/bin/code \
  --install-extension aegis-remote.vsix --force
```

## 📁 Important Files

### Local Configuration
- `charts/aegis-services/values/common.yaml` - Hub base defaults (shared)
- `charts/aegis-services/values/local.yaml` - Hub overrides (local)
- `charts/aegis-spoke/values-local.yaml` - Spoke config (local)

### Cloud Configuration
- `charts/aegis-services/values/common.yaml` - Hub base defaults (shared)
- `charts/aegis-services/values/cloud.yaml` - Hub overlay (cloud)
- `charts/aegis-services/values-cloud-generated.yaml` - Hub config (from Terraform)
- `charts/aegis-spoke/values-cloud.yaml` - Spoke config (static)
- `charts/aegis-spoke/values-cloud-generated.yaml` - Spoke config (from Terraform)

### Terraform
- `terraform/outputs.tf` - Terraform outputs
- `terraform/generate-helm-values.sh` - Auto-generate Helm values
- `terraform/terraform.tfvars` - Your Terraform variables

### VS Code Extension
- `/path/to/aegis-vscode-remote/extension/README.md` - Extension docs
- `~/Library/Application Support/Code/User/settings.json` - VS Code settings

## 🔑 Important Values

### Local
- **Platform API HTTP**: http://localhost:8080
- **Platform API gRPC**: localhost:8081
- **Proxy**: http://localhost:10085
- **Backstage**: http://localhost:7008
- **Auth**: dev-user@example.com / supersecret
- **JWT Secret**: a-very-secret-key-for-local-dev-must-be-32-chars

### Cloud (Examples)
- **Platform API HTTP**: https://platform-api.yourdomain.com
- **Platform API gRPC**: platform-api-grpc.yourdomain.com:443
- **Proxy**: https://proxy.yourdomain.com
- **DB**: From Terraform output `rds_endpoint`
- **JWT Secret**: From Terraform output `jwt_secret_value`

## 🆘 Common Issues

### "Extension commands not found"
```bash
# VS Code must be launched with proposed APIs
/Applications/Visual\ Studio\ Code.app/Contents/Resources/app/bin/code \
  --enable-proposed-api aegis.aegis-remote
```

### "Can't connect to platform-api"
```bash
# Check port-forward is running (local)
kubectl port-forward -n default svc/aegis-services-aegis-services-platform-api ${PF_PLATFORM_HTTP_PORT:-10080}:8080 ${PF_PLATFORM_GRPC_PORT:-10081}:8081

# Check ingress (cloud)
kubectl get ingress -n aegis-system
kubectl describe ingress aegis-services-platform-api-ingress -n aegis-system
```

### "Database connection error"
```bash
# Check RDS endpoint
terraform output rds_endpoint

# Check secrets
kubectl get secret aegis-platform-secrets -n aegis-system -o yaml

# Check platform-api logs
kubectl logs -n aegis-system -l app.kubernetes.io/component=platform-api
```

### "Proxy authentication failed"
```bash
# Verify JWT secrets match
kubectl get secret aegis-platform-secrets -n aegis-system -o jsonpath='{.data.proxy-jwt-secret}' | base64 -d
terraform output -raw jwt_secret_value

# They should be identical
```
