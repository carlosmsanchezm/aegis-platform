terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.60"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

locals {
  user_name = var.iam_user_name != "" ? var.iam_user_name : "aegis-pulumi-provisioner"
}

resource "aws_iam_user" "pulumi" {
  name = local.user_name
  path = "/service/"

  tags = merge(
    var.tags,
    {
      "Service" = "aegis-platform"
      "Managed" = "terraform"
    }
  )
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    actions = [
      "sts:AssumeRole",
    ]

    resources = [
      var.target_role_arn,
    ]
  }
}

resource "aws_iam_user_policy" "pulumi" {
  name   = "${local.user_name}-assume-role"
  user   = aws_iam_user.pulumi.name
  policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_access_key" "pulumi" {
  user = aws_iam_user.pulumi.name
}

resource "aws_iam_user_policy_attachment" "pulumi_admin" {
  user       = aws_iam_user.pulumi.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}
