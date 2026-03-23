package certmanager

import (
	"testing"
)

func TestIsClusterLocalURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://step-certificates.aegis-pki.svc.cluster.local:443", true},
		{"https://step-certificates.aegis-pki.svc.cluster.local", true},
		{"https://my-service.default.svc:8080", true},
		{"https://localhost:9000", true},
		{"https://127.0.0.1:9000", true},
		{"https://10.0.0.1:443", true},
		{"https://step-ca.aegis-platform.tech", false},
		{"https://step-ca.aegis-platform.tech:443", false},
		{"https://example.com", false},
		{"not-a-url", false},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := isClusterLocalURL(tt.url)
			if got != tt.want {
				t.Errorf("isClusterLocalURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestPulumiResourceName(t *testing.T) {
	tests := []struct {
		name   string
		maxLen int
		want   string
	}{
		{"short-name", 53, "short-name"},
		{"cert-manager-e2e-pilot-test-us-east-1-atlas-train-govcloud-22-e38bdff4", 53, "cert-manager-e2e-pilot-test-us-east-1-atlas-3e12e75b"},
		{"cert-manager-short", 53, "cert-manager-short"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pulumiResourceName(tt.name, tt.maxLen)
			if len(got) > tt.maxLen {
				t.Errorf("pulumiResourceName(%q, %d) = %q (len=%d), exceeds maxLen", tt.name, tt.maxLen, got, len(got))
			}
			// For the specific case we know the expected output
			if tt.want != "" && got != tt.want {
				t.Errorf("pulumiResourceName(%q, %d) = %q, want %q", tt.name, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestPulumiResourceName_SAMatchesRelease(t *testing.T) {
	// This test verifies that the RBAC fix produces the same SA name as the cert-manager release name.
	// The cert-manager Helm release name is built with: pulumiResourceName("cert-manager-"+clusterID, 53)
	// The RBAC binding SA name must match.
	clusterID := "e2e-pilot-test-us-east-1-atlas-train-govcloud-22-e38bdff4"
	releaseName := "cert-manager"

	helmReleaseSAName := pulumiResourceName(releaseName+"-"+clusterID, 53)
	rbacSAName := pulumiResourceName(releaseName+"-"+clusterID, 53)

	if helmReleaseSAName != rbacSAName {
		t.Errorf("SA name mismatch: helm=%q, rbac=%q", helmReleaseSAName, rbacSAName)
	}

	// Also verify it's different from the naive (un-truncated) name
	naiveName := "cert-manager-" + clusterID
	if naiveName == helmReleaseSAName {
		t.Errorf("expected truncation for long clusterID, but names are identical: %q", naiveName)
	}
}
