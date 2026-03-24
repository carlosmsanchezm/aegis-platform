package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	goruntime "runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clientconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/controllers"
	"github.com/yourorg/aegis/services/platform-api/internal/kubeclients"
	"github.com/yourorg/aegis/services/platform-api/internal/placement"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning/pulumi/aws"
	"github.com/yourorg/aegis/services/platform-api/internal/server"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
	"github.com/yourorg/aegis/services/platform-api/internal/store/postgres"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	logger.Info("PLATFORM_API_BUILD_VERSION_20251216_2030_RECONCILE_FIX")

	// FIPS 140-2 startup verification (SC-13)
	checkFIPSMode(logger)

	grpcAddr := getenv("GRPC_ADDR", ":8081")
	httpAddr := getenv("HTTP_ADDR", ":8080")

	// Initialize store based on AEGIS_STORE_BACKEND environment variable
	var st store.Store
	storeBackend := getenv("AEGIS_STORE_BACKEND", "postgres")
	if storeBackend == "postgres" {
		dsn := buildPostgresDSN()
		var err error
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			st, err = postgres.New(dsn, logger)
			if err == nil {
				break
			}
			logger.Warn("failed to connect to postgres, retrying...",
				zap.Int("attempt", i+1),
				zap.Int("max_retries", maxRetries),
				zap.Error(err))
			time.Sleep(time.Duration(1<<uint(i)) * time.Second) // exponential backoff: 1s, 2s, 4s, 8s, 16s
		}
		if err != nil {
			logger.Fatal("failed to connect to postgres after retries", zap.Error(err))
		}
		logger.Info("using PostgreSQL store", zap.String("backend", "postgres"))

		// Stale cluster cleanup uses two-phase detection (mark → delete)
		// to avoid false positives during pod restarts. Configured via
		// AEGIS_STALE_CLUSTER_THRESHOLD and AEGIS_CLEANUP_INTERVAL env vars.
		// Started below after context is created.
	} else {
		st = store.NewMemStore()
		logger.Warn("WARNING: using in-memory store — data will not persist across restarts. Set AEGIS_STORE_BACKEND=postgres for production use.",
			zap.String("backend", "memory"))
	}

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		logger.Fatal("failed to add client-go scheme", zap.Error(err))
	}
	if err := infraapi.AddToScheme(scheme); err != nil {
		logger.Fatal("failed to add management API scheme", zap.Error(err))
	}

	overlay := placement.NewPolicyOverlay()

	// Set up signal handling for graceful shutdown.
	// Pulumi operations can take several minutes, so we give them a grace period
	// before forcefully cancelling. The grace period is configurable via
	// AEGIS_SHUTDOWN_GRACE_PERIOD (default: 60s).
	ctx, cancel := context.WithCancel(context.Background())
	gracePeriod := 60 * time.Second
	if gp := os.Getenv("AEGIS_SHUTDOWN_GRACE_PERIOD"); gp != "" {
		if parsed, err := time.ParseDuration(gp); err == nil {
			gracePeriod = parsed
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("received shutdown signal, starting graceful shutdown",
			zap.String("signal", sig.String()),
			zap.Duration("grace_period", gracePeriod))

		// Start the grace period timer
		graceTicker := time.NewTimer(gracePeriod)
		defer graceTicker.Stop()

		// Cancel context after grace period
		select {
		case <-graceTicker.C:
			logger.Warn("grace period expired, forcing shutdown")
		case sig := <-sigCh:
			logger.Warn("received second signal, forcing immediate shutdown",
				zap.String("signal", sig.String()))
		}
		cancel()
	}()
	defer cancel()

	var mgr ctrl.Manager
	if getenvBool("AEGIS_INFRA_CONTROLLER", false) {
		managerOpts := ctrl.Options{
			Scheme: scheme,
			Metrics: metricsserver.Options{
				BindAddress: getenv("AEGIS_CONTROLLER_METRICS_ADDR", "0"),
			},
		}
		var err error
		mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), managerOpts)
		if err != nil {
			logger.Fatal("failed to initialize controller manager", zap.Error(err))
		}
		provisioner := aws.NewRunner(logger, st)
		controllerCfg := controllers.Config{
			Logger:                    logger,
			Store:                     st,
			Provisioner:               provisioner,
			KubeconfigSecretName:      getenv("AEGIS_KUBECONFIG_SECRET_NAME", ""),
			KubeconfigSecretNamespace: getenv("AEGIS_KUBECONFIG_SECRET_NAMESPACE", ""),
			PolicyOverlay:             overlay,
		}
		if err := controllers.SetupWithManager(mgr, controllerCfg); err != nil {
			logger.Fatal("failed to register infrastructure controllers", zap.Error(err))
		}
		go func() {
			if err := mgr.Start(ctx); err != nil {
				if err != context.Canceled {
					logger.Error("controller manager exited", zap.Error(err))
				}
				cancel()
			}
		}()
	} else {
		logger.Info("management-plane controllers disabled; AEGIS_INFRA_CONTROLLER!=true")
	}

	infraNamespace := getenv("AEGIS_INFRA_NAMESPACE", "")
	var infraClient client.Client
	if cfg, err := clientconfig.GetConfig(); err != nil {
		logger.Warn("kubeconfig not available; cluster provisioning disabled", zap.Error(err))
	} else {
		infraClient, err = client.New(cfg, client.Options{Scheme: scheme})
		if err != nil {
			logger.Warn("failed to initialize infra client", zap.Error(err))
		}
	}

	kubeconfigsDir := getenv("KUBECONFIGS_DIR", "/tmp/kubeconfigs")
	targetNamespace := getenv("AEGIS_NAMESPACE", "default")
	kubeClientManager := kubeclients.New(kubeconfigsDir)
	if infraClient != nil {
		secretName := getenv("AEGIS_KUBECONFIG_SECRET_NAME", "aegis-kubeconfigs")
		secretNS := getenv("AEGIS_KUBECONFIG_SECRET_NAMESPACE", "aegis-system")
		kubeClientManager.WithSecretFallback(infraClient, secretName, secretNS)
		logger.Info("kubeclients: Secret fallback enabled",
			zap.String("secret", secretNS+"/"+secretName))
	}
	svc := server.New(logger, st, kubeClientManager, targetNamespace, overlay, infraClient, infraNamespace)
	kubeClientManager.WithAuthProvider(svc)
	logger.Info("kubeclients: EKS token auth provider enabled")

	// Start two-phase stale cluster cleanup (mark stale → delete after grace period).
	cleanup := controllers.DefaultCleanupConfig(logger, st)
	go cleanup.Start(ctx)

	logger.Info("starting platform API", zap.String("grpc_addr", grpcAddr), zap.String("http_addr", httpAddr))

	if err := server.Run(ctx, logger, grpcAddr, httpAddr, svc); err != nil {
		logger.Error("platform API server exited", zap.Error(err))
		log.Fatal(err)
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func buildPostgresDSN() string {
	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "postgres")
	password := os.Getenv("DB_PASSWORD") // Required, no default
	dbname := getenv("DB_NAME", "postgres")
	sslmode := getenv("DB_SSLMODE", "disable")

	// URL-encode password to handle special characters
	encodedPassword := url.QueryEscape(password)

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, encodedPassword, host, port, dbname, sslmode)
}

// checkFIPSMode verifies BoringCrypto is active at startup (SC-13 FIPS 140-2).
// When AEGIS_FIPS_ENABLED=true but BoringCrypto is not linked, the process
// exits immediately (fail-closed).
func checkFIPSMode(logger *zap.Logger) {
	boringActive := strings.Contains(goruntime.Version(), "boringcrypto")
	if boringActive {
		logger.Info("FIPS mode: enabled (BoringCrypto)", zap.String("go_version", goruntime.Version()))
	} else {
		logger.Info("FIPS mode: disabled (standard Go crypto)", zap.String("go_version", goruntime.Version()))
	}
	if os.Getenv("AEGIS_FIPS_ENABLED") == "true" && !boringActive {
		logger.Fatal("AEGIS_FIPS_ENABLED=true but BoringCrypto is not active; binary must be built with GOEXPERIMENT=boringcrypto")
	}
}

func getenvBool(key string, def bool) bool {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	if parsed, err := strconv.ParseBool(val); err == nil {
		return parsed
	}
	return def
}
