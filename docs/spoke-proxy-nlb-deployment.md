# Spoke-Proxy NLB Deployment Guide

This guide explains how to deploy the spoke-proxy with a stable Network Load Balancer (NLB) for VS Code remote connections.

## Quick Start for Demo

```bash
# 1. Deploy NLB infrastructure (creates EIP + NLB)
cd terraform && terraform apply -target=aws_eip.spoke_proxy -target=aws_lb.spoke_proxy ...

# 2. Get the stable proxy URL
export SPOKE_PROXY_HOST="spoke-proxy.$(terraform output -raw spoke_proxy_eip).nip.io:443"

# 3. Set in platform-api deployment (tells Pulumi to use this URL)
kubectl set env deployment/platform-api -n aegis-system AEGIS_SPOKE_PROXY_HOST=$SPOKE_PROXY_HOST

# 4. Provision cluster via UI - it will use the stable URL

# 5. After cluster is provisioned, register nodes with NLB target group
./scripts/register-nlb-targets.sh <cluster-name>
```

## Architecture

```
VS Code Extension (external client)
       │
       │ WebSocket over TLS (port 443)
       ▼
┌──────────────────────────┐
│  AWS Network Load Balancer│ ◄── Elastic IP (stable)
│  (public, internet-facing)│     e.g., 52.1.2.3
└──────────────────────────┘
       │
       │ TCP passthrough to NodePort 31484
       ▼
┌──────────────────────────┐
│  EKS Worker Nodes        │
│  ┌────────────────────┐  │
│  │  Spoke-proxy Pod   │  │
│  │  (NodePort 31484)  │  │
│  └────────────────────┘  │
└──────────────────────────┘
       │
       │ Internal routing
       ▼
┌──────────────────────────┐
│  Workspace Pod           │
└──────────────────────────┘
```

## Benefits

- **Stable URL**: Elastic IP doesn't change when pods restart
- **No DNS required**: Use nip.io (e.g., `spoke-proxy.52.1.2.3.nip.io`)
- **TCP passthrough**: NLB preserves WebSocket/TLS connections
- **Auto-scaling**: ASG attachment automatically registers new nodes

## Prerequisites

1. EKS cluster deployed via terraform
2. AWS CLI configured with appropriate credentials
3. Helm 3.x installed
4. kubectl configured to access the EKS cluster

## Deployment Steps

### Step 1: Deploy the NLB Infrastructure

```bash
cd terraform

# Initialize terraform (if not already done)
terraform init

# Preview the changes
terraform plan -target=aws_eip.spoke_proxy -target=aws_lb.spoke_proxy

# Apply the spoke-proxy NLB resources
terraform apply -target=aws_eip.spoke_proxy \
                -target=aws_security_group.spoke_proxy_nlb \
                -target=aws_lb.spoke_proxy \
                -target=aws_lb_target_group.spoke_proxy \
                -target=aws_lb_listener.spoke_proxy \
                -target=aws_autoscaling_attachment.spoke_proxy \
                -target=aws_route53_record.spoke_proxy \
                -target=aws_route53_record.spoke_proxy_wildcard
```

### Step 2: Get the Elastic IP

```bash
# Get the Elastic IP for the spoke-proxy
export SPOKE_PROXY_EIP=$(terraform output -raw spoke_proxy_eip)
echo "Spoke-proxy Elastic IP: $SPOKE_PROXY_EIP"

# The nip.io URL will be:
echo "Spoke-proxy URL: wss://spoke-proxy.$SPOKE_PROXY_EIP.nip.io:443"
```

### Step 3: Generate TLS Certificate

The spoke-proxy needs a TLS certificate with SANs matching the nip.io domain:

```bash
# Generate a self-signed certificate (for testing)
openssl req -x509 -newkey rsa:2048 \
  -keyout /tmp/spoke-proxy.key \
  -out /tmp/spoke-proxy.crt \
  -days 365 -nodes \
  -subj "/CN=spoke-proxy.aegis.local" \
  -addext "subjectAltName=DNS:spoke-proxy.aegis.local,DNS:*.nip.io,DNS:*.*.nip.io,DNS:*.*.*.nip.io,DNS:*.*.*.*.nip.io,DNS:*.*.*.*.*.nip.io,DNS:spoke-proxy.$SPOKE_PROXY_EIP.nip.io,DNS:localhost,IP:$SPOKE_PROXY_EIP,IP:127.0.0.1"

# View the certificate SANs
openssl x509 -in /tmp/spoke-proxy.crt -text -noout | grep -A1 "Subject Alternative Name"
```

### Step 4: Deploy aegis-spoke with NLB Configuration

```bash
# Get kubeconfig for the EKS cluster
aws eks update-kubeconfig --name aegis-spoke-prod --region us-east-1

# Create namespace
kubectl create namespace aegis-system --dry-run=client -o yaml | kubectl apply -f -

# Create TLS secret
kubectl create secret tls spoke-proxy-tls \
  --cert=/tmp/spoke-proxy.crt \
  --key=/tmp/spoke-proxy.key \
  -n aegis-system \
  --dry-run=client -o yaml | kubectl apply -f -

# Deploy aegis-spoke
helm upgrade --install aegis-spoke ./charts/aegis-spoke \
  -n aegis-system \
  -f charts/aegis-spoke/values.yaml \
  -f charts/aegis-spoke/values-nlb.yaml \
  --set k8sAgent.env.AEGIS_CLUSTER_ID=demo-cluster-1 \
  --set k8sAgent.env.AEGIS_REGION=us-east-1 \
  --set k8sAgent.env.AEGIS_PROVIDER=aws \
  --set k8sAgent.env.AEGIS_PROXY_INGRESS_HOST="spoke-proxy.$SPOKE_PROXY_EIP.nip.io:443" \
  --set proxy.tls.cert="$(cat /tmp/spoke-proxy.crt)" \
  --set proxy.tls.key="$(cat /tmp/spoke-proxy.key)"
```

### Step 5: Verify Deployment

```bash
# Check pods are running
kubectl get pods -n aegis-system

# Check the NodePort service
kubectl get svc -n aegis-system | grep proxy

# Test connectivity (from outside the cluster)
curl -k https://spoke-proxy.$SPOKE_PROXY_EIP.nip.io:443/health
```

### Step 6: Configure VS Code Extension

The VS Code extension needs the spoke-proxy certificate in its trust bundle:

```bash
# Copy the spoke-proxy certificate to your trust bundle
cat /tmp/spoke-proxy.crt >> ~/aegis-local-trust.pem

# Launch VS Code with the trust bundle
NODE_EXTRA_CA_CERTS=~/aegis-local-trust.pem code
```

## Using with cert-manager (Production)

For production deployments, use cert-manager to automatically issue certificates:

```yaml
# values-nlb-certmanager.yaml
proxy:
  tls:
    certManager:
      enabled: true
      issuerRef:
        name: letsencrypt-prod  # Or your internal CA
        kind: ClusterIssuer
        group: cert-manager.io
      dnsNames:
        - "spoke-proxy.${SPOKE_PROXY_EIP}.nip.io"
```

## Troubleshooting

### NLB Target Health Check Failing

```bash
# Check if NodePort is accessible on workers
kubectl get nodes -o wide
ssh ec2-user@<node-ip> "curl -k https://localhost:31484/health"

# Check security group allows traffic
aws ec2 describe-security-groups --group-ids <node-sg-id>
```

### Certificate SAN Mismatch

```bash
# Verify certificate SANs include the nip.io domain
openssl s_client -connect spoke-proxy.$SPOKE_PROXY_EIP.nip.io:443 -servername spoke-proxy.$SPOKE_PROXY_EIP.nip.io < /dev/null 2>/dev/null | openssl x509 -text -noout | grep -A1 "Subject Alternative Name"
```

### Proxy URL Not Registered

Check the k8s-agent logs to verify it's reporting the correct proxy URL:

```bash
kubectl logs -n aegis-system -l app.kubernetes.io/component=k8s-agent | grep proxy
```

## Cost Considerations

| Resource | Cost (us-east-1) |
|----------|------------------|
| Elastic IP (in use) | Free |
| NLB (per hour) | ~$0.0225/hr |
| NLB (per LCU) | ~$0.006/LCU |

Estimated monthly cost: ~$16-20/month for low traffic.

## Clean Up

```bash
# Remove the NLB infrastructure
cd terraform
terraform destroy -target=aws_autoscaling_attachment.spoke_proxy \
                  -target=aws_lb_listener.spoke_proxy \
                  -target=aws_lb_target_group.spoke_proxy \
                  -target=aws_lb.spoke_proxy \
                  -target=aws_security_group.spoke_proxy_nlb \
                  -target=aws_route53_record.spoke_proxy_wildcard \
                  -target=aws_route53_record.spoke_proxy \
                  -target=aws_eip.spoke_proxy
```
