package spokeinstall

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	certManagerIssuerName  = "aegis-internal"
	certManagerIssuerKind  = "StepClusterIssuer"
	certManagerIssuerGroup = "certmanager.step.sm"
	certDuration           = 8760 * time.Hour // 1 year
	certRenewBefore        = 720 * time.Hour  // 30 days before expiry
)

// RequestSpokeProxyCert creates a cert-manager Certificate resource on the hub cluster
// to get a TLS cert for a spoke proxy, signed by the hub's step-ca. It waits for
// cert-manager to issue the cert and returns the PEM-encoded cert and key.
func RequestSpokeProxyCert(ctx context.Context, hubClient client.Client, clusterID, namespace string) (certPEM, keyPEM string, err error) {
	if namespace == "" {
		namespace = "aegis-system"
	}

	certName := fmt.Sprintf("aegis-spoke-proxy-%s", sanitize(clusterID))
	secretName := certName + "-tls"

	// Check if cert already exists
	existing := &certmanagerv1.Certificate{}
	err = hubClient.Get(ctx, types.NamespacedName{Name: certName, Namespace: namespace}, existing)
	if err == nil && existing.Status.Conditions != nil {
		// Cert exists — check if it's ready
		for _, cond := range existing.Status.Conditions {
			if cond.Type == certmanagerv1.CertificateConditionReady && cond.Status == cmmeta.ConditionTrue {
				// Cert is ready — read from secret
				return readCertFromSecret(ctx, hubClient, secretName, namespace)
			}
		}
	}

	if err != nil && !apierrors.IsNotFound(err) {
		return "", "", fmt.Errorf("check existing certificate: %w", err)
	}

	// Create the Certificate resource
	cert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "aegis",
				"aegis.yourorg.dev/cluster":    clusterID,
			},
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: secretName,
			Duration:   &metav1.Duration{Duration: certDuration},
			RenewBefore: &metav1.Duration{Duration: certRenewBefore},
			CommonName: fmt.Sprintf("spoke-proxy-%s.aegis.local", sanitize(clusterID)),
			DNSNames: []string{
				"*.nip.io",
				fmt.Sprintf("spoke-proxy-%s.aegis.local", sanitize(clusterID)),
				"spoke-proxy.aegis.local",
				"localhost",
			},
			Usages: []certmanagerv1.KeyUsage{
				certmanagerv1.UsageServerAuth,
				certmanagerv1.UsageDigitalSignature,
				certmanagerv1.UsageKeyEncipherment,
			},
			IssuerRef: cmmeta.ObjectReference{
				Name:  certManagerIssuerName,
				Kind:  certManagerIssuerKind,
				Group: certManagerIssuerGroup,
			},
		},
	}

	if apierrors.IsNotFound(err) {
		if err := hubClient.Create(ctx, cert); err != nil {
			return "", "", fmt.Errorf("create certificate resource: %w", err)
		}
	} else {
		// Update existing
		if err := hubClient.Update(ctx, cert); err != nil {
			return "", "", fmt.Errorf("update certificate resource: %w", err)
		}
	}

	// Wait for cert-manager to issue the cert (poll the secret)
	return waitForCert(ctx, hubClient, secretName, namespace, 2*time.Minute)
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

func waitForCert(ctx context.Context, hubClient client.Client, secretName, namespace string, timeout time.Duration) (string, string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		certPEM, keyPEM, err := readCertFromSecret(ctx, hubClient, secretName, namespace)
		if err == nil {
			return certPEM, keyPEM, nil
		}
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
	return "", "", fmt.Errorf("timeout waiting for cert-manager to issue cert %s/%s", namespace, secretName)
}

func readCertFromSecret(ctx context.Context, hubClient client.Client, secretName, namespace string) (string, string, error) {
	secret := &corev1.Secret{}
	if err := hubClient.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, secret); err != nil {
		return "", "", err
	}

	certPEM, ok := secret.Data["tls.crt"]
	if !ok || len(certPEM) == 0 {
		return "", "", fmt.Errorf("secret %s has no tls.crt", secretName)
	}

	keyPEM, ok := secret.Data["tls.key"]
	if !ok || len(keyPEM) == 0 {
		return "", "", fmt.Errorf("secret %s has no tls.key", secretName)
	}

	return string(certPEM), string(keyPEM), nil
}
