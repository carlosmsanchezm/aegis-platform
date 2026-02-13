// Package certmanager provides installation of cert-manager and step-issuer
// to spoke clusters for automated TLS certificate management.
package certmanager

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// Environment variables for cert-manager configuration
	envCertManagerEnabled        = "AEGIS_CERT_MANAGER_ENABLED"
	envCertManagerChart          = "AEGIS_CERT_MANAGER_CHART"
	envCertManagerVersion        = "AEGIS_CERT_MANAGER_VERSION"
	envCertManagerRepo           = "AEGIS_CERT_MANAGER_REPO"
	envStepIssuerChart           = "AEGIS_STEP_ISSUER_CHART"
	envStepIssuerVersion         = "AEGIS_STEP_ISSUER_VERSION"
	envStepIssuerRepo            = "AEGIS_STEP_ISSUER_REPO"
	envStepCaURL                 = "AEGIS_STEP_CA_URL"
	envStepCaRootCAB64           = "AEGIS_STEP_CA_ROOT_CA_B64"
	envStepCaRootCAFile          = "AEGIS_STEP_CA_ROOT_CA_FILE"
	envStepCaTLSCAB64            = "AEGIS_STEP_CA_TLS_CA_B64"    // TLS verification CA (e.g., public CA for Cloudflare)
	envStepCaTLSCAFile           = "AEGIS_STEP_CA_TLS_CA_FILE"   // Path to TLS verification CA file
	envStepProvisionerName       = "AEGIS_STEP_PROVISIONER_NAME"
	envStepProvisionerPasswordFile = "AEGIS_STEP_PROVISIONER_PASSWORD_FILE"
	envStepProvisionerKID        = "AEGIS_STEP_PROVISIONER_KID"
	envStepProvisionerPassword   = "AEGIS_STEP_PROVISIONER_PASSWORD"
	envClusterIssuerName         = "AEGIS_CLUSTER_ISSUER_NAME"

	// Defaults
	defaultCertManagerNamespace     = "cert-manager"
	defaultCertManagerChart         = "cert-manager"
	defaultCertManagerRepo          = "https://charts.jetstack.io"
	defaultCertManagerVersion       = "v1.16.2"
	defaultStepIssuerChart          = "step-issuer"
	defaultStepIssuerRepo           = "https://smallstep.github.io/helm-charts"
	defaultStepIssuerVersion        = "1.9.11"
	defaultClusterIssuerName        = "aegis-internal"
	defaultProvisionerName          = "aegis"
	defaultHelmTimeout              = 10 * time.Minute
)

// HelmConfig describes a helm release configuration.
type HelmConfig struct {
	ChartPath   string
	Repository  string
	Version     string
	Namespace   string
	ReleaseName string
	Timeout     time.Duration
}

// StepCAConfig contains connection info for the hub's step-ca.
type StepCAConfig struct {
	URL                 string
	RootCABase64        string // Internal CA used for certificate chain verification
	TLSCABase64         string // TLS connection CA (for Cloudflare/proxy scenarios; falls back to RootCABase64)
	ProvisionerName     string
	ProvisionerKID      string
	ProvisionerPassword string
}

// Config captures the cert-manager configuration for a spoke cluster.
type Config struct {
	CertManager       HelmConfig
	StepIssuer        HelmConfig
	StepCA            StepCAConfig
	ClusterIssuerName string
	Enable            bool
}

// Installer allows installation of cert-manager infrastructure.
type Installer interface {
	// Install installs cert-manager and certificate issuance infrastructure.
	// Returns resources that should be used as dependencies for other components
	// that need to create Certificate resources (e.g., the aegis-spoke helm chart).
	Install(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, cfg Config, depends []pulumi.Resource) ([]pulumi.Resource, error)
}

type defaultInstaller struct{}

// NewInstaller returns the default installer implementation.
func NewInstaller() Installer {
	return defaultInstaller{}
}

// ResolveFromEnv builds a Config using environment variables and defaults.
func ResolveFromEnv() Config {
	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv(envCertManagerEnabled)), "true")

	certManagerChart := strings.TrimSpace(os.Getenv(envCertManagerChart))
	if certManagerChart == "" {
		certManagerChart = defaultCertManagerChart
	}
	certManagerRepo := strings.TrimSpace(os.Getenv(envCertManagerRepo))
	if certManagerRepo == "" {
		certManagerRepo = defaultCertManagerRepo
	}
	certManagerVersion := strings.TrimSpace(os.Getenv(envCertManagerVersion))
	if certManagerVersion == "" {
		certManagerVersion = defaultCertManagerVersion
	}

	stepIssuerChart := strings.TrimSpace(os.Getenv(envStepIssuerChart))
	if stepIssuerChart == "" {
		stepIssuerChart = defaultStepIssuerChart
	}
	stepIssuerRepo := strings.TrimSpace(os.Getenv(envStepIssuerRepo))
	if stepIssuerRepo == "" {
		stepIssuerRepo = defaultStepIssuerRepo
	}
	stepIssuerVersion := strings.TrimSpace(os.Getenv(envStepIssuerVersion))
	if stepIssuerVersion == "" {
		stepIssuerVersion = defaultStepIssuerVersion
	}

	stepCaURL := strings.TrimSpace(os.Getenv(envStepCaURL))
	// Try file-based CA first, then fall back to base64 env var
	var stepCaRootCA string
	if caFile := strings.TrimSpace(os.Getenv(envStepCaRootCAFile)); caFile != "" {
		if data, err := os.ReadFile(caFile); err == nil {
			stepCaRootCA = base64.StdEncoding.EncodeToString(data)
		}
	}
	if stepCaRootCA == "" {
		stepCaRootCA = strings.TrimSpace(os.Getenv(envStepCaRootCAB64))
	}

	// TLS CA for connection verification (e.g., public CA when connecting through Cloudflare)
	// Falls back to RootCA if not set
	var stepCaTLSCA string
	if caFile := strings.TrimSpace(os.Getenv(envStepCaTLSCAFile)); caFile != "" {
		if data, err := os.ReadFile(caFile); err == nil {
			stepCaTLSCA = base64.StdEncoding.EncodeToString(data)
		}
	}
	if stepCaTLSCA == "" {
		stepCaTLSCA = strings.TrimSpace(os.Getenv(envStepCaTLSCAB64))
	}
	// Fall back to internal root CA if TLS CA not specified
	if stepCaTLSCA == "" {
		stepCaTLSCA = stepCaRootCA
	}

	provisionerName := strings.TrimSpace(os.Getenv(envStepProvisionerName))
	if provisionerName == "" {
		provisionerName = defaultProvisionerName
	}
	provisionerKID := strings.TrimSpace(os.Getenv(envStepProvisionerKID))
	// Try file-based password first, then fall back to env var
	var provisionerPassword string
	if pwFile := strings.TrimSpace(os.Getenv(envStepProvisionerPasswordFile)); pwFile != "" {
		if data, err := os.ReadFile(pwFile); err == nil {
			provisionerPassword = strings.TrimSpace(string(data))
		}
	}
	if provisionerPassword == "" {
		provisionerPassword = strings.TrimSpace(os.Getenv(envStepProvisionerPassword))
	}

	clusterIssuerName := strings.TrimSpace(os.Getenv(envClusterIssuerName))
	if clusterIssuerName == "" {
		clusterIssuerName = defaultClusterIssuerName
	}

	return Config{
		Enable: enabled,
		CertManager: HelmConfig{
			ChartPath:   certManagerChart,
			Repository:  certManagerRepo,
			Version:     certManagerVersion,
			Namespace:   defaultCertManagerNamespace,
			ReleaseName: "cert-manager",
			Timeout:     defaultHelmTimeout,
		},
		StepIssuer: HelmConfig{
			ChartPath:   stepIssuerChart,
			Repository:  stepIssuerRepo,
			Version:     stepIssuerVersion,
			Namespace:   defaultCertManagerNamespace,
			ReleaseName: "step-issuer",
			Timeout:     defaultHelmTimeout,
		},
		StepCA: StepCAConfig{
			URL:                 stepCaURL,
			RootCABase64:        stepCaRootCA,
			TLSCABase64:         stepCaTLSCA,
			ProvisionerName:     provisionerName,
			ProvisionerKID:      provisionerKID,
			ProvisionerPassword: provisionerPassword,
		},
		ClusterIssuerName: clusterIssuerName,
	}
}

// Install installs cert-manager and configures certificate issuance on the spoke cluster.
// If step-ca configuration is provided, it installs step-issuer and creates a StepClusterIssuer.
// Otherwise, it creates a self-signed ClusterIssuer for testing/development.
// Returns the resources that should be used as dependencies for components that create Certificate resources.
func (d defaultInstaller) Install(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, cfg Config, depends []pulumi.Resource) ([]pulumi.Resource, error) {
	if !cfg.Enable {
		return nil, nil
	}
	if kubeProvider == nil {
		return nil, fmt.Errorf("kubernetes provider is required for cert-manager installation")
	}

	namespace := cfg.CertManager.Namespace
	if namespace == "" {
		namespace = defaultCertManagerNamespace
	}

	// Install cert-manager
	certMgrRelease, err := installCertManager(ctx, clusterID, kubeProvider, cfg, depends)
	if err != nil {
		return nil, fmt.Errorf("failed to install cert-manager: %w", err)
	}

	// Check if step-ca is configured - if so, use step-issuer; otherwise use self-signed
	useStepCA := cfg.StepCA.URL != "" && cfg.StepCA.RootCABase64 != "" &&
		cfg.StepCA.ProvisionerKID != "" && cfg.StepCA.ProvisionerPassword != ""

	if useStepCA {
		// Install step-issuer (depends on cert-manager)
		stepIssuerRelease, err := installStepIssuer(ctx, clusterID, kubeProvider, cfg, []pulumi.Resource{certMgrRelease})
		if err != nil {
			return nil, fmt.Errorf("failed to install step-issuer: %w", err)
		}

		// Create RBAC for cert-manager to approve StepClusterIssuer certificate requests
		approverRBAC, err := createStepIssuerApproverRBAC(ctx, clusterID, kubeProvider, cfg, namespace, []pulumi.Resource{stepIssuerRelease})
		if err != nil {
			return nil, fmt.Errorf("failed to create step-issuer approver RBAC: %w", err)
		}

		// Create provisioner password secret (depends on step-issuer)
		provSecret, err := createProvisionerSecret(ctx, clusterID, kubeProvider, cfg, namespace, []pulumi.Resource{stepIssuerRelease})
		if err != nil {
			return nil, fmt.Errorf("failed to create provisioner secret: %w", err)
		}

		// Create StepClusterIssuer (depends on secret and RBAC)
		stepClusterIssuer, err := createStepClusterIssuer(ctx, clusterID, kubeProvider, cfg, namespace, []pulumi.Resource{provSecret, approverRBAC})
		if err != nil {
			return nil, fmt.Errorf("failed to create StepClusterIssuer: %w", err)
		}
		// Return the StepClusterIssuer as the final resource that other components should depend on
		return []pulumi.Resource{stepClusterIssuer}, nil
	}

	// Create self-signed ClusterIssuer for testing/development
	caClusterIssuer, err := createSelfSignedClusterIssuer(ctx, clusterID, kubeProvider, cfg, []pulumi.Resource{certMgrRelease})
	if err != nil {
		return nil, fmt.Errorf("failed to create self-signed ClusterIssuer: %w", err)
	}
	// Return the CA ClusterIssuer as the final resource that other components should depend on
	return []pulumi.Resource{caClusterIssuer}, nil
}

func installCertManager(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, cfg Config, depends []pulumi.Resource) (*helm.Release, error) {
	namespace := cfg.CertManager.Namespace
	if namespace == "" {
		namespace = defaultCertManagerNamespace
	}

	timeout := cfg.CertManager.Timeout
	if timeout == 0 {
		timeout = defaultHelmTimeout
	}

	releaseName := pulumiResourceName(cfg.CertManager.ReleaseName+"-"+clusterID, 53)

	values := pulumi.Map{
		"installCRDs": pulumi.Bool(true),
		"resources": pulumi.Map{
			"requests": pulumi.Map{
				"cpu":    pulumi.String("50m"),
				"memory": pulumi.String("64Mi"),
			},
			"limits": pulumi.Map{
				"cpu":    pulumi.String("200m"),
				"memory": pulumi.String("256Mi"),
			},
		},
	}

	args := &helm.ReleaseArgs{
		Name:            pulumi.StringPtr(releaseName),
		Namespace:       pulumi.StringPtr(namespace),
		Chart:           pulumi.String(cfg.CertManager.ChartPath),
		Values:          values,
		Timeout:         pulumi.IntPtr(int(timeout.Seconds())),
		CreateNamespace: pulumi.BoolPtr(true),
		WaitForJobs:     pulumi.BoolPtr(true),
	}

	if cfg.CertManager.Repository != "" {
		opts := helm.RepositoryOptsArgs{Repo: pulumi.StringPtr(cfg.CertManager.Repository)}
		args.RepositoryOpts = opts.ToRepositoryOptsPtrOutput()
	}
	if cfg.CertManager.Version != "" {
		args.Version = pulumi.StringPtr(cfg.CertManager.Version)
	}

	return helm.NewRelease(ctx, pulumiResourceName(clusterID+"-cert-manager", 53), args, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
}

func installStepIssuer(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, cfg Config, depends []pulumi.Resource) (*helm.Release, error) {
	namespace := cfg.StepIssuer.Namespace
	if namespace == "" {
		namespace = defaultCertManagerNamespace
	}

	timeout := cfg.StepIssuer.Timeout
	if timeout == 0 {
		timeout = defaultHelmTimeout
	}

	releaseName := pulumiResourceName(cfg.StepIssuer.ReleaseName+"-"+clusterID, 53)

	values := pulumi.Map{
		"resources": pulumi.Map{
			"requests": pulumi.Map{
				"cpu":    pulumi.String("25m"),
				"memory": pulumi.String("32Mi"),
			},
			"limits": pulumi.Map{
				"cpu":    pulumi.String("100m"),
				"memory": pulumi.String("128Mi"),
			},
		},
	}

	args := &helm.ReleaseArgs{
		Name:            pulumi.StringPtr(releaseName),
		Namespace:       pulumi.StringPtr(namespace),
		Chart:           pulumi.String(cfg.StepIssuer.ChartPath),
		Values:          values,
		Timeout:         pulumi.IntPtr(int(timeout.Seconds())),
		CreateNamespace: pulumi.BoolPtr(true),
		WaitForJobs:     pulumi.BoolPtr(true),
	}

	if cfg.StepIssuer.Repository != "" {
		opts := helm.RepositoryOptsArgs{Repo: pulumi.StringPtr(cfg.StepIssuer.Repository)}
		args.RepositoryOpts = opts.ToRepositoryOptsPtrOutput()
	}
	if cfg.StepIssuer.Version != "" {
		args.Version = pulumi.StringPtr(cfg.StepIssuer.Version)
	}

	return helm.NewRelease(ctx, pulumiResourceName(clusterID+"-step-issuer", 53), args, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
}

func createProvisionerSecret(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, cfg Config, namespace string, depends []pulumi.Resource) (*corev1.Secret, error) {
	secretName := "step-ca-provisioner-password"

	return corev1.NewSecret(ctx, pulumiResourceName(clusterID+"-step-prov-secret", 53), &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(secretName),
			Namespace: pulumi.String(namespace),
			Labels: pulumi.StringMap{
				"app.kubernetes.io/managed-by": pulumi.String("aegis-platform"),
				"app.kubernetes.io/component":  pulumi.String("pki"),
			},
		},
		Type: pulumi.String("Opaque"),
		StringData: pulumi.StringMap{
			"password": pulumi.String(cfg.StepCA.ProvisionerPassword),
		},
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
}

func createStepClusterIssuer(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, cfg Config, namespace string, depends []pulumi.Resource) (*apiextensions.CustomResource, error) {
	issuerName := cfg.ClusterIssuerName
	if issuerName == "" {
		issuerName = defaultClusterIssuerName
	}

	// StepClusterIssuer is a custom resource from step-issuer
	// Use TLSCABase64 for the caBundle - this is the CA that verifies the TLS connection to step-ca
	// (e.g., public CA roots when connecting through Cloudflare tunnel)
	return apiextensions.NewCustomResource(ctx, pulumiResourceName(clusterID+"-step-cluster-issuer", 53), &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("certmanager.step.sm/v1beta1"),
		Kind:       pulumi.String("StepClusterIssuer"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(issuerName),
			Labels: pulumi.StringMap{
				"app.kubernetes.io/managed-by": pulumi.String("aegis-platform"),
				"app.kubernetes.io/component":  pulumi.String("pki"),
			},
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"url":      pulumi.String(cfg.StepCA.URL),
				"caBundle": pulumi.String(cfg.StepCA.TLSCABase64),
				"provisioner": pulumi.Map{
					"name": pulumi.String(cfg.StepCA.ProvisionerName),
					"kid":  pulumi.String(cfg.StepCA.ProvisionerKID),
					"passwordRef": pulumi.Map{
						"name":      pulumi.String("step-ca-provisioner-password"),
						"namespace": pulumi.String(namespace),
						"key":       pulumi.String("password"),
					},
				},
			},
		},
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
}

// createSelfSignedClusterIssuer creates a self-signed ClusterIssuer for testing/development.
// This is used when step-ca is not configured, allowing cert-manager to issue certificates
// without an external CA. Returns the CA ClusterIssuer resource.
func createSelfSignedClusterIssuer(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, cfg Config, depends []pulumi.Resource) (*apiextensions.CustomResource, error) {
	issuerName := cfg.ClusterIssuerName
	if issuerName == "" {
		issuerName = defaultClusterIssuerName
	}

	// First create a self-signed issuer to bootstrap the CA
	_, err := apiextensions.NewCustomResource(ctx, pulumiResourceName(clusterID+"-selfsigned-issuer", 53), &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("ClusterIssuer"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("selfsigned-bootstrap"),
			Labels: pulumi.StringMap{
				"app.kubernetes.io/managed-by": pulumi.String("aegis-platform"),
				"app.kubernetes.io/component":  pulumi.String("pki"),
			},
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"selfSigned": pulumi.Map{},
			},
		},
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
	if err != nil {
		return nil, fmt.Errorf("failed to create self-signed bootstrap issuer: %w", err)
	}

	// Create a CA certificate using the self-signed issuer
	caCert, err := apiextensions.NewCustomResource(ctx, pulumiResourceName(clusterID+"-ca-certificate", 53), &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("Certificate"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("aegis-ca"),
			Namespace: pulumi.String("cert-manager"),
			Labels: pulumi.StringMap{
				"app.kubernetes.io/managed-by": pulumi.String("aegis-platform"),
				"app.kubernetes.io/component":  pulumi.String("pki"),
			},
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"isCA":       pulumi.Bool(true),
				"commonName": pulumi.String("Aegis Spoke CA"),
				"secretName": pulumi.String("aegis-ca-secret"),
				"duration":   pulumi.String("87600h"), // 10 years
				"renewBefore": pulumi.String("720h"),  // 30 days
				"privateKey": pulumi.Map{
					"algorithm": pulumi.String("ECDSA"),
					"size":      pulumi.Int(256),
				},
				"issuerRef": pulumi.Map{
					"name":  pulumi.String("selfsigned-bootstrap"),
					"kind":  pulumi.String("ClusterIssuer"),
					"group": pulumi.String("cert-manager.io"),
				},
			},
		},
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Create the CA ClusterIssuer that will be used to issue certificates
	caClusterIssuer, err := apiextensions.NewCustomResource(ctx, pulumiResourceName(clusterID+"-ca-cluster-issuer", 53), &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("ClusterIssuer"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(issuerName),
			Labels: pulumi.StringMap{
				"app.kubernetes.io/managed-by": pulumi.String("aegis-platform"),
				"app.kubernetes.io/component":  pulumi.String("pki"),
			},
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"ca": pulumi.Map{
					"secretName": pulumi.String("aegis-ca-secret"),
				},
			},
		},
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{caCert}))
	if err != nil {
		return nil, fmt.Errorf("failed to create CA ClusterIssuer: %w", err)
	}

	return caClusterIssuer, nil
}

// createStepIssuerApproverRBAC creates RBAC resources to allow cert-manager to approve
// CertificateRequests for StepClusterIssuer. This is required because cert-manager's
// default approver doesn't have permission to approve requests for custom issuer types.
func createStepIssuerApproverRBAC(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, cfg Config, namespace string, depends []pulumi.Resource) (pulumi.Resource, error) {
	issuerName := cfg.ClusterIssuerName
	if issuerName == "" {
		issuerName = defaultClusterIssuerName
	}

	// Create a ClusterRole that allows approving CertificateRequests for StepClusterIssuers
	clusterRole, err := rbacv1.NewClusterRole(ctx, pulumiResourceName(clusterID+"-step-issuer-approver", 53), &rbacv1.ClusterRoleArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(fmt.Sprintf("cert-manager-approval:certmanager-step-sm:%s", issuerName)),
			Labels: pulumi.StringMap{
				"app.kubernetes.io/managed-by": pulumi.String("aegis-platform"),
				"app.kubernetes.io/component":  pulumi.String("pki"),
			},
		},
		Rules: rbacv1.PolicyRuleArray{
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("cert-manager.io")},
				Resources: pulumi.StringArray{
					pulumi.String("signers"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("approve"),
				},
				// Resource name format: <issuerGroup>/<issuerKind>.<issuerName>
				ResourceNames: pulumi.StringArray{
					pulumi.String(fmt.Sprintf("stepclusterissuers.certmanager.step.sm/%s", issuerName)),
				},
			},
		},
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
	if err != nil {
		return nil, fmt.Errorf("failed to create step-issuer approver ClusterRole: %w", err)
	}

	// Create ClusterRoleBinding to bind the approver role to cert-manager's service account
	clusterRoleBinding, err := rbacv1.NewClusterRoleBinding(ctx, pulumiResourceName(clusterID+"-step-issuer-approver-binding", 53), &rbacv1.ClusterRoleBindingArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(fmt.Sprintf("cert-manager-approval:certmanager-step-sm:%s", issuerName)),
			Labels: pulumi.StringMap{
				"app.kubernetes.io/managed-by": pulumi.String("aegis-platform"),
				"app.kubernetes.io/component":  pulumi.String("pki"),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("ClusterRole"),
			Name:     clusterRole.Metadata.Name().Elem(),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String(fmt.Sprintf("cert-manager-%s", clusterID)),
				Namespace: pulumi.String(namespace),
			},
		},
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{clusterRole}))
	if err != nil {
		return nil, fmt.Errorf("failed to create step-issuer approver ClusterRoleBinding: %w", err)
	}

	return clusterRoleBinding, nil
}

// pulumiResourceName truncates a name to fit within Pulumi's resource name limits.
func pulumiResourceName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen]
}
