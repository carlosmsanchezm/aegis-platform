# AWS Development Relay Infrastructure
# This creates an EC2 instance that acts as a reverse SSH tunnel relay
# to allow AWS spoke clusters to connect to your local platform-api
#
# Architecture:
#   Local Machine (platform-api, keycloak)
#         |
#         | SSH Reverse Tunnel (-R 8081:localhost:8081)
#         v
#   EC2 Instance (relay)
#         |
#         | NLB
#         v
#   EKS Spoke Cluster
#
# Usage:
#   cd terraform/pulumi-stack
#   terraform apply -target=module.dev_relay
#
#   # Then run the SSH tunnel script:
#   ../scripts/start-aws-tunnel.sh

locals {
  relay_name = "aegis-dev-relay"
  relay_tags = merge(var.tags, {
    "Name"    = local.relay_name
    "Purpose" = "aegis-development-relay"
    "Managed" = "terraform"
  })
}

# Data sources
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

# Security Group for the relay
resource "aws_security_group" "relay" {
  name        = "${local.relay_name}-sg"
  description = "Security group for aegis development relay"
  vpc_id      = data.aws_vpc.default.id

  # SSH from anywhere (for establishing reverse tunnel)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "SSH access for reverse tunnel"
  }

  # gRPC port for platform-api (forwarded from local)
  ingress {
    from_port   = 8081
    to_port     = 8081
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.default.cidr_block]
    description = "gRPC platform-api from VPC"
  }

  # HTTPS port for keycloak (forwarded from local)
  ingress {
    from_port   = 8443
    to_port     = 8443
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.default.cidr_block]
    description = "HTTPS keycloak from VPC"
  }

  # Outbound
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = local.relay_tags
}

# SSH Key Pair
resource "aws_key_pair" "relay" {
  key_name   = "${local.relay_name}-key"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIN+yI36ozEF9TcJEGMmGOe5Fg7nZDFNH1mTI5spat9fP aegis-relay"

  tags = local.relay_tags
}

# IAM Role for the EC2 instance
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

# EC2 Instance
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

# Elastic IP for stable address
resource "aws_eip" "relay" {
  instance = aws_instance.relay.id
  domain   = "vpc"

  tags = local.relay_tags
}

# NLB for the relay (TCP passthrough, preserves gRPC)
resource "aws_lb" "relay" {
  name               = "${local.relay_name}-nlb"
  internal           = true
  load_balancer_type = "network"
  subnets            = data.aws_subnets.default.ids

  enable_cross_zone_load_balancing = true

  tags = local.relay_tags
}

# Target Group for platform-api (gRPC)
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

# Target Group for keycloak (HTTPS)
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

# Outputs
output "relay_public_ip" {
  description = "Public IP of the relay instance (for SSH tunnel)"
  value       = aws_eip.relay.public_ip
}

output "relay_nlb_dns" {
  description = "DNS name of the internal NLB"
  value       = aws_lb.relay.dns_name
}

output "relay_platform_api_endpoint" {
  description = "Endpoint for platform-api (gRPC)"
  value       = "${aws_lb.relay.dns_name}:8081"
}

output "relay_keycloak_endpoint" {
  description = "Endpoint for keycloak (HTTPS)"
  value       = "${aws_lb.relay.dns_name}:8443"
}

output "ssh_tunnel_command" {
  description = "Command to establish SSH reverse tunnel"
  value       = "ssh -i ~/.ssh/aegis-relay -R 0.0.0.0:8081:localhost:8081 -R 0.0.0.0:8443:localhost:8443 -N ec2-user@${aws_eip.relay.public_ip}"
}

# Render the spoke values file with the current NLB endpoints so we don't have to
# hard-code the relay host. This runs as part of terraform apply.
resource "local_file" "aegis_spoke_values" {
  filename = "${path.module}/../../charts/aegis-spoke/values-aws-relay.yaml"
  content  = <<-EOF
    # Auto-generated by terraform/pulumi-stack (dev relay)
    # Generated at: ${formatdate("YYYY-MM-DD'T'HH:mm:ss'Z'", timestamp())}
    #
    # NLB DNS: ${aws_lb.relay.dns_name}
    # Platform API: ${aws_lb.relay.dns_name}:8081
    # Keycloak: ${aws_lb.relay.dns_name}:8443
    # Relay public IP (SSH tunnel target): ${aws_eip.relay.public_ip}
    #
    # Usage:
    #   helm upgrade aegis-spoke ./charts/aegis-spoke -n aegis-system \\
    #     -f charts/aegis-spoke/values.yaml \\
    #     -f charts/aegis-spoke/values-aws-relay.yaml \\
    #     --set k8sAgent.env.AEGIS_CLUSTER_ID=<cluster-id> \\
    #     --set k8sAgent.env.AEGIS_REGION=us-east-1 \\
    #     --set k8sAgent.env.AEGIS_PROVIDER=aws

    k8sAgent:
      image:
        pullPolicy: Always
        tag: "dev"
      env:
        # gRPC to platform-api via AWS NLB (TCP passthrough, preserves HTTP/2)
        AEGIS_CP_GRPC: "${aws_lb.relay.dns_name}:8081"
        AEGIS_CP_GRPC_INSECURE: "false"
        AEGIS_CP_GRPC_SKIP_VERIFY: "true"
        AEGIS_CP_GRPC_SERVER_NAME: ""
        AEGIS_FLAVORS: "cpu-small,t4-1gpu"
        AEGIS_DEFAULT_IMAGE: "docker.io/carlosmsanchez/aegis-workspace-vscode:latest"

        # OIDC client credentials (Keycloak via AWS NLB)
        AEGIS_CP_OIDC_TOKEN_URL: "https://${aws_lb.relay.dns_name}:8443/realms/aegis/protocol/openid-connect/token"
        AEGIS_CP_OIDC_CLIENT_ID: "spoke-agent"
        AEGIS_CP_OIDC_CLIENT_SECRET: "rEC99sBBWQAbRgg0xRQFBsMC8rt6pZOB"
        AEGIS_CP_OIDC_AUDIENCE: "aegis-platform"
        AEGIS_CP_OIDC_SKIP_TLS_VERIFY: "true"

        # Skip CA verification (self-signed certs through tunnel)
        AEGIS_PLATFORM_CA_B64: ""
        AEGIS_CP_OIDC_CA_B64: ""

    proxy:
      enabled: false
  EOF
}
