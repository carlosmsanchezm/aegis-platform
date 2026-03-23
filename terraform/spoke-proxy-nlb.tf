# Spoke Proxy Network Load Balancer Infrastructure
# This creates a public-facing NLB with stable Elastic IPs for VS Code remote connections
#
# IMPORTANT: This uses the DEFAULT VPC where Pulumi provisions EKS clusters.
# The NLB must be in the same VPC as the cluster nodes to route traffic correctly.
#
# Architecture:
#   VS Code Extension (external client)
#         |
#         | WebSocket over TLS (public internet)
#         v
#   NLB (public, internet-facing) ← Elastic IP for stable DNS
#         |
#         | Target: NodePort or IP targets
#         v
#   Spoke-proxy Pod (EKS cluster)
#         |
#         v
#   Workspace Pod
#
# Benefits:
#   - Stable IP address (Elastic IP doesn't change)
#   - Works with nip.io for dynamic DNS (spoke-proxy.<EIP>.nip.io)
#   - Or use Route53 for custom DNS (spoke-proxy.aegist.dev)
#   - TCP passthrough preserves WebSocket/TLS

################################################################################
# Data sources for default VPC (where Pulumi creates clusters)
################################################################################

data "aws_vpc" "default" {
  count   = var.enable_spoke_proxy_nlb ? 1 : 0
  default = true
}

data "aws_subnets" "default_public" {
  count = var.enable_spoke_proxy_nlb ? 1 : 0

  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default[0].id]
  }

  filter {
    name   = "default-for-az"
    values = ["true"]
  }
}

################################################################################
# Local values for spoke-proxy NLB
################################################################################

locals {
  spoke_proxy_name = "aegis-spoke-proxy"
  spoke_proxy_tags = {
    Name      = local.spoke_proxy_name
    Purpose   = "spoke-proxy-nlb"
    ManagedBy = "terraform"
  }
}

################################################################################
# Elastic IP for stable spoke-proxy URL
################################################################################

# Note: EIP is no longer used since NLB spans multiple AZs.
# The NLB DNS name provides a stable endpoint instead.
# Keeping this commented for reference in case single-AZ deployment is needed.
# resource "aws_eip" "spoke_proxy" {
#   count  = var.enable_spoke_proxy_nlb ? 1 : 0
#   domain = "vpc"
#
#   tags = merge(local.spoke_proxy_tags, {
#     Name = "${local.spoke_proxy_name}-eip"
#   })
# }

################################################################################
# Security Group for spoke-proxy NLB targets
################################################################################

resource "aws_security_group" "spoke_proxy_nlb" {
  count       = var.enable_spoke_proxy_nlb ? 1 : 0
  name_prefix = "spoke-proxy-nlb-"
  description = "Security group for spoke-proxy NLB traffic"
  vpc_id      = data.aws_vpc.default[0].id

  # Allow WebSocket traffic from anywhere (VS Code clients)
  ingress {
    from_port   = var.spoke_proxy_port
    to_port     = var.spoke_proxy_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "WebSocket traffic from VS Code clients"
  }

  # Allow health checks from NLB
  ingress {
    from_port   = var.spoke_proxy_port
    to_port     = var.spoke_proxy_port
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.default[0].cidr_block]
    description = "Health checks from NLB"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = merge(local.spoke_proxy_tags, {
    Name = "${local.spoke_proxy_name}-spoke-proxy-nlb-sg"
  })
}

################################################################################
# Network Load Balancer for spoke-proxy
################################################################################

resource "aws_lb" "spoke_proxy" {
  count              = var.enable_spoke_proxy_nlb ? 1 : 0
  name               = "${local.spoke_proxy_name}-spoke-proxy"
  internal           = false # Public-facing for VS Code connections
  load_balancer_type = "network"

  # Use ALL public subnets to span all AZs where EKS nodes may be provisioned.
  # This ensures the NLB can route traffic to nodes in any AZ.
  # Note: We can only assign one EIP, so we use subnets directly (AWS assigns IPs).
  subnets = data.aws_subnets.default_public[0].ids

  enable_cross_zone_load_balancing = true
  enable_deletion_protection       = false # Set to true for production

  tags = merge(local.spoke_proxy_tags, {
    Name = "${local.spoke_proxy_name}-spoke-proxy-nlb"
  })
}

################################################################################
# Target Group for spoke-proxy (NodePort targets)
################################################################################

resource "aws_lb_target_group" "spoke_proxy" {
  count       = var.enable_spoke_proxy_nlb ? 1 : 0
  name        = "${local.spoke_proxy_name}-spoke-proxy"
  port        = var.spoke_proxy_port
  protocol    = "TCP"
  vpc_id      = data.aws_vpc.default[0].id
  target_type = "instance" # Target EKS worker nodes

  health_check {
    enabled             = true
    protocol            = "TCP"
    port                = var.spoke_proxy_port
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
  }

  tags = merge(local.spoke_proxy_tags, {
    Name = "${local.spoke_proxy_name}-spoke-proxy-tg"
  })
}

################################################################################
# NLB Listener for spoke-proxy
################################################################################

resource "aws_lb_listener" "spoke_proxy" {
  count             = var.enable_spoke_proxy_nlb ? 1 : 0
  load_balancer_arn = aws_lb.spoke_proxy[0].arn
  port              = 443 # Standard HTTPS port for easier firewall rules
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.spoke_proxy[0].arn
  }

  tags = merge(local.spoke_proxy_tags, {
    Name = "${local.spoke_proxy_name}-spoke-proxy-listener"
  })
}

################################################################################
# Target Group Registration
# NOTE: When clusters are provisioned by Pulumi, the ASG/nodes are not managed
# by this terraform. Use the AWS CLI to attach targets after cluster creation:
#
#   # Get node instance IDs from the Pulumi-created cluster
#   INSTANCE_IDS=$(aws ec2 describe-instances \
#     --filters "Name=tag:eks:cluster-name,Values=<cluster-name>" \
#     --query 'Reservations[*].Instances[*].InstanceId' --output text)
#
#   # Register instances with target group
#   aws elbv2 register-targets \
#     --target-group-arn $(terraform output -raw spoke_proxy_target_group_arn) \
#     --targets $(echo $INSTANCE_IDS | xargs -n1 printf 'Id=%s ')
#
# Alternatively, configure the Pulumi runner to create an ASG attachment.
################################################################################

################################################################################
# Route53 DNS Record for spoke-proxy (Optional)
# Uncomment if you have a Route53 hosted zone configured
################################################################################

# resource "aws_route53_record" "spoke_proxy" {
#   count   = var.enable_spoke_proxy_nlb ? 1 : 0
#   zone_id = aws_route53_zone.aegist.zone_id
#   name    = "spoke-proxy.aegist.dev"
#   type    = "A"
#   ttl     = 60
#   records = [aws_eip.spoke_proxy[0].public_ip]
# }

# resource "aws_route53_record" "spoke_proxy_wildcard" {
#   count   = var.enable_spoke_proxy_nlb ? 1 : 0
#   zone_id = aws_route53_zone.aegist.zone_id
#   name    = "*.spoke-proxy.aegist.dev"
#   type    = "A"
#   ttl     = 60
#   records = [aws_eip.spoke_proxy[0].public_ip]
# }

################################################################################
# Outputs
################################################################################

output "spoke_proxy_nlb_dns" {
  description = "DNS name of the spoke-proxy NLB"
  value       = try(aws_lb.spoke_proxy[0].dns_name, "")
}

output "spoke_proxy_target_group_arn" {
  description = "ARN of the spoke-proxy target group (for registering instances)"
  value       = try(aws_lb_target_group.spoke_proxy[0].arn, "")
}

# EIP outputs removed - NLB now spans all AZs and uses DNS name instead of EIP
# output "spoke_proxy_eip" {
#   description = "Elastic IP address for spoke-proxy (stable URL)"
#   value       = try(aws_eip.spoke_proxy[0].public_ip, "")
# }

# output "spoke_proxy_url_nip_io" {
#   description = "Spoke-proxy URL using nip.io (no DNS required)"
#   value       = try("wss://spoke-proxy.${aws_eip.spoke_proxy[0].public_ip}.nip.io:443", "")
# }

output "spoke_proxy_url" {
  description = "Spoke-proxy URL using NLB DNS name"
  value       = try("wss://${aws_lb.spoke_proxy[0].dns_name}:443", "")
}

# output "spoke_proxy_url_dns" {
#   description = "Spoke-proxy URL using Route53 DNS"
#   value       = try("wss://${aws_route53_record.spoke_proxy[0].fqdn}:443", "")
# }

output "spoke_proxy_helm_values" {
  description = "Helm values to use for aegis-spoke deployment"
  value = try(<<-EOF
    # Add to your aegis-spoke helm values:
    proxy:
      service:
        type: NodePort
        nodePort: ${var.spoke_proxy_port}
      ingress:
        enabled: false  # NLB handles external access
        hostname: "${aws_lb.spoke_proxy[0].dns_name}"

    k8sAgent:
      env:
        AEGIS_PROXY_INGRESS_HOST: "${aws_lb.spoke_proxy[0].dns_name}:443"
  EOF
  , "")
}
