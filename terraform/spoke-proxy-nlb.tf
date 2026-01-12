# Spoke Proxy Network Load Balancer Infrastructure
# This creates a public-facing NLB with stable Elastic IPs for VS Code remote connections
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
# Variables for spoke-proxy NLB
################################################################################

variable "enable_spoke_proxy_nlb" {
  description = "Enable the spoke-proxy NLB for VS Code remote connections"
  type        = bool
  default     = true
}

variable "spoke_proxy_port" {
  description = "Port for spoke-proxy service (NodePort in EKS)"
  type        = number
  default     = 31484
}

################################################################################
# Elastic IP for stable spoke-proxy URL
################################################################################

resource "aws_eip" "spoke_proxy" {
  count  = var.enable_spoke_proxy_nlb ? 1 : 0
  domain = "vpc"

  tags = merge(local.common_tags, {
    Name    = "${local.cluster_name}-spoke-proxy-eip"
    Purpose = "spoke-proxy-stable-ip"
  })
}

################################################################################
# Security Group for spoke-proxy NLB targets
################################################################################

resource "aws_security_group" "spoke_proxy_nlb" {
  count       = var.enable_spoke_proxy_nlb ? 1 : 0
  name_prefix = "${local.cluster_name}-spoke-proxy-nlb-"
  description = "Security group for spoke-proxy NLB traffic"
  vpc_id      = module.vpc.vpc_id

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
    cidr_blocks = [var.vpc_cidr]
    description = "Health checks from NLB"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-spoke-proxy-nlb-sg"
  })
}

################################################################################
# Network Load Balancer for spoke-proxy
################################################################################

resource "aws_lb" "spoke_proxy" {
  count              = var.enable_spoke_proxy_nlb ? 1 : 0
  name               = "${local.cluster_name}-spoke-proxy"
  internal           = false # Public-facing for VS Code connections
  load_balancer_type = "network"

  # Use public subnets with Elastic IP
  dynamic "subnet_mapping" {
    for_each = [module.vpc.public_subnets[0]] # Single AZ for cost optimization
    content {
      subnet_id     = subnet_mapping.value
      allocation_id = aws_eip.spoke_proxy[0].id
    }
  }

  enable_cross_zone_load_balancing = true
  enable_deletion_protection       = false # Set to true for production

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-spoke-proxy-nlb"
  })
}

################################################################################
# Target Group for spoke-proxy (NodePort targets)
################################################################################

resource "aws_lb_target_group" "spoke_proxy" {
  count       = var.enable_spoke_proxy_nlb ? 1 : 0
  name        = "${local.cluster_name}-spoke-proxy"
  port        = var.spoke_proxy_port
  protocol    = "TCP"
  vpc_id      = module.vpc.vpc_id
  target_type = "instance" # Target EKS worker nodes

  health_check {
    enabled             = true
    protocol            = "TCP"
    port                = var.spoke_proxy_port
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
  }

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-spoke-proxy-tg"
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

  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-spoke-proxy-listener"
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
# Route53 DNS Record for spoke-proxy
################################################################################

resource "aws_route53_record" "spoke_proxy" {
  count   = var.enable_spoke_proxy_nlb ? 1 : 0
  zone_id = aws_route53_zone.aegist.zone_id
  name    = "spoke-proxy.aegist.dev"
  type    = "A"
  ttl     = 60

  records = [aws_eip.spoke_proxy[0].public_ip]
}

# Wildcard record for cluster-specific subdomains
# e.g., spoke-proxy-demo-1.aegist.dev
resource "aws_route53_record" "spoke_proxy_wildcard" {
  count   = var.enable_spoke_proxy_nlb ? 1 : 0
  zone_id = aws_route53_zone.aegist.zone_id
  name    = "*.spoke-proxy.aegist.dev"
  type    = "A"
  ttl     = 60

  records = [aws_eip.spoke_proxy[0].public_ip]
}

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

output "spoke_proxy_eip" {
  description = "Elastic IP address for spoke-proxy (stable URL)"
  value       = try(aws_eip.spoke_proxy[0].public_ip, "")
}

output "spoke_proxy_url_nip_io" {
  description = "Spoke-proxy URL using nip.io (no DNS required)"
  value       = try("wss://spoke-proxy.${aws_eip.spoke_proxy[0].public_ip}.nip.io:443", "")
}

output "spoke_proxy_url_dns" {
  description = "Spoke-proxy URL using Route53 DNS"
  value       = try("wss://${aws_route53_record.spoke_proxy[0].fqdn}:443", "")
}

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
        hostname: "spoke-proxy.${aws_eip.spoke_proxy[0].public_ip}.nip.io"

    k8sAgent:
      env:
        AEGIS_PROXY_INGRESS_HOST: "spoke-proxy.${aws_eip.spoke_proxy[0].public_ip}.nip.io:443"
  EOF
  , "")
}
