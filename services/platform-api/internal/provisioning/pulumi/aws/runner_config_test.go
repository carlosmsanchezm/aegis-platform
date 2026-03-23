package aws

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

// ----------------------------------------------------------------------------- //
// resolvePlatformConfig tests

func TestResolvePlatformConfig_AllSet(t *testing.T) {
	t.Setenv("AEGIS_PLATFORM_API_ENDPOINT", "platform-api.example.com:8081")
	t.Setenv("AEGIS_PLATFORM_API_CA_B64", "dGVzdC1jYQ==")
	t.Setenv("AEGIS_SPOKE_IMAGE_REPO", "123456789.dkr.ecr.us-east-1.amazonaws.com/aegis/k8s-agent")
	t.Setenv("AEGIS_SPOKE_IMAGE_TAG", "abc1234")
	t.Setenv("AEGIS_PLATFORM_API_GRPC_INSECURE", "false")

	r := &Runner{log: zap.NewNop()}
	cfg := r.resolvePlatformConfig()

	if cfg.Endpoint != "platform-api.example.com:8081" {
		t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, "platform-api.example.com:8081")
	}
	if cfg.CABundleBase64 != "dGVzdC1jYQ==" {
		t.Errorf("CABundleBase64 = %q, want %q", cfg.CABundleBase64, "dGVzdC1jYQ==")
	}
	if cfg.ImageRepo != "123456789.dkr.ecr.us-east-1.amazonaws.com/aegis/k8s-agent" {
		t.Errorf("ImageRepo = %q", cfg.ImageRepo)
	}
	if cfg.ImageTag != "abc1234" {
		t.Errorf("ImageTag = %q, want %q", cfg.ImageTag, "abc1234")
	}
	if cfg.InsecureGRPC {
		t.Error("InsecureGRPC should be false")
	}
}

func TestResolvePlatformConfig_Empty(t *testing.T) {
	// Ensure env vars are not set
	for _, key := range []string{
		"AEGIS_PLATFORM_API_ENDPOINT", "AEGIS_PLATFORM_API_CA_B64",
		"AEGIS_SPOKE_IMAGE_REPO", "AEGIS_SPOKE_IMAGE_TAG",
		"AEGIS_PLATFORM_API_GRPC_INSECURE", "AEGIS_PLATFORM_API_CA_FILE",
	} {
		t.Setenv(key, "")
	}

	r := &Runner{log: zap.NewNop()}
	cfg := r.resolvePlatformConfig()

	if cfg.Endpoint != "" {
		t.Errorf("Endpoint = %q, want empty", cfg.Endpoint)
	}
	if cfg.ImageRepo != "" {
		t.Errorf("ImageRepo = %q, want empty", cfg.ImageRepo)
	}
}

func TestResolvePlatformConfig_CAFromFile(t *testing.T) {
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.crt")
	if err := os.WriteFile(caFile, []byte("test-ca-content"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("AEGIS_PLATFORM_API_CA_B64", "")
	t.Setenv("AEGIS_PLATFORM_API_CA_FILE", caFile)

	r := &Runner{log: zap.NewNop()}
	cfg := r.resolvePlatformConfig()

	decoded, err := base64.StdEncoding.DecodeString(cfg.CABundleBase64)
	if err != nil {
		t.Fatalf("CABundleBase64 not valid base64: %v", err)
	}
	if string(decoded) != "test-ca-content" {
		t.Errorf("decoded CA = %q, want %q", string(decoded), "test-ca-content")
	}
}

func TestResolvePlatformConfig_InsecureGRPC(t *testing.T) {
	tests := []struct {
		envVal string
		want   bool
	}{
		{"true", true},
		{"True", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"notabool", false}, // invalid → default false
		{"", false},
	}
	for _, tt := range tests {
		t.Run("insecure="+tt.envVal, func(t *testing.T) {
			t.Setenv("AEGIS_PLATFORM_API_GRPC_INSECURE", tt.envVal)
			r := &Runner{log: zap.NewNop()}
			cfg := r.resolvePlatformConfig()
			if cfg.InsecureGRPC != tt.want {
				t.Errorf("InsecureGRPC = %v, want %v (env=%q)", cfg.InsecureGRPC, tt.want, tt.envVal)
			}
		})
	}
}

// ----------------------------------------------------------------------------- //
// resolveHelmConfig tests

func TestResolveHelmConfig_Defaults(t *testing.T) {
	for _, key := range []string{
		"AEGIS_SPOKE_CHART", "AEGIS_SPOKE_CHART_VERSION",
		"AEGIS_SPOKE_REPO", "AEGIS_SPOKE_VALUES_FILE",
		"AEGIS_SPOKE_HELM_TIMEOUT_SECONDS",
	} {
		t.Setenv(key, "")
	}

	r := &Runner{log: zap.NewNop()}
	cfg := r.resolveHelmConfig()

	if cfg.Namespace != "aegis-system" {
		t.Errorf("Namespace = %q, want %q", cfg.Namespace, "aegis-system")
	}
	if cfg.Timeout != 15*time.Minute {
		t.Errorf("Timeout = %v, want 15m", cfg.Timeout)
	}
	// ChartPath should be non-empty (falls back to repo root detection)
	if cfg.ChartPath == "" {
		t.Error("ChartPath should not be empty (should fallback to repo root)")
	}
}

func TestResolveHelmConfig_ExplicitValuesFile(t *testing.T) {
	t.Setenv("AEGIS_SPOKE_VALUES_FILE", "/root/charts/aegis-spoke/values-cloud-remote.yaml")

	r := &Runner{log: zap.NewNop()}
	cfg := r.resolveHelmConfig()

	if cfg.ValuesFile != "/root/charts/aegis-spoke/values-cloud-remote.yaml" {
		t.Errorf("ValuesFile = %q", cfg.ValuesFile)
	}
}

func TestResolveHelmConfig_CustomTimeout(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    time.Duration
	}{
		{"valid", "300", 300 * time.Second},
		{"invalid_string", "notanumber", 15 * time.Minute},
		{"zero", "0", 15 * time.Minute},
		{"negative", "-5", 15 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AEGIS_SPOKE_HELM_TIMEOUT_SECONDS", tt.envVal)
			r := &Runner{log: zap.NewNop()}
			cfg := r.resolveHelmConfig()
			if cfg.Timeout != tt.want {
				t.Errorf("Timeout = %v, want %v (env=%q)", cfg.Timeout, tt.want, tt.envVal)
			}
		})
	}
}

// ----------------------------------------------------------------------------- //
// validateProgramInput tests

func newTestRunner() *Runner {
	log, _ := zap.NewDevelopment()
	return &Runner{log: log}
}

func validCloudInput(t *testing.T) *programInput {
	t.Helper()
	// Create a temp dir to act as values file and chart path
	dir := t.TempDir()
	valuesFile := filepath.Join(dir, "values-cloud.yaml")
	if err := os.WriteFile(valuesFile, []byte("k8sAgent: {}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set OIDC env vars for a complete config
	t.Setenv("AEGIS_SPOKE_OIDC_TOKEN_URL", "https://keycloak.example.com/realms/aegis/protocol/openid-connect/token")
	t.Setenv("AEGIS_SPOKE_OIDC_CLIENT_ID", "spoke-agent")
	t.Setenv("AEGIS_SPOKE_OIDC_CLIENT_SECRET", "test-secret")
	t.Setenv("AEGIS_SPOKE_OIDC_AUDIENCE", "aegis-platform")

	return &programInput{
		Platform: platformConfig{
			Endpoint:  "platform-api.example.com:8081",
			ImageRepo: "123456789.dkr.ecr.us-east-1.amazonaws.com/aegis/k8s-agent",
			ImageTag:  "abc1234",
		},
		SpokeHelm: helmConfig{
			ChartPath:  dir,
			ValuesFile: valuesFile,
		},
	}
}

func TestValidate_FullCloudConfig(t *testing.T) {
	r := newTestRunner()
	input := validCloudInput(t)

	if err := r.validateProgramInput(input); err != nil {
		t.Errorf("expected nil error for valid config, got: %v", err)
	}
}

func TestValidate_SkipHelm(t *testing.T) {
	r := newTestRunner()
	input := &programInput{SkipHelm: true}

	if err := r.validateProgramInput(input); err != nil {
		t.Errorf("expected nil error when SkipHelm=true, got: %v", err)
	}
}

func TestValidate_MissingEndpoint(t *testing.T) {
	r := newTestRunner()
	input := validCloudInput(t)
	input.Platform.Endpoint = ""

	err := r.validateProgramInput(input)
	if err == nil {
		t.Fatal("expected error for missing endpoint")
	}
	if !strings.Contains(err.Error(), "AEGIS_PLATFORM_API_ENDPOINT") {
		t.Errorf("error should mention AEGIS_PLATFORM_API_ENDPOINT, got: %v", err)
	}
}

func TestValidate_InvalidEndpoint(t *testing.T) {
	r := newTestRunner()
	input := validCloudInput(t)
	// grpcEndpointHostPort strips schemes and paths; only whitespace-only produces empty
	input.Platform.Endpoint = "   "

	err := r.validateProgramInput(input)
	if err == nil {
		t.Fatal("expected error for whitespace-only endpoint")
	}
	if !strings.Contains(err.Error(), "AEGIS_PLATFORM_API_ENDPOINT") {
		t.Errorf("error should mention AEGIS_PLATFORM_API_ENDPOINT, got: %v", err)
	}
}

func TestValidate_MissingImageRepo(t *testing.T) {
	r := newTestRunner()
	input := validCloudInput(t)
	input.Platform.ImageRepo = ""

	err := r.validateProgramInput(input)
	if err == nil {
		t.Fatal("expected error for missing image repo")
	}
	if !strings.Contains(err.Error(), "AEGIS_SPOKE_IMAGE_REPO") {
		t.Errorf("error should mention AEGIS_SPOKE_IMAGE_REPO, got: %v", err)
	}
}

func TestValidate_InvalidImageRepo(t *testing.T) {
	r := newTestRunner()
	input := validCloudInput(t)
	input.Platform.ImageRepo = "noslash"

	err := r.validateProgramInput(input)
	if err == nil {
		t.Fatal("expected error for invalid image repo")
	}
	if !strings.Contains(err.Error(), "missing '/'") {
		t.Errorf("error should mention missing '/', got: %v", err)
	}
}

func TestValidate_InvalidValuesFile(t *testing.T) {
	r := newTestRunner()
	input := validCloudInput(t)
	input.SpokeHelm.ValuesFile = "/nonexistent/path/values.yaml"

	err := r.validateProgramInput(input)
	if err == nil {
		t.Fatal("expected error for nonexistent values file")
	}
	if !strings.Contains(err.Error(), "AEGIS_SPOKE_VALUES_FILE") {
		t.Errorf("error should mention AEGIS_SPOKE_VALUES_FILE, got: %v", err)
	}
}

func TestValidate_InvalidChartPath(t *testing.T) {
	r := newTestRunner()
	input := validCloudInput(t)
	input.SpokeHelm.ChartPath = "/nonexistent/chart/path"

	err := r.validateProgramInput(input)
	if err == nil {
		t.Fatal("expected error for nonexistent chart path")
	}
	if !strings.Contains(err.Error(), "spoke chart path") {
		t.Errorf("error should mention 'spoke chart path', got: %v", err)
	}
}

func TestValidate_PartialOIDC(t *testing.T) {
	r := newTestRunner()
	input := validCloudInput(t)

	// Set only token URL, clear the rest
	t.Setenv("AEGIS_SPOKE_OIDC_TOKEN_URL", "https://keycloak.example.com/token")
	t.Setenv("AEGIS_SPOKE_OIDC_CLIENT_ID", "")
	t.Setenv("AEGIS_SPOKE_OIDC_CLIENT_SECRET", "")
	t.Setenv("AEGIS_SPOKE_OIDC_AUDIENCE", "")

	err := r.validateProgramInput(input)
	if err == nil {
		t.Fatal("expected error for partial OIDC config")
	}
	if !strings.Contains(err.Error(), "partial OIDC") {
		t.Errorf("error should mention 'partial OIDC', got: %v", err)
	}
}

func TestValidate_AllErrors(t *testing.T) {
	r := newTestRunner()

	// Clear all OIDC but set one to trigger partial error
	t.Setenv("AEGIS_SPOKE_OIDC_TOKEN_URL", "https://keycloak.example.com/token")
	t.Setenv("AEGIS_SPOKE_OIDC_CLIENT_ID", "")
	t.Setenv("AEGIS_SPOKE_OIDC_CLIENT_SECRET", "")
	t.Setenv("AEGIS_SPOKE_OIDC_AUDIENCE", "")

	input := &programInput{
		Platform: platformConfig{
			Endpoint:  "",
			ImageRepo: "",
		},
		SpokeHelm: helmConfig{
			ValuesFile: "/nonexistent/values.yaml",
			ChartPath:  "/nonexistent/chart",
		},
	}

	err := r.validateProgramInput(input)
	if err == nil {
		t.Fatal("expected aggregated errors")
	}

	errMsg := err.Error()
	checks := []string{
		"AEGIS_PLATFORM_API_ENDPOINT",
		"AEGIS_SPOKE_IMAGE_REPO",
		"AEGIS_SPOKE_VALUES_FILE",
		"spoke chart path",
		"partial OIDC",
	}
	for _, check := range checks {
		if !strings.Contains(errMsg, check) {
			t.Errorf("aggregated error should contain %q, got: %s", check, errMsg)
		}
	}
}

// ----------------------------------------------------------------------------- //
// grpcEndpointHostPort tests

func TestGrpcEndpointHostPort(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"host_port", "platform-api.example.com:8081", "platform-api.example.com:8081"},
		{"grpcs_scheme", "grpcs://platform-api.example.com:8081", "platform-api.example.com:8081"},
		{"grpc_scheme", "grpc://host:8081", "host:8081"},
		{"https_scheme", "https://host:443", "host:443"},
		{"http_scheme", "http://host:8080", "host:8080"},
		{"with_path", "host:8081/extra/path", "host:8081"},
		{"scheme_and_path", "grpcs://host:8081/path", "host:8081"},
		{"empty", "", ""},
		{"whitespace", "  ", ""},
		{"trailing_slash", "host:8081/", "host:8081"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grpcEndpointHostPort(tt.input)
			if got != tt.want {
				t.Errorf("grpcEndpointHostPort(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ----------------------------------------------------------------------------- //
// spokeProxyImageRepo tests

func TestSpokeProxyImageRepo(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"ecr", "195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/k8s-agent", "195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/proxy"},
		{"dockerhub", "carlosmsanchez/aegis-k8s-agent", "carlosmsanchez/aegis-proxy"},
		{"generic_suffix", "registry.example.com/org/k8s-agent", "registry.example.com/org/proxy"},
		{"generic_fallback", "my-registry/custom-k8s-agent-image", "my-registry/custom-proxy-image"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := spokeProxyImageRepo(tt.input)
			if got != tt.want {
				t.Errorf("spokeProxyImageRepo(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ----------------------------------------------------------------------------- //
// buildClusterID tests

func TestBuildClusterID(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		region    string
		cluster   string
		want      string
	}{
		{"project_and_name", "e2e-pilot-test", "us-east-1", "training-1", "e2e-pilot-test-training-1"},
		{"no_region_in_id", "demo", "us-west-2", "cluster-1", "demo-cluster-1"},
		{"project_only", "myproject", "", "gpu-cluster", "myproject-gpu-cluster"},
		{"empty_project", "", "us-east-1", "standalone", "standalone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildClusterID(tt.projectID, tt.region, tt.cluster)
			if got != tt.want {
				t.Errorf("buildClusterID(%q, %q, %q) = %q, want %q", tt.projectID, tt.region, tt.cluster, got, tt.want)
			}
		})
	}
}

func TestBuildClusterIDWithSuffix(t *testing.T) {
	// buildClusterIDWithSuffix should now produce the same result as buildClusterID
	// (suffix is ignored for new clusters)
	got := buildClusterIDWithSuffix("e2e-pilot-test", "us-east-1", "training-1", "ae4924f8")
	want := "e2e-pilot-test-training-1"
	if got != want {
		t.Errorf("buildClusterIDWithSuffix = %q, want %q (suffix should be ignored)", got, want)
	}
}
