{{/*
Validation helpers for aegis-spoke chart
*/}}

{{/*
Validate proxy configuration.
Remote spoke clusters must have either:
1. proxy.enabled=true (deploys spoke-proxy pod)
2. OR proxy.url set (uses external proxy)

Without one of these, VS Code/SSH connections will fail because the k8s-agent
won't report a proxy_url in heartbeats, causing platform-api to fall back to
the hub proxy which cannot reach workloads on remote clusters.
*/}}
{{- define "aegis-spoke.validateProxyConfig" -}}
{{- if not .Values.proxy.enabled }}
{{- if not .Values.proxy.url }}
{{- $localHosts := list "localhost" "127.0.0.1" "host.docker.internal" "localtest.me" "svc.cluster.local" }}
{{- $isLocal := false }}
{{- $cpGrpc := .Values.k8sAgent.env.AEGIS_CP_GRPC | default "" }}
{{- range $localHosts }}
{{- if contains . $cpGrpc }}
{{- $isLocal = true }}
{{- end }}
{{- end }}
{{- if not $isLocal }}
{{- fail "\n\nERROR: aegis-spoke proxy misconfiguration detected!\n\nFor remote/cloud spoke clusters, you must either:\n  1. Set proxy.enabled=true (recommended) to deploy the spoke-proxy pod\n  2. OR set proxy.url to an external proxy URL\n\nWithout a proxy configuration, VS Code/SSH connections will fail because\nworkloads on remote clusters are not reachable from the hub proxy.\n\nTo fix:\n  helm upgrade aegis-spoke charts/aegis-spoke \\\n    --set proxy.enabled=true \\\n    --set proxy.ingress.hostname=spoke-proxy.YOUR-IP.nip.io:31484 \\\n    ... other flags\n\nIf this is intentionally a local development deployment, you can:\n  - Set k8sAgent.env.AEGIS_CP_GRPC to include 'localhost' or 'host.docker.internal'\n  - OR set proxy.url to your hub proxy URL (e.g., wss://proxy.localtest.me)" }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
