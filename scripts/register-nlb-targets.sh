#!/bin/bash
# register-nlb-targets.sh - Register EKS node instances with the spoke-proxy NLB target group
#
# Usage: ./scripts/register-nlb-targets.sh <cluster-name>
#
# This script is used after Pulumi provisions a cluster to register the cluster's
# nodes with the pre-created NLB target group for stable spoke-proxy connections.

set -euo pipefail

CLUSTER_NAME="${1:-}"
if [[ -z "$CLUSTER_NAME" ]]; then
    echo "Usage: $0 <cluster-name>"
    echo "Example: $0 demo-1-us-east-1-demo-3"
    exit 1
fi

# Get the target group ARN from terraform
cd "$(dirname "$0")/../terraform"

TARGET_GROUP_ARN=$(terraform output -raw spoke_proxy_target_group_arn 2>/dev/null || echo "")
if [[ -z "$TARGET_GROUP_ARN" ]]; then
    echo "Error: Could not get spoke_proxy_target_group_arn from terraform"
    echo "Make sure you've run: terraform apply -target=aws_lb_target_group.spoke_proxy"
    exit 1
fi

echo "Target Group ARN: $TARGET_GROUP_ARN"

# Find all instances for this cluster
echo "Finding instances for cluster: $CLUSTER_NAME"
INSTANCE_IDS=$(aws ec2 describe-instances \
    --filters "Name=tag:eks:cluster-name,Values=$CLUSTER_NAME" "Name=instance-state-name,Values=running" \
    --query 'Reservations[*].Instances[*].InstanceId' \
    --output text | tr '\t' ' ')

if [[ -z "$INSTANCE_IDS" ]]; then
    echo "Error: No running instances found for cluster $CLUSTER_NAME"
    echo "Make sure the cluster has been provisioned and has running nodes."
    exit 1
fi

echo "Found instances: $INSTANCE_IDS"

# Build the targets array for the API call
TARGETS=""
for ID in $INSTANCE_IDS; do
    TARGETS="$TARGETS Id=$ID"
done

# Register targets
echo "Registering targets with NLB..."
aws elbv2 register-targets \
    --target-group-arn "$TARGET_GROUP_ARN" \
    --targets $TARGETS

echo "Successfully registered instances with target group"

# Check health status
echo ""
echo "Checking target health (may take a moment to become healthy)..."
sleep 5
aws elbv2 describe-target-health \
    --target-group-arn "$TARGET_GROUP_ARN" \
    --query 'TargetHealthDescriptions[*].{Instance:Target.Id,Health:TargetHealth.State}' \
    --output table

echo ""
echo "Done! The spoke-proxy should now be accessible via the NLB."
echo ""
echo "Proxy URL: wss://spoke-proxy.$(terraform output -raw spoke_proxy_eip).nip.io:443"
