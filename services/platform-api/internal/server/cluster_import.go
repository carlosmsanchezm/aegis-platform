package server

import (
	"context"
	"crypto/rand"
	"errors"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	"github.com/yourorg/aegis/services/platform-api/internal/kubeclients"
	"github.com/yourorg/aegis/services/platform-api/internal/spokeinstall"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

const (
	defaultImportKubeconfigSecretName      = "aegis-kubeconfigs"
	defaultImportKubeconfigSecretNamespace = "aegis-system"
	defaultImportCPGRPC                    = "aegis-services-platform-api.aegis-system.svc.cluster.local:8081"
	defaultImportCPGRPCInsecure            = "true"
	defaultImportFlavors                   = "cpu-small,cpu-medium,cpu-large"
	defaultImportInstallCommand            = "helm upgrade --install aegis-spoke ./charts/aegis-spoke -n aegis-system -f values.yaml"
	defaultAgentScriptBaseURL              = "https://aegis.run/install-agent"

	labelClusterName = "aegis.yourorg.dev/clusterName"
	labelProjectID   = "aegis.yourorg.dev/projectId"
	labelImported    = "imported"
)

var allowedImportProviders = map[string]struct{}{
	"local":          {},
	"baremetal":      {},
	"existing-aws":   {},
	"existing-gcp":   {},
	"existing-azure": {},
	"airgapped":      {},
}

var allowedImportMethods = map[string]struct{}{
	"provisioned": {},
	"kubeconfig":  {},
	"assume_role": {},
	"agent_only":  {},
}

func (s *Server) ImportCluster(ctx context.Context, req *aegis.ImportClusterRequest) (*aegis.ImportClusterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}

	projectID := strings.TrimSpace(req.GetProjectId())
	clusterID := strings.TrimSpace(req.GetClusterId())
	provider := strings.ToLower(strings.TrimSpace(req.GetProvider()))
	region := strings.TrimSpace(req.GetRegion())
	name := strings.TrimSpace(req.GetName())
	importMethod := strings.ToLower(strings.TrimSpace(req.GetImportMethod()))

	if projectID == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}
	if err := s.authorize(ctx, projectID, "", "importCluster"); err != nil {
		return nil, err
	}
	if clusterID == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}
	if _, ok := allowedImportProviders[provider]; !ok {
		return nil, status.Errorf(codes.InvalidArgument, "provider %q not supported", req.GetProvider())
	}
	if region == "" {
		return nil, status.Error(codes.InvalidArgument, "region is required")
	}
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if importMethod == "" {
		return nil, status.Error(codes.InvalidArgument, "import_method is required")
	}
	if _, ok := allowedImportMethods[importMethod]; !ok {
		return nil, status.Errorf(codes.InvalidArgument, "import_method %q not supported", req.GetImportMethod())
	}

	warnings := []string{}
	labels := cloneStringMap(req.GetLabels())
	if labels == nil {
		labels = map[string]string{}
	}
	labels[labelClusterName] = name
	labels[labelImported] = "true"

	assumeRoleARN := strings.TrimSpace(req.GetAssumeRoleArn())
	if assumeRoleARN != "" && importMethod != "assume_role" {
		warnings = append(warnings, "assume_role_arn ignored unless import_method=assume_role")
		assumeRoleARN = ""
	}
	if importMethod == "assume_role" {
		if assumeRoleARN == "" {
			return nil, status.Error(codes.InvalidArgument, "assume_role_arn is required when import_method=assume_role")
		}
		if provider != "existing-aws" {
			return nil, status.Error(codes.InvalidArgument, "import_method=assume_role is only supported for provider=existing-aws")
		}
		if err := validateAssumeRoleARN(assumeRoleARN); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	kubeconfigB64 := strings.TrimSpace(req.GetKubeconfig())
	var kubeconfig []byte
	kubeconfigSecretRef := ""
	var rollbackKubeconfig func(context.Context) error
	if kubeconfigB64 != "" && importMethod != "kubeconfig" {
		warnings = append(warnings, "kubeconfig ignored unless import_method=kubeconfig")
		kubeconfigB64 = ""
	}
	if importMethod == "kubeconfig" {
		if kubeconfigB64 == "" {
			return nil, status.Error(codes.InvalidArgument, "kubeconfig is required when import_method=kubeconfig")
		}
		if s.infraClient == nil {
			return nil, status.Error(codes.FailedPrecondition, "kubeconfig uploads are not configured")
		}
		decoded, err := base64.StdEncoding.DecodeString(kubeconfigB64)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "kubeconfig must be base64 encoded")
		}
		kubeconfig = decoded

		// Sanitize kubeconfig: strip environment-specific exec env vars
		sanitized, warnings := kubeclients.SanitizeKubeconfig(kubeconfig)
		for _, w := range warnings {
			s.log.Warn("kubeconfig sanitized during import",
				zap.String("cluster_id", clusterID),
				zap.String("warning", w))
		}
		kubeconfig = sanitized

		if existingProjectID, ok := s.store.GetClusterProjectID(clusterID); ok && existingProjectID != "" && !strings.EqualFold(existingProjectID, projectID) {
			return nil, status.Error(codes.PermissionDenied, "cluster_id is already associated with a different project")
		}

		ref, rollback, err := s.upsertClusterKubeconfigSecret(ctx, clusterID, kubeconfig)
		if err != nil {
			return nil, err
		}
		kubeconfigSecretRef = ref
		rollbackKubeconfig = rollback
	}

	// Extract endpoint+CA from kubeconfig if available (for programmatic token auth fallback)
	var clusterEndpoint, clusterCA string
	if len(kubeconfig) > 0 {
		clusterEndpoint, clusterCA = kubeclients.ExtractClusterEndpointCA(kubeconfig)
	}

	if err := s.store.UpsertClusterImport(store.ClusterImport{
		ClusterID:           clusterID,
		ProjectID:           projectID,
		Provider:            provider,
		Region:              region,
		Labels:              labels,
		ImportMethod:        importMethod,
		ImportedAt:          time.Now(),
		KubeconfigSecretRef: kubeconfigSecretRef,
		AssumeRoleARN:       assumeRoleARN,
		ClusterEndpoint:     clusterEndpoint,
		ClusterCA:           clusterCA,
	}); err != nil {
		if rollbackKubeconfig != nil {
			_ = rollbackKubeconfig(ctx)
		}
		if errors.Is(err, store.ErrClusterProjectConflict) {
			return nil, status.Error(codes.PermissionDenied, "cluster_id is already associated with a different project")
		}
		return nil, status.Errorf(codes.Internal, "import cluster: %v", err)
	}

	statusValue := "pending_agent"
	if info := s.store.GetClusterInfo(clusterID); info != nil && !info.LastHeartbeat.IsZero() {
		statusValue = "active"
	}

	// Automatically install the spoke agent on the remote cluster if credentials were provided.
	// This runs in a background goroutine — the import response returns immediately with
	// status "installing". The UI polls ListClusters to detect when the agent heartbeats.
	if (importMethod == "kubeconfig" || importMethod == "assume_role") && s.kubeClients != nil {
		statusValue = "installing"
		go s.installSpokeOnImportedCluster(clusterID, projectID, provider, region)
	}

	helmEnv := map[string]string{
		"AEGIS_CLUSTER_ID":       clusterID,
		"AEGIS_PROVIDER":         provider,
		"AEGIS_REGION":           region,
		"AEGIS_CP_GRPC":          getenv("AEGIS_IMPORT_CP_GRPC", defaultImportCPGRPC),
		"AEGIS_CP_GRPC_INSECURE": getenv("AEGIS_IMPORT_CP_GRPC_INSECURE", defaultImportCPGRPCInsecure),
		"AEGIS_FLAVORS":          getenv("AEGIS_IMPORT_DEFAULT_FLAVORS", defaultImportFlavors),
	}

	installCommand := getenv("AEGIS_IMPORT_INSTALL_COMMAND", defaultImportInstallCommand)
	agentScriptURL := s.buildAgentScriptURL(clusterID)

	return &aegis.ImportClusterResponse{
		ClusterId: clusterID,
		Status:    statusValue,
		HelmValues: &aegis.HelmValues{
			K8SAgent: &aegis.K8SAgentHelmValues{Env: helmEnv},
		},
		InstallCommand: installCommand,
		AgentScriptUrl: agentScriptURL,
		Warnings:       warnings,
	}, nil
}

func (s *Server) upsertClusterKubeconfigSecret(ctx context.Context, clusterID string, kubeconfig []byte) (string, func(context.Context) error, error) {
	if s == nil || s.infraClient == nil {
		return "", nil, status.Error(codes.FailedPrecondition, "kubeconfig uploads are not configured")
	}
	if strings.TrimSpace(clusterID) == "" {
		return "", nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}
	if len(kubeconfig) == 0 {
		return "", nil, status.Error(codes.InvalidArgument, "kubeconfig payload required")
	}

	secretName := getenv("AEGIS_KUBECONFIG_SECRET_NAME", defaultImportKubeconfigSecretName)
	secretNamespace := getenv("AEGIS_KUBECONFIG_SECRET_NAMESPACE", defaultImportKubeconfigSecretNamespace)
	secretKey := fmt.Sprintf("%s.kubeconfig", clusterID)
	ref := fmt.Sprintf("%s/%s:%s", secretNamespace, secretName, secretKey)

	var before []byte
	var hadBefore bool
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		secret := &corev1.Secret{}
		key := types.NamespacedName{Name: secretName, Namespace: secretNamespace}
		getErr := s.infraClient.Get(ctx, key, secret)
		if apierrors.IsNotFound(getErr) {
			before = nil
			hadBefore = false
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: secretNamespace},
				Type:       corev1.SecretTypeOpaque,
				Data:       map[string][]byte{secretKey: kubeconfig},
			}
			if err := s.infraClient.Create(ctx, secret); err != nil {
				return err
			}
			return nil
		}
		if getErr != nil {
			return getErr
		}
		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}
		if v, ok := secret.Data[secretKey]; ok {
			hadBefore = true
			before = append([]byte(nil), v...)
		} else {
			hadBefore = false
			before = nil
		}
		secret.Data[secretKey] = kubeconfig
		return s.infraClient.Update(ctx, secret)
	})
	if err != nil {
		return "", nil, status.Errorf(codes.Internal, "write kubeconfig secret %s: %v", ref, err)
	}

	rollback := func(rollbackCtx context.Context) error {
		if rollbackCtx == nil {
			rollbackCtx = context.Background()
		}
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			secret := &corev1.Secret{}
			key := types.NamespacedName{Name: secretName, Namespace: secretNamespace}
			getErr := s.infraClient.Get(rollbackCtx, key, secret)
			if apierrors.IsNotFound(getErr) {
				return nil
			}
			if getErr != nil {
				return getErr
			}
			if secret.Data == nil {
				secret.Data = map[string][]byte{}
			}
			if hadBefore {
				secret.Data[secretKey] = before
			} else {
				delete(secret.Data, secretKey)
			}
			return s.infraClient.Update(rollbackCtx, secret)
		})
	}

	return ref, rollback, nil
}

func (s *Server) buildAgentScriptURL(clusterID string) string {
	base := strings.TrimSpace(os.Getenv("AEGIS_AGENT_SCRIPT_BASE_URL"))
	if base == "" {
		base = defaultAgentScriptBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return ""
	}
	q := parsed.Query()
	q.Set("cluster", clusterID)
	if tok := newOpaqueToken(16); tok != "" {
		q.Set("token", tok)
	}
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

func newOpaqueToken(nbytes int) string {
	if nbytes <= 0 {
		return ""
	}
	buf := make([]byte, nbytes)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

func validateAssumeRoleARN(arn string) error {
	trimmed := strings.TrimSpace(arn)
	if trimmed == "" {
		return fmt.Errorf("assume_role_arn is required")
	}
	// arn:partition:service:region:account-id:resource
	parts := strings.SplitN(trimmed, ":", 6)
	if len(parts) != 6 || parts[0] != "arn" {
		return fmt.Errorf("assume_role_arn must be a valid ARN")
	}
	if parts[2] != "iam" {
		return fmt.Errorf("assume_role_arn must be an IAM role ARN")
	}
	if acct := parts[4]; len(acct) != 12 {
		return fmt.Errorf("assume_role_arn must include a 12-digit account id")
	}
	resource := parts[5]
	if !strings.HasPrefix(resource, "role/") || strings.TrimPrefix(resource, "role/") == "" {
		return fmt.Errorf("assume_role_arn must reference an IAM role resource")
	}
	return nil
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// inferImportOIDCTokenURL derives the Keycloak token URL from the hub's own OIDC configuration.
func inferImportOIDCTokenURL() string {
	if v := strings.TrimSpace(os.Getenv("AEGIS_IMPORT_CP_OIDC_TOKEN_URL")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("OIDC_ISSUER_URL")); v != "" {
		return v + "/protocol/openid-connect/token"
	}
	return ""
}

// inferImportOIDCClientSecret derives the spoke OIDC client secret from environment.
func inferImportOIDCClientSecret() string {
	if v := strings.TrimSpace(os.Getenv("AEGIS_IMPORT_CP_OIDC_CLIENT_SECRET")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("AEGIS_SPOKE_OIDC_CLIENT_SECRET"))
}

// inferImportOIDCAudience derives the OIDC audience from environment.
func inferImportOIDCAudience() string {
	if v := strings.TrimSpace(os.Getenv("AEGIS_IMPORT_CP_OIDC_AUDIENCE")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("AEGIS_CP_OIDC_AUDIENCE"))
}

// installSpokeOnImportedCluster runs in a background goroutine after a cluster import
// with kubeconfig or assume_role credentials. It connects to the remote cluster and
// installs the aegis-spoke Helm chart using the spokeinstall package.
func (s *Server) installSpokeOnImportedCluster(clusterID, projectID, provider, region string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	restCfg, err := s.kubeClients.RestConfigFor(clusterID)
	if err != nil {
		s.log.Error("import: failed to get kubeconfig for remote cluster",
			zap.String("cluster_id", clusterID),
			zap.Error(err))
		return
	}

	chartPath := getenv("AEGIS_SPOKE_CHART_PATH", "charts/aegis-spoke")

	cfg := spokeinstall.SpokeInstallConfig{
		ClusterID:        clusterID,
		Provider:         provider,
		Region:           region,
		HubGRPC:          getenv("AEGIS_IMPORT_CP_GRPC", defaultImportCPGRPC),
		HubGRPCInsecure:  getenv("AEGIS_IMPORT_CP_GRPC_INSECURE", defaultImportCPGRPCInsecure) == "true",
		OIDCTokenURL:     inferImportOIDCTokenURL(),
		OIDCClientID:     getenv("AEGIS_IMPORT_CP_OIDC_CLIENT_ID", "spoke-agent"),
		OIDCClientSecret: inferImportOIDCClientSecret(),
		OIDCAudience:     inferImportOIDCAudience(),
		CABundleB64:      strings.TrimSpace(os.Getenv("AEGIS_PLATFORM_CA_B64")),
		ProxyJWTSecret:   strings.TrimSpace(os.Getenv("AEGIS_PROXY_JWT_SECRET")),
		AgentImageRepo:   strings.TrimSpace(os.Getenv("AEGIS_IMPORT_AGENT_IMAGE_REPO")),
		AgentImageTag:    strings.TrimSpace(os.Getenv("AEGIS_IMPORT_AGENT_IMAGE_TAG")),
		ProxyImageRepo:   strings.TrimSpace(os.Getenv("AEGIS_IMPORT_PROXY_IMAGE_REPO")),
		ProxyImageTag:    strings.TrimSpace(os.Getenv("AEGIS_IMPORT_PROXY_IMAGE_TAG")),
		ProxyEnabled:     getenv("AEGIS_IMPORT_PROXY_ENABLED", "true") == "true",
		Flavors:          getenv("AEGIS_IMPORT_DEFAULT_FLAVORS", defaultImportFlavors),
	}

	// Request TLS cert from hub's step-ca for the spoke proxy
	if s.infraClient != nil {
		certPEM, keyPEM, certErr := spokeinstall.RequestSpokeProxyCert(ctx, s.infraClient, clusterID, "aegis-system")
		if certErr != nil {
			s.log.Warn("import: failed to get spoke proxy cert from step-ca, spoke proxy will lack TLS",
				zap.String("cluster_id", clusterID),
				zap.Error(certErr))
		} else {
			cfg.ProxyTLSCert = certPEM
			cfg.ProxyTLSKey = keyPEM
			s.log.Info("import: spoke proxy cert issued by hub step-ca",
				zap.String("cluster_id", clusterID))
		}

		// Read hub CA for spoke agent TLS verification
		hubCA, caErr := spokeinstall.ReadHubCA(ctx, s.infraClient, "aegis-system")
		if caErr != nil {
			s.log.Warn("import: failed to read hub CA from trust bundle",
				zap.String("cluster_id", clusterID),
				zap.Error(caErr))
		} else {
			cfg.HubCABundle = hubCA
		}
	}

	s.log.Info("import: starting spoke installation on remote cluster",
		zap.String("cluster_id", clusterID),
		zap.String("provider", provider),
		zap.String("region", region),
		zap.Bool("has_tls_cert", cfg.ProxyTLSCert != ""),
		zap.Bool("has_hub_ca", cfg.HubCABundle != ""))

	if err := spokeinstall.Install(ctx, s.log, restCfg, chartPath, cfg); err != nil {
		s.log.Error("import: spoke installation failed",
			zap.String("cluster_id", clusterID),
			zap.Error(err))
		return
	}

	s.log.Info("import: spoke installation completed, waiting for agent heartbeat",
		zap.String("cluster_id", clusterID))
}
