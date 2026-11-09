variable "aws_region" {
  type        = string
  description = "AWS region for the relay infrastructure."
  default     = "us-east-1"
}

variable "aws_profile" {
  type        = string
  description = "AWS CLI profile to use."
  default     = "default"
}

variable "relay_ssh_public_key" {
  type        = string
  description = "SSH public key for relay access. Generate with: ssh-keygen -t ed25519 -f ~/.ssh/aegis-relay -N ''"
  default     = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIN+yI36ozEF9TcJEGMmGOe5Fg7nZDFNH1mTI5spat9fP aegis-relay"
}

variable "tags" {
  type        = map(string)
  description = "Additional tags to apply to all resources."
  default = {
    "Project" = "aegis-platform"
    "Env"     = "dev"
  }
}
