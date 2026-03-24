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
  value       = "ssh -i ~/.ssh/aegis-relay -R 0.0.0.0:8081:localhost:8081 -R 0.0.0.0:8443:localhost:8443 -R 0.0.0.0:9443:localhost:9443 -N ec2-user@${aws_eip.relay.public_ip}"
}

output "next_steps" {
  description = "Steps to complete the setup after terraform apply"
  value       = <<-EOT

    === SETUP COMPLETE ===

    1. Generate self-signed TLS certs for spoke proxy (one-time):

       openssl req -x509 -newkey rsa:2048 \
         -keyout /tmp/proxy-key.pem -out /tmp/proxy-cert.pem \
         -days 365 -nodes -subj "/CN=spoke-proxy.aegis.local"

    2. Start the SSH tunnel (keep running in a terminal):

       ./scripts/start-aws-tunnel.sh

    3. Deploy spoke with proxy URL fix:

       cd ${abspath(path.module)}/../.. && \
       helm upgrade aegis-spoke ./charts/aegis-spoke -n aegis-system \
         --kube-context eks-e2e-pilot \
         --reset-values \
         -f charts/aegis-spoke/values.yaml \
         -f charts/aegis-spoke/values-aws-relay.yaml \
         --set k8sAgent.env.AEGIS_CLUSTER_ID=e2e-pilot-test-us-east-1-atlas-train-govcloud-22-e38bdff4 \
         --set k8sAgent.env.AEGIS_REGION=us-east-1 \
         --set k8sAgent.env.AEGIS_CP_OIDC_CLIENT_SECRET=rEC99sBBWQAbRgg0xRQFBsMC8rt6pZOB \
         --set k8sAgent.env.AEGIS_PROVIDER=aws \
         --set-file proxy.tls.cert=/tmp/proxy-cert.pem \
         --set-file proxy.tls.key=/tmp/proxy-key.pem

    4. Verify:
       - Agent logs: kubectl logs -l app.kubernetes.io/component=k8s-agent -n aegis-system --context eks-e2e-pilot --tail=20
       - Platform-API: kubectl logs -l app.kubernetes.io/component=platform-api -n aegis-system --context docker-desktop --tail=20 | grep "cluster registered"

  EOT
}
