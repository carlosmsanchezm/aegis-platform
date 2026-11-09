terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.60"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

locals {
  user_name                = var.iam_user_name != "" ? var.iam_user_name : "aegis-pulumi-provisioner"
  pulumi_state_bucket_arn  = "arn:aws:s3:::${var.pulumi_state_bucket}"
  pulumi_state_objects_arn = "arn:aws:s3:::${var.pulumi_state_bucket}/${var.pulumi_state_prefix}/*"
  pulumi_meta_objects_arn  = "arn:aws:s3:::${var.pulumi_state_bucket}/.pulumi/*"
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

data "aws_iam_policy_document" "pulumi" {
  statement {
    effect = "Allow"

    actions = [
      "sts:AssumeRole",
    ]

    resources = [
      var.target_role_arn,
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:ListBucket",
      "s3:ListBucketMultipartUploads",
    ]

    resources = [
      local.pulumi_state_bucket_arn,
    ]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values = [
        "${var.pulumi_state_prefix}/*",
        var.pulumi_state_prefix,
      ]
    }
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:AbortMultipartUpload",
    ]

    resources = [
      local.pulumi_state_objects_arn,
      local.pulumi_meta_objects_arn,
    ]
  }
}

resource "aws_iam_user_policy" "pulumi" {
  name   = "${local.user_name}-assume-role"
  user   = aws_iam_user.pulumi.name
  policy = data.aws_iam_policy_document.pulumi.json
}

resource "aws_iam_access_key" "pulumi" {
  user = aws_iam_user.pulumi.name
}
