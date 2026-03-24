locals {
  relay_name = "aegis-dev-relay"
  relay_tags = merge(var.tags, {
    "Name"    = local.relay_name
    "Purpose" = "aegis-development-relay"
    "Managed" = "terraform"
  })
}

# ---------------------------------------------------------------------------
# Data sources
# ---------------------------------------------------------------------------

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
  filter {
    name   = "default-for-az"
    values = ["true"]
  }
}

data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# ---------------------------------------------------------------------------
# Security Group
# ---------------------------------------------------------------------------

resource "aws_security_group" "relay" {
  name        = "${local.relay_name}-sg"
  description = "Security group for aegis development relay"
  vpc_id      = data.aws_vpc.default.id

  # SSH from anywhere (for establishing reverse tunnel from local machine)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "SSH access for reverse tunnel"
  }

  # gRPC port for platform-api (forwarded from local via SSH tunnel)
  ingress {
    from_port   = 8081
    to_port     = 8081
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.default.cidr_block]
    description = "gRPC platform-api from VPC (EKS nodes)"
  }

  # HTTPS port for keycloak (forwarded from local via SSH tunnel)
  ingress {
    from_port   = 8443
    to_port     = 8443
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.default.cidr_block]
    description = "HTTPS keycloak from VPC (EKS nodes)"
  }

  # HTTPS port for step-ca (forwarded from local via SSH tunnel)
  ingress {
    from_port   = 9443
    to_port     = 9443
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.default.cidr_block]
    description = "HTTPS step-ca from VPC (EKS nodes)"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = local.relay_tags
}

# ---------------------------------------------------------------------------
# SSH Key Pair
# ---------------------------------------------------------------------------

resource "aws_key_pair" "relay" {
  key_name   = "${local.relay_name}-key"
  public_key = var.relay_ssh_public_key

  tags = local.relay_tags
}

# ---------------------------------------------------------------------------
# IAM Role (for SSM access)
# ---------------------------------------------------------------------------

resource "aws_iam_role" "relay" {
  name = "${local.relay_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = local.relay_tags
}

resource "aws_iam_role_policy_attachment" "relay_ssm" {
  role       = aws_iam_role.relay.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "relay" {
  name = "${local.relay_name}-profile"
  role = aws_iam_role.relay.name

  tags = local.relay_tags
}

# ---------------------------------------------------------------------------
# EC2 Instance
# ---------------------------------------------------------------------------

resource "aws_instance" "relay" {
  ami                         = data.aws_ami.amazon_linux_2023.id
  instance_type               = "t3.micro"
  key_name                    = aws_key_pair.relay.key_name
  vpc_security_group_ids      = [aws_security_group.relay.id]
  subnet_id                   = tolist(data.aws_subnets.default.ids)[0]
  associate_public_ip_address = true
  iam_instance_profile        = aws_iam_instance_profile.relay.name

  user_data = <<-EOF
    #!/bin/bash
    set -e

    # Update system
    dnf update -y

    # Configure SSH for reverse tunneling
    cat >> /etc/ssh/sshd_config << 'SSHCONFIG'
    # Allow reverse port forwarding
    GatewayPorts yes
    AllowTcpForwarding yes
    # Keep connections alive
    ClientAliveInterval 30
    ClientAliveCountMax 3
    SSHCONFIG

    # Restart SSH
    systemctl restart sshd

    # Create a marker file to indicate setup is complete
    touch /tmp/relay-setup-complete

    echo "Aegis dev relay setup complete"
  EOF

  root_block_device {
    volume_size           = 8
    volume_type           = "gp3"
    delete_on_termination = true
  }

  tags = merge(local.relay_tags, {
    "Name" = local.relay_name
  })

  lifecycle {
    ignore_changes = [ami]
  }
}

# ---------------------------------------------------------------------------
# Elastic IP (stable address for SSH tunnel)
# ---------------------------------------------------------------------------

resource "aws_eip" "relay" {
  instance = aws_instance.relay.id
  domain   = "vpc"

  tags = local.relay_tags
}

# ---------------------------------------------------------------------------
# Internal NLB (TCP passthrough for gRPC + Keycloak)
# ---------------------------------------------------------------------------

resource "aws_lb" "relay" {
  name               = "${local.relay_name}-nlb"
  internal           = true
  load_balancer_type = "network"
  subnets            = data.aws_subnets.default.ids

  enable_cross_zone_load_balancing = true

  tags = local.relay_tags
}

# Target Group: platform-api gRPC (port 8081)
resource "aws_lb_target_group" "platform_api" {
  name        = "${local.relay_name}-grpc"
  port        = 8081
  protocol    = "TCP"
  vpc_id      = data.aws_vpc.default.id
  target_type = "instance"

  health_check {
    enabled             = true
    protocol            = "TCP"
    port                = 8081
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
  }

  tags = local.relay_tags
}

# Target Group: Keycloak HTTPS (port 8443)
resource "aws_lb_target_group" "keycloak" {
  name        = "${local.relay_name}-kc"
  port        = 8443
  protocol    = "TCP"
  vpc_id      = data.aws_vpc.default.id
  target_type = "instance"

  health_check {
    enabled             = true
    protocol            = "TCP"
    port                = 8443
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
  }

  tags = local.relay_tags
}

# Target Group: step-ca HTTPS (port 9443)
resource "aws_lb_target_group" "step_ca" {
  name        = "${local.relay_name}-stepca"
  port        = 9443
  protocol    = "TCP"
  vpc_id      = data.aws_vpc.default.id
  target_type = "instance"

  health_check {
    enabled             = true
    protocol            = "TCP"
    port                = 9443
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
  }

  tags = local.relay_tags
}

# Register EC2 instance with target groups
resource "aws_lb_target_group_attachment" "platform_api" {
  target_group_arn = aws_lb_target_group.platform_api.arn
  target_id        = aws_instance.relay.id
  port             = 8081
}

resource "aws_lb_target_group_attachment" "keycloak" {
  target_group_arn = aws_lb_target_group.keycloak.arn
  target_id        = aws_instance.relay.id
  port             = 8443
}

resource "aws_lb_target_group_attachment" "step_ca" {
  target_group_arn = aws_lb_target_group.step_ca.arn
  target_id        = aws_instance.relay.id
  port             = 9443
}

# NLB Listeners
resource "aws_lb_listener" "platform_api" {
  load_balancer_arn = aws_lb.relay.arn
  port              = 8081
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.platform_api.arn
  }

  tags = local.relay_tags
}

resource "aws_lb_listener" "keycloak" {
  load_balancer_arn = aws_lb.relay.arn
  port              = 8443
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.keycloak.arn
  }

  tags = local.relay_tags
}

resource "aws_lb_listener" "step_ca" {
  load_balancer_arn = aws_lb.relay.arn
  port              = 9443
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.step_ca.arn
  }

  tags = local.relay_tags
}

# ---------------------------------------------------------------------------
# Auto-generate Helm values with correct NLB endpoint
# ---------------------------------------------------------------------------

# Hub overlay: NLB-dependent env vars for platform-api.
# This file is layered automatically by scripts/aegis.sh when it exists.
resource "local_file" "aegis_hub_relay_values" {
  filename = "${path.module}/../../charts/aegis-services/values/local-relay.yaml"
  content  = <<-YAML
    # Auto-generated by terraform/dev-relay — DO NOT EDIT MANUALLY
    # Generated at: ${formatdate("YYYY-MM-DD'T'HH:mm:ss'Z'", timestamp())}
    #
    # This overlay sets NLB-dependent env vars on platform-api so the hub
    # knows where provisioned spoke clusters should connect back to.
    # It is automatically layered by scripts/aegis.sh when the file exists.
    platformApi:
      env:
        AEGIS_PLATFORM_API_ENDPOINT: "${aws_lb.relay.dns_name}:8081"
        AEGIS_STEP_CA_URL: "https://${aws_lb.relay.dns_name}:9443"
        AEGIS_SPOKE_OIDC_TOKEN_URL: "https://${aws_lb.relay.dns_name}:8443/realms/aegis/protocol/openid-connect/token"
  YAML
}

resource "local_file" "step_ca_dns_names" {
  filename = "${path.module}/../../.step-ca-external-dns"
  content  = aws_lb.relay.dns_name
}

resource "local_file" "aegis_spoke_values" {
  filename = "${path.module}/../../charts/aegis-spoke/values-aws-relay.yaml"
  content  = <<-YAML
    # Auto-generated by terraform/dev-relay — DO NOT EDIT MANUALLY
    # Generated at: ${formatdate("YYYY-MM-DD'T'HH:mm:ss'Z'", timestamp())}
    # AWS Account: 195714074609
    #
    # Architecture — all traffic through one NLB relay:
    #   gRPC:     ${aws_lb.relay.dns_name}:8081 → SSH tunnel → local platform-api
    #   Keycloak: ${aws_lb.relay.dns_name}:8443 → SSH tunnel → local keycloak
    #   step-ca:  ${aws_lb.relay.dns_name}:9443 → SSH tunnel → local step-ca
    # Relay public IP (SSH tunnel target): ${aws_eip.relay.public_ip}
    #
    # Usage:
    #   helm upgrade aegis-spoke ./charts/aegis-spoke -n aegis-system \
    #     --reset-values \
    #     -f charts/aegis-spoke/values.yaml \
    #     -f charts/aegis-spoke/values-aws-relay.yaml \
    #     --set k8sAgent.env.AEGIS_CLUSTER_ID=<cluster-id> \
    #     --set k8sAgent.env.AEGIS_REGION=us-east-1 \
    #     --set k8sAgent.env.AEGIS_PROVIDER=aws

    k8sAgent:
      image:
        repository: 195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/k8s-agent
        pullPolicy: Always
        tag: "dev"
      env:
        # gRPC to platform-api via internal NLB (TCP passthrough, preserves HTTP/2)
        AEGIS_CP_GRPC: "${aws_lb.relay.dns_name}:8081"
        AEGIS_CP_GRPC_INSECURE: "false"
        AEGIS_CP_GRPC_SKIP_VERIFY: "true"
        AEGIS_CP_GRPC_SERVER_NAME: ""
        AEGIS_FLAVORS: "cpu-small"
        AEGIS_DEFAULT_IMAGE: "195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:latest"

        # OIDC via internal NLB (TCP passthrough to local Keycloak)
        AEGIS_CP_OIDC_TOKEN_URL: "https://${aws_lb.relay.dns_name}:8443/realms/aegis/protocol/openid-connect/token"
        AEGIS_CP_OIDC_CLIENT_ID: "spoke-agent"
        AEGIS_CP_OIDC_CLIENT_SECRET: ""  # REQUIRED: Set via --set or values override
        AEGIS_CP_OIDC_AUDIENCE: "aegis-platform"
        AEGIS_CP_OIDC_SKIP_TLS_VERIFY: "true"

        # Skip CA verification (self-signed certs through tunnel)
        AEGIS_PLATFORM_CA_B64: ""
        AEGIS_CP_OIDC_CA_B64: ""

    # Spoke proxy — enabled for the proxy URL fix test
    # Uses LoadBalancer Service (auto-creates NLB with stable DNS)
    # k8s-agent auto-discovers the proxy NLB hostname
    proxy:
      enabled: true
      image:
        repository: 195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/proxy
        pullPolicy: Always
        tag: "dev"
      service:
        type: LoadBalancer
        port: 443
        annotations:
          service.beta.kubernetes.io/aws-load-balancer-type: "external"
          service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: "instance"
          service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
      # url intentionally empty — agent auto-discovers from LoadBalancer Service
      url: ""
      ingress:
        enabled: false
        hostname: ""
      jwtSecret: "a-very-secret-key-for-local-dev-must-be-32-chars"
      tls:
        terminateAtIngress: false
  YAML
}
