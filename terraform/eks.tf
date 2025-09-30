# EKS Cluster Configuration using AWS EKS module

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = local.cluster_name
  cluster_version = var.cluster_version

  # Network configuration
  vpc_id                    = module.vpc.vpc_id
  subnet_ids                = module.vpc.private_subnets
  control_plane_subnet_ids  = module.vpc.private_subnets

  # Cluster endpoint configuration
  cluster_endpoint_public_access  = true
  cluster_endpoint_private_access = true
  cluster_endpoint_public_access_cidrs = ["0.0.0.0/0"]

  # OIDC Identity provider
  enable_irsa = true

  # Enable cluster creator admin permissions
  enable_cluster_creator_admin_permissions = true

  # EKS Managed Node Groups
  eks_managed_node_groups = {
    # CPU Workers (similar to your t3.medium setup)
    cpu_workers = {
      name            = "cpu-workers"
      instance_types  = [var.cpu_instance_type]
      capacity_type   = var.use_spot_instances ? "SPOT" : "ON_DEMAND"

      min_size        = 1
      max_size        = 3
      desired_size    = var.cpu_desired_capacity

      # Launch template configuration
      create_launch_template = true
      launch_template_name   = "${local.cluster_name}-cpu-workers"

      # Use latest AMI
      ami_type = "AL2_x86_64"

      # Instance configuration
      disk_size = 50

      # Network configuration
      subnet_ids = module.vpc.private_subnets

      # Security groups
      vpc_security_group_ids = [aws_security_group.eks_nodes.id]

      # Labels
      labels = {
        Environment = var.environment
        NodeType    = "cpu-worker"
      }

      # Taints - none for CPU workers
      taints = {}

      tags = merge(local.common_tags, {
        Name = "${local.cluster_name}-cpu-workers"
      })
    }

    # GPU Workers (similar to your g4dn.xlarge setup)
    gpu_workers = {
      name            = "gpu-workers-g4"
      instance_types  = [var.gpu_instance_type]
      capacity_type   = var.use_spot_instances ? "SPOT" : "ON_DEMAND"

      min_size        = 0
      max_size        = var.gpu_max_capacity
      desired_size    = var.gpu_desired_capacity

      # Launch template configuration
      create_launch_template = true
      launch_template_name   = "${local.cluster_name}-gpu-workers"

      # Use GPU-optimized AMI
      ami_type = "AL2_x86_64_GPU"

      # Instance configuration
      disk_size = 100  # Larger disk for GPU workloads

      # Network configuration
      subnet_ids = module.vpc.private_subnets

      # Security groups
      vpc_security_group_ids = [aws_security_group.eks_nodes.id]

      # Labels (matching your eksctl config)
      labels = {
        Environment = var.environment
        NodeType    = "gpu-worker"
        "aegis.io/gpu-flavor" = "nvidia-tesla-t4"
      }

      # Taints (matching your eksctl config)
      taints = [
        {
          key    = "nvidia.com/gpu"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      ]

      tags = merge(local.common_tags, {
        Name = "${local.cluster_name}-gpu-workers"
      })
    }
  }

  # Cluster security group
  create_cluster_security_group = true
  cluster_security_group_name   = "${local.cluster_name}-cluster-sg"

  # Node security group
  create_node_security_group = true
  node_security_group_name   = "${local.cluster_name}-node-sg"

  # EKS Addons
  cluster_addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
      configuration_values = jsonencode({
        env = {
          ENABLE_PREFIX_DELEGATION = "true"
          WARM_PREFIX_TARGET       = "1"
        }
      })
    }
    aws-ebs-csi-driver = {
      most_recent = true
    }
  }

  # Enable cluster logging
  cluster_enabled_log_types = ["api", "audit", "authenticator"]

  tags = local.common_tags
}

# Additional security group for EKS nodes
resource "aws_security_group" "eks_nodes" {
  name_prefix = "${local.cluster_name}-nodes-"
  vpc_id      = module.vpc.vpc_id

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }


  tags = merge(local.common_tags, {
    Name = "${local.cluster_name}-nodes-sg"
  })
}