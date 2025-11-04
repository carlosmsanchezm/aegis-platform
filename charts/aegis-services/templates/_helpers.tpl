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

{{- define "aegis-services.platformApi.oidcCASecretName" -}}
{{- $platform := .Values.platformApi | default dict -}}
{{- $auth := $platform.auth | default dict -}}
{{- $oidc := $auth.oidc | default dict -}}
{{- $ca := $oidc.caBundle | default dict -}}
{{- if $ca.secretName -}}
{{- $ca.secretName | trunc 63 | trimSuffix "-" -}}
{{- else if $ca.create -}}
{{- printf "%s-oidc-ca" (include "aegis-services.platformApi.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- "" -}}
{{- end -}}
{{- end -}}

{{- define "aegis-services.platformApi.oidcCAMountPath" -}}
{{- $platform := .Values.platformApi | default dict -}}
{{- $auth := $platform.auth | default dict -}}
{{- $oidc := $auth.oidc | default dict -}}
{{- $ca := $oidc.caBundle | default dict -}}
{{- default "/etc/aegis-platform-api/oidc" $ca.mountPath -}}
{{- end -}}

{{- define "aegis-services.platformApi.oidcCAFileName" -}}
{{- $platform := .Values.platformApi | default dict -}}
{{- $auth := $platform.auth | default dict -}}
{{- $oidc := $auth.oidc | default dict -}}
{{- $ca := $oidc.caBundle | default dict -}}
{{- default "ca.crt" $ca.fileName -}}
{{- end -}}

{{- define "aegis-services.platformApi.oidcCAKey" -}}
{{- $platform := .Values.platformApi | default dict -}}
{{- $auth := $platform.auth | default dict -}}
{{- $oidc := $auth.oidc | default dict -}}
{{- $ca := $oidc.caBundle | default dict -}}
{{- default "ca.crt" $ca.key -}}
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

{{/* Backstage component names */}}
{{- define "aegis-services.backstage.fullname" -}}
{{- printf "%s-backstage" (include "aegis-services.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "aegis-services.backstage.selectorLabels" -}}
{{ include "aegis-services.selectorLabels" . }}
app.kubernetes.io/component: backstage
{{- end -}}

{{- define "aegis-services.backstage.labels" -}}
{{ include "aegis-services.labels" . }}
app.kubernetes.io/component: backstage
{{- end -}}

{{- define "aegis-services.backstage.serviceName" -}}
{{- include "aegis-services.backstage.fullname" . -}}
{{- end -}}

{{- define "aegis-services.backstage.caBundleSecretName" -}}
{{- $name := .Values.backstage.caBundle.secretName | default (printf "%s-backstage-ca" (include "aegis-services.fullname" .)) -}}
{{- $name | trunc 63 | trimSuffix "-" -}}
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

{{/* Keycloak helpers */}}
{{- define "aegis-services.keycloak.namespace" -}}
{{- $kc := .Values.keycloak | default dict }}
{{- if $kc.namespace -}}
{{- $kc.namespace -}}
{{- else -}}
{{- .Release.Namespace -}}
{{- end -}}
{{- end -}}

{{- define "aegis-services.keycloak.fullname" -}}
{{- printf "%s-keycloak" (include "aegis-services.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "aegis-services.keycloak.serviceName" -}}
{{- include "aegis-services.keycloak.fullname" . -}}
{{- end -}}

{{- define "aegis-services.keycloak.labels" -}}
{{ include "aegis-services.labels" . }}
app.kubernetes.io/component: keycloak
{{- end -}}

{{- define "aegis-services.keycloak.selectorLabels" -}}
{{ include "aegis-services.selectorLabels" . }}
app.kubernetes.io/component: keycloak
{{- end -}}

{{- define "aegis-services.keycloak.postgresFullname" -}}
{{- printf "%s-db" (include "aegis-services.keycloak.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "aegis-services.keycloak.postgresServiceName" -}}
{{- include "aegis-services.keycloak.postgresFullname" . -}}
{{- end -}}

{{- define "aegis-services.keycloak.postgres.labels" -}}
{{ include "aegis-services.labels" . }}
app.kubernetes.io/component: keycloak-postgres
{{- end -}}

{{- define "aegis-services.keycloak.postgres.selectorLabels" -}}
{{ include "aegis-services.selectorLabels" . }}
app.kubernetes.io/component: keycloak-postgres
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
