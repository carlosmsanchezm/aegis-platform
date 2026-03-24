package kubeclients

import (
	"strings"
	"testing"
)

const kubeconfigWithProfile = `apiVersion: v1
clusters:
- cluster:
    server: https://example.eks.amazonaws.com
    certificate-authority-data: dGVzdA==
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-cluster
  name: test-ctx
current-context: test-ctx
kind: Config
users:
- name: test-cluster
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws
      args:
      - eks
      - get-token
      - --cluster-name
      - test-cluster
      env:
      - name: AWS_PROFILE
        value: aegis-new
      - name: AWS_DEFAULT_REGION
        value: us-east-1
`

const kubeconfigClean = `apiVersion: v1
clusters:
- cluster:
    server: https://example.eks.amazonaws.com
    certificate-authority-data: dGVzdA==
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-cluster
  name: test-ctx
current-context: test-ctx
kind: Config
users:
- name: test-cluster
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws
      args:
      - eks
      - get-token
      - --cluster-name
      - test-cluster
      env:
      - name: AWS_DEFAULT_REGION
        value: us-east-1
`

func TestSanitizeKubeconfig_RemovesAWSProfile(t *testing.T) {
	out, warnings := SanitizeKubeconfig([]byte(kubeconfigWithProfile))
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], "AWS_PROFILE") {
		t.Fatalf("expected warning about AWS_PROFILE, got: %s", warnings[0])
	}
	if strings.Contains(string(out), "aegis-new") {
		t.Fatalf("sanitized kubeconfig still contains AWS_PROFILE value")
	}
	if !strings.Contains(string(out), "AWS_DEFAULT_REGION") {
		t.Fatalf("sanitized kubeconfig lost safe env var AWS_DEFAULT_REGION")
	}
}

func TestSanitizeKubeconfig_NoChangesWhenClean(t *testing.T) {
	out, warnings := SanitizeKubeconfig([]byte(kubeconfigClean))
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for clean kubeconfig, got: %v", warnings)
	}
	// Should return original bytes unchanged
	if string(out) != kubeconfigClean {
		t.Fatalf("clean kubeconfig was modified")
	}
}

func TestSanitizeKubeconfig_InvalidYAML(t *testing.T) {
	raw := []byte("not: valid: kubeconfig: {{{}}")
	out, warnings := SanitizeKubeconfig(raw)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning for invalid YAML, got %d", len(warnings))
	}
	if !strings.Contains(warnings[0], "failed to parse") {
		t.Fatalf("expected parse error warning, got: %s", warnings[0])
	}
	if string(out) != string(raw) {
		t.Fatalf("invalid YAML should be returned as-is")
	}
}

func TestSanitizeKubeconfig_MultipleEnvVarsRemoved(t *testing.T) {
	kc := `apiVersion: v1
clusters:
- cluster:
    server: https://example.com
  name: c
contexts:
- context:
    cluster: c
    user: u
  name: ctx
current-context: ctx
kind: Config
users:
- name: u
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws
      env:
      - name: AWS_PROFILE
        value: prod
      - name: AWS_CONFIG_FILE
        value: /home/user/.aws/config
      - name: AWS_SHARED_CREDENTIALS_FILE
        value: /home/user/.aws/credentials
      - name: SAFE_VAR
        value: keep-me
`
	out, warnings := SanitizeKubeconfig([]byte(kc))
	if len(warnings) != 3 {
		t.Fatalf("expected 3 warnings, got %d: %v", len(warnings), warnings)
	}
	s := string(out)
	if strings.Contains(s, "AWS_PROFILE") || strings.Contains(s, "AWS_CONFIG_FILE") || strings.Contains(s, "AWS_SHARED_CREDENTIALS_FILE") {
		t.Fatalf("sanitized kubeconfig still contains dangerous env vars")
	}
	if !strings.Contains(s, "SAFE_VAR") {
		t.Fatalf("sanitized kubeconfig lost safe env var")
	}
}

func TestSanitizeKubeconfig_NoExecSection(t *testing.T) {
	kc := `apiVersion: v1
clusters:
- cluster:
    server: https://example.com
  name: c
contexts:
- context:
    cluster: c
    user: u
  name: ctx
current-context: ctx
kind: Config
users:
- name: u
  user:
    token: some-static-token
`
	out, warnings := SanitizeKubeconfig([]byte(kc))
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for token-based kubeconfig, got: %v", warnings)
	}
	if string(out) != kc {
		t.Fatalf("kubeconfig without exec should be returned unchanged")
	}
}
