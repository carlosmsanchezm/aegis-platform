bucket         = "aegis-tf-state-123456789012"
key            = "aegis/prod/terraform.tfstate"
region         = "us-east-1"
encrypt        = true
dynamodb_table = "aegis-terraform-locks"
