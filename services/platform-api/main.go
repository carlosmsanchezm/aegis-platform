package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
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

	grpcAddr := getenv("GRPC_ADDR", ":8081")
	httpAddr := getenv("HTTP_ADDR", ":8080")

	// Initialize store based on AEGIS_STORE_BACKEND environment variable
	var st store.Store
	storeBackend := getenv("AEGIS_STORE_BACKEND", "memory")
	if storeBackend == "postgres" {
		dsn := buildPostgresDSN()
		var err error
		st, err = postgres.New(dsn, logger)
		if err != nil {
			logger.Fatal("failed to initialize postgres store", zap.Error(err))
		}
		logger.Info("using PostgreSQL store", zap.String("backend", "postgres"))
	} else {
		st = store.NewMemStore()
		logger.Info("using in-memory store", zap.String("backend", "memory"))
	}

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		logger.Fatal("failed to add client-go scheme", zap.Error(err))
	}
	if err := infraapi.AddToScheme(scheme); err != nil {
		logger.Fatal("failed to add management API scheme", zap.Error(err))
	}

	overlay := placement.NewPolicyOverlay()

	ctx, cancel := context.WithCancel(context.Background())
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
		provisioner := aws.NewRunner(logger)
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

	kubeconfigsDir := getenv("KUBECONFIGS_DIR", "/tmp/kubeconfigs")
	targetNamespace := getenv("AEGIS_NAMESPACE", "default")
	kubeClientManager := kubeclients.New(kubeconfigsDir)
	svc := server.New(logger, st, kubeClientManager, targetNamespace, overlay)

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
