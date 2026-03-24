# Cloudflare DNS records for aegis-platform.tech
# Points to the K8s LoadBalancer NLB DNS names after deployment
#
# DNS content is managed by generate-cloud-deployment.sh (direct API calls)
# after LoadBalancers are provisioned. lifecycle.ignore_changes prevents
# standalone `terraform apply` from reverting records to placeholder.

data "cloudflare_zone" "aegis" {
  name = "aegis-platform.tech"
}

resource "cloudflare_record" "platform_api" {
  zone_id = data.cloudflare_zone.aegis.id
  name    = "platform-api"
  type    = "CNAME"
  content = var.platform_api_lb_hostname != "" ? var.platform_api_lb_hostname : "placeholder.elb.amazonaws.com"
  proxied = false # TCP passthrough needed for gRPC

  lifecycle {
    ignore_changes = [content]
  }
}

resource "cloudflare_record" "keycloak" {
  zone_id = data.cloudflare_zone.aegis.id
  name    = "keycloak"
  type    = "CNAME"
  content = var.keycloak_lb_hostname != "" ? var.keycloak_lb_hostname : "placeholder.elb.amazonaws.com"
  proxied = true

  lifecycle {
    ignore_changes = [content]
  }
}

resource "cloudflare_record" "proxy" {
  zone_id = data.cloudflare_zone.aegis.id
  name    = "proxy"
  type    = "CNAME"
  content = var.proxy_lb_hostname != "" ? var.proxy_lb_hostname : "placeholder.elb.amazonaws.com"
  proxied = false

  lifecycle {
    ignore_changes = [content]
  }
}

resource "cloudflare_record" "ui" {
  zone_id = data.cloudflare_zone.aegis.id
  name    = "ui"
  type    = "CNAME"
  content = var.ui_lb_hostname != "" ? var.ui_lb_hostname : "placeholder.elb.amazonaws.com"
  proxied = true

  lifecycle {
    ignore_changes = [content]
  }
}

################################################################################
# Outputs
################################################################################

output "cloudflare_zone_id" {
  description = "Cloudflare zone ID for aegis-platform.tech"
  value       = data.cloudflare_zone.aegis.id
}

output "cloudflare_api_hostname" {
  description = "Cloudflare DNS hostname for platform-api"
  value       = cloudflare_record.platform_api.hostname
}

output "cloudflare_keycloak_hostname" {
  description = "Cloudflare DNS hostname for Keycloak"
  value       = cloudflare_record.keycloak.hostname
}

output "cloudflare_proxy_hostname" {
  description = "Cloudflare DNS hostname for proxy"
  value       = cloudflare_record.proxy.hostname
}

output "cloudflare_ui_hostname" {
  description = "Cloudflare DNS hostname for aegis-ui"
  value       = cloudflare_record.ui.hostname
}
