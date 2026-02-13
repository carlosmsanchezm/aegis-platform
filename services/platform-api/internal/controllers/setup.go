package controllers

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/yourorg/aegis/services/platform-api/internal/placement"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

// Config captures the shared dependencies required by the controller set.
type Config struct {
	Logger                    *zap.Logger
	Store                     store.Store
	Provisioner               provisioning.AWSProvisioner
	KubeconfigSecretName      string
	KubeconfigSecretNamespace string
	PolicyOverlay             *placement.PolicyOverlay
	StatusSyncInterval        time.Duration
}

// SetupWithManager wires all management-plane controllers into the provided manager.
func SetupWithManager(mgr ctrl.Manager, cfg Config) error {
	logger := cfg.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	mgrOpts := controller.Options{
		MaxConcurrentReconciles: 5,
	}
	secretName := cfg.KubeconfigSecretName
	if secretName == "" {
		secretName = "aegis-kubeconfigs"
	}
	secretNamespace := cfg.KubeconfigSecretNamespace
	if secretNamespace == "" {
		secretNamespace = "aegis-system"
	}
	policyOverlay := cfg.PolicyOverlay
	if policyOverlay == nil {
		policyOverlay = placement.NewPolicyOverlay()
	}
	interval := cfg.StatusSyncInterval
	if interval <= 0 {
		interval = 30 * time.Second
	}

	projectInfra := &ProjectInfraReconciler{
		Client:                    mgr.GetClient(),
		APIReader:                 mgr.GetAPIReader(), // Uncached reader for critical status checks
		Scheme:                    mgr.GetScheme(),
		Log:                       logger.Named("projectinfra"),
		Provisioner:               cfg.Provisioner,
		Store:                     cfg.Store,
		KubeconfigSecretName:      secretName,
		KubeconfigSecretNamespace: secretNamespace,
		LocalLocks:                map[string]*sync.Mutex{},
		HolderIdentity:            hostname(),
	}
	if err := projectInfra.SetupWithManager(mgr, mgrOpts); err != nil {
		return err
	}

	projectPlacement := &ProjectPlacementReconciler{
		Client:        mgr.GetClient(),
		Log:           logger.Named("projectplacement"),
		PolicyOverlay: policyOverlay,
	}
	if err := projectPlacement.SetupWithManager(mgr); err != nil {
		return err
	}

	statusSync := &AegisClusterStatusSync{
		Client:    mgr.GetClient(),
		Log:       logger.Named("clusterstatus"),
		Store:     cfg.Store,
		Namespace: secretNamespace,
		Interval:  interval,
	}
	if err := mgr.Add(statusSync); err != nil {
		return err
	}

	return nil
}

func hostname() string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "platform-api"
}
