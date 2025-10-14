{{/*
Expand the name of the chart.
*/}}
{{- define "aegis-services.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "aegis-services.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "aegis-services.labels" -}}
app.kubernetes.io/name: {{ include "aegis-services.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "aegis-services.selectorLabels" -}}
app.kubernetes.io/name: {{ include "aegis-services.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/* Platform API component names */}}
{{- define "aegis-services.platformApi.fullname" -}}
{{- printf "%s-platform-api" (include "aegis-services.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "aegis-services.platformApi.secretName" -}}
{{- printf "%s-platform-api-secret" (include "aegis-services.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "aegis-services.platformApi.tlsSecretName" -}}
{{- printf "%s-platform-api-tls" (include "aegis-services.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Proxy component names */}}
{{- define "aegis-services.proxy.fullname" -}}
{{- printf "%s-proxy" (include "aegis-services.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "aegis-services.proxy.secretName" -}}
{{- printf "%s-proxy-secret" (include "aegis-services.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "aegis-services.proxy.tlsSecretName" -}}
{{- printf "%s-proxy-tls" (include "aegis-services.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Component selector labels */}}
{{- define "aegis-services.platformApi.selectorLabels" -}}
{{ include "aegis-services.selectorLabels" . }}
app.kubernetes.io/component: platform-api
{{- end -}}

{{- define "aegis-services.proxy.selectorLabels" -}}
{{ include "aegis-services.selectorLabels" . }}
app.kubernetes.io/component: proxy
{{- end -}}

{{- define "aegis-services.proxy.tlsIngress" -}}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "aegis-services.proxy.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "aegis-services.labels" . | nindent 4 }}
    app.kubernetes.io/component: proxy
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    nginx.ingress.kubernetes.io/proxy-ssl-secret: "{{ .Release.Namespace }}/{{ include "aegis-services.proxy.tlsSecretName" . }}"
    nginx.ingress.kubernetes.io/proxy-ssl-verify: "off"
    nginx.ingress.kubernetes.io/proxy-ssl-server-name: "{{ .Values.proxy.publicHost }}"
    {{- with .Values.proxy.ingress.annotations }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
spec:
  ingressClassName: {{ .Values.proxy.ingress.className | default "ingress-nginx" }}
  tls:
    - secretName: {{ include "aegis-services.proxy.tlsSecretName" . }}
      hosts:
        - {{ .Values.proxy.publicHost }}
  rules:
    - host: {{ .Values.proxy.publicHost }}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: {{ include "aegis-services.proxy.fullname" . }}
                port:
                  number: {{ .Values.proxy.service.port }}
{{- end -}}
