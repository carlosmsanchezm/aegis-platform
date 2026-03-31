// Package spokeinstall — pki.go installs cert-manager + step-issuer on remote
// clusters during the import flow. This mirrors what the Pulumi-based
// certmanager.Installer does, but uses helm.sh/helm/v3 directly.
package spokeinstall

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmcli "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PKIConfig holds all parameters needed to install cert-manager + step-issuer
// on a remote spoke cluster.
type PKIConfig struct {
	StepCAURL            string // Hub step-ca URL (e.g., NLB DNS or internal svc)
	StepCARootCAB64      string // Root CA PEM (base64) — the step-ca root cert
	StepCATLSCAB64       string // TLS CA for verifying step-ca connection (may differ from root if behind proxy)
	ProvisionerName      string // e.g., "aegis"
	ProvisionerKID       string // JWK Key ID from step-ca config
	ProvisionerPassword  string // Provisioner password for JWK auth
	ClusterIssuerName    string // e.g., "aegis-internal"
	CertManagerNamespace string // default: "cert-manager"
}

// InstallPKI installs cert-manager + step-issuer on a remote cluster and creates
// a StepClusterIssuer pointing to the hub's step-ca.
func InstallPKI(ctx context.Context, log *zap.Logger, restConfig *rest.Config, cfg PKIConfig) error {
	if cfg.CertManagerNamespace == "" {
		cfg.CertManagerNamespace = "cert-manager"
	}
	if cfg.ClusterIssuerName == "" {
		cfg.ClusterIssuerName = "aegis-internal"
	}

	log.Info("pki: starting cert-manager + step-issuer installation",
		zap.String("step_ca_url", cfg.StepCAURL),
		zap.String("provisioner", cfg.ProvisionerName),
		zap.String("issuer_name", cfg.ClusterIssuerName))

	// Step 1: Install cert-manager
	if err := installCertManager(ctx, log, restConfig, cfg.CertManagerNamespace); err != nil {
		return fmt.Errorf("install cert-manager: %w", err)
	}

	// Step 2: Install step-issuer
	if err := installStepIssuer(ctx, log, restConfig, cfg.CertManagerNamespace); err != nil {
		return fmt.Errorf("install step-issuer: %w", err)
	}

	// Step 3: Create RBAC for step-issuer approver
	if err := createStepIssuerRBAC(ctx, restConfig, cfg.CertManagerNamespace); err != nil {
		return fmt.Errorf("create step-issuer RBAC: %w", err)
	}

	// Step 4: Create provisioner password secret
	if err := createProvisionerPasswordSecret(ctx, restConfig, cfg.CertManagerNamespace, cfg.ProvisionerPassword); err != nil {
		return fmt.Errorf("create provisioner password secret: %w", err)
	}

	// Step 5: Create StepClusterIssuer
	if err := createStepClusterIssuer(ctx, log, restConfig, cfg); err != nil {
		return fmt.Errorf("create StepClusterIssuer: %w", err)
	}

	log.Info("pki: cert-manager + step-issuer installation complete",
		zap.String("issuer_name", cfg.ClusterIssuerName))

	return nil
}

func installCertManager(ctx context.Context, log *zap.Logger, restConfig *rest.Config, namespace string) error {
	log.Info("pki: installing cert-manager")

	settings := helmcli.New()
	repoEntry := &repo.Entry{
		Name: "jetstack",
		URL:  "https://charts.jetstack.io",
	}

	// Download chart
	chartPath, err := downloadChart(settings, repoEntry, "cert-manager", "")
	if err != nil {
		return fmt.Errorf("download cert-manager chart: %w", err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("load cert-manager chart: %w", err)
	}

	rcg := &restClientGetter{restConfig: restConfig, namespace: namespace}
	actionCfg := new(action.Configuration)
	if err := actionCfg.Init(rcg, namespace, "secret", func(format string, v ...interface{}) {
		log.Debug(fmt.Sprintf(format, v...), zap.String("component", "helm-certmgr"))
	}); err != nil {
		return fmt.Errorf("init helm config: %w", err)
	}

	// Check if already installed
	histClient := action.NewHistory(actionCfg)
	histClient.Max = 1
	if _, err := histClient.Run("cert-manager"); err == nil {
		log.Info("pki: cert-manager already installed, skipping")
		return nil
	}

	installClient := action.NewInstall(actionCfg)
	installClient.ReleaseName = "cert-manager"
	installClient.Namespace = namespace
	installClient.CreateNamespace = true
	installClient.Timeout = 5 * time.Minute
	installClient.Wait = true

	values := map[string]interface{}{
		"installCRDs": true,
	}

	if _, err := installClient.RunWithContext(ctx, chart, values); err != nil {
		return fmt.Errorf("helm install cert-manager: %w", err)
	}

	log.Info("pki: cert-manager installed")
	return nil
}

func installStepIssuer(ctx context.Context, log *zap.Logger, restConfig *rest.Config, namespace string) error {
	log.Info("pki: installing step-issuer")

	settings := helmcli.New()
	repoEntry := &repo.Entry{
		Name: "smallstep",
		URL:  "https://smallstep.github.io/helm-charts",
	}

	chartPath, err := downloadChart(settings, repoEntry, "step-issuer", "")
	if err != nil {
		return fmt.Errorf("download step-issuer chart: %w", err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("load step-issuer chart: %w", err)
	}

	rcg := &restClientGetter{restConfig: restConfig, namespace: namespace}
	actionCfg := new(action.Configuration)
	if err := actionCfg.Init(rcg, namespace, "secret", func(format string, v ...interface{}) {
		log.Debug(fmt.Sprintf(format, v...), zap.String("component", "helm-step-issuer"))
	}); err != nil {
		return fmt.Errorf("init helm config: %w", err)
	}

	// Check if already installed
	histClient := action.NewHistory(actionCfg)
	histClient.Max = 1
	if _, err := histClient.Run("step-issuer"); err == nil {
		log.Info("pki: step-issuer already installed, skipping")
		return nil
	}

	installClient := action.NewInstall(actionCfg)
	installClient.ReleaseName = "step-issuer"
	installClient.Namespace = namespace
	installClient.CreateNamespace = false // Already created by cert-manager
	installClient.Timeout = 5 * time.Minute
	installClient.Wait = true

	if _, err := installClient.RunWithContext(ctx, chart, nil); err != nil {
		return fmt.Errorf("helm install step-issuer: %w", err)
	}

	log.Info("pki: step-issuer installed")
	return nil
}

func createStepIssuerRBAC(ctx context.Context, restConfig *rest.Config, namespace string) error {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	// ClusterRole
	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "aegis-step-issuer-approver",
			Labels: map[string]string{"app.kubernetes.io/managed-by": "aegis"},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"cert-manager.io"},
				Resources:     []string{"certificaterequests/approval"},
				Verbs:         []string{"update"},
				ResourceNames: []string{"stepclusterissuers.certmanager.step.sm/*", "stepissuers.certmanager.step.sm/*"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"signers"},
				Verbs:     []string{"approve"},
				ResourceNames: []string{
					"stepclusterissuers.certmanager.step.sm/*",
					"stepissuers.certmanager.step.sm/*",
				},
			},
		},
	}

	_, err = clientset.RbacV1().ClusterRoles().Create(ctx, role, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		_, err = clientset.RbacV1().ClusterRoles().Update(ctx, role, metav1.UpdateOptions{})
	}
	if err != nil {
		return fmt.Errorf("create ClusterRole: %w", err)
	}

	// ClusterRoleBinding
	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "aegis-step-issuer-approver",
			Labels: map[string]string{"app.kubernetes.io/managed-by": "aegis"},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "aegis-step-issuer-approver",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "cert-manager",
				Namespace: namespace,
			},
		},
	}

	_, err = clientset.RbacV1().ClusterRoleBindings().Create(ctx, binding, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		_, err = clientset.RbacV1().ClusterRoleBindings().Update(ctx, binding, metav1.UpdateOptions{})
	}
	return err
}

func createProvisionerPasswordSecret(ctx context.Context, restConfig *rest.Config, namespace, password string) error {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "step-ca-provisioner-password",
			Namespace: namespace,
			Labels:    map[string]string{"app.kubernetes.io/managed-by": "aegis"},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"password": []byte(password),
		},
	}

	_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		_, err = clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	}
	return err
}

func createStepClusterIssuer(ctx context.Context, log *zap.Logger, restConfig *rest.Config, cfg PKIConfig) error {
	dynClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("create dynamic client: %w", err)
	}

	// Use TLS CA if available, otherwise root CA
	caBundle := cfg.StepCATLSCAB64
	if caBundle == "" {
		caBundle = cfg.StepCARootCAB64
	}

	issuer := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "certmanager.step.sm/v1beta1",
			"kind":       "StepClusterIssuer",
			"metadata": map[string]interface{}{
				"name": cfg.ClusterIssuerName,
				"labels": map[string]interface{}{
					"app.kubernetes.io/managed-by": "aegis",
				},
			},
			"spec": map[string]interface{}{
				"url":      cfg.StepCAURL,
				"caBundle": caBundle,
				"provisioner": map[string]interface{}{
					"name": cfg.ProvisionerName,
					"kid":  cfg.ProvisionerKID,
					"passwordRef": map[string]interface{}{
						"name":      "step-ca-provisioner-password",
						"namespace": cfg.CertManagerNamespace,
						"key":       "password",
					},
				},
			},
		},
	}

	gvr := schema.GroupVersionResource{
		Group:    "certmanager.step.sm",
		Version:  "v1beta1",
		Resource: "stepclusterissuers",
	}

	_, err = dynClient.Resource(gvr).Create(ctx, issuer, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		existing, getErr := dynClient.Resource(gvr).Get(ctx, cfg.ClusterIssuerName, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("get existing StepClusterIssuer: %w", getErr)
		}
		issuer.SetResourceVersion(existing.GetResourceVersion())
		_, err = dynClient.Resource(gvr).Update(ctx, issuer, metav1.UpdateOptions{})
	}
	if err != nil {
		return fmt.Errorf("apply StepClusterIssuer: %w", err)
	}

	// Wait for issuer to be ready
	log.Info("pki: waiting for StepClusterIssuer to be ready")
	for i := 0; i < 30; i++ {
		obj, err := dynClient.Resource(gvr).Get(ctx, cfg.ClusterIssuerName, metav1.GetOptions{})
		if err == nil {
			conditions, found, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
			if found {
				for _, c := range conditions {
					cond, ok := c.(map[string]interface{})
					if !ok {
						continue
					}
					if cond["type"] == "Ready" && cond["status"] == "True" {
						log.Info("pki: StepClusterIssuer is ready")
						return nil
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

	log.Warn("pki: StepClusterIssuer not ready after 150s, proceeding anyway")
	return nil
}

// downloadChart downloads a Helm chart from a remote repository.
func downloadChart(settings *helmcli.EnvSettings, repoEntry *repo.Entry, chartName, version string) (string, error) {
	// Add repo to Helm's repo list so LocateChart can find it
	providers := getter.All(settings)
	r, err := repo.NewChartRepository(repoEntry, providers)
	if err != nil {
		return "", fmt.Errorf("create chart repository %s: %w", repoEntry.Name, err)
	}
	if _, err := r.DownloadIndexFile(); err != nil {
		return "", fmt.Errorf("download repo index for %s: %w", repoEntry.Name, err)
	}

	// Use ChartPathOptions with RepoURL to locate and download the chart
	chartPathOpts := action.ChartPathOptions{
		RepoURL: repoEntry.URL,
		Version: version,
	}

	cp, err := chartPathOpts.LocateChart(chartName, settings)
	if err != nil {
		return "", fmt.Errorf("locate chart %s from %s: %w", chartName, repoEntry.URL, err)
	}

	return cp, nil
}

// ReadHubCA reads the hub's root CA from the trust bundle secret.
func ReadHubCA(ctx context.Context, hubClient client.Client, namespace string) (string, error) {
	if namespace == "" {
		namespace = "aegis-system"
	}

	secret := &corev1.Secret{}
	if err := hubClient.Get(ctx, types.NamespacedName{Name: "aegis-trust-bundle", Namespace: namespace}, secret); err != nil {
		return "", fmt.Errorf("read trust bundle secret: %w", err)
	}

	caPEM, ok := secret.Data["ca.crt"]
	if !ok || len(caPEM) == 0 {
		return "", fmt.Errorf("trust bundle secret has no ca.crt key")
	}

	return base64.StdEncoding.EncodeToString(caPEM), nil
}
