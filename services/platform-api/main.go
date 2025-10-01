package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/platform-api/internal/kubeclients"
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

	kubeconfigsDir := getenv("KUBECONFIGS_DIR", "/tmp/kubeconfigs")
	targetNamespace := getenv("AEGIS_NAMESPACE", "default")
	kubeClientManager := kubeclients.New(kubeconfigsDir)
	svc := server.New(logger, st, kubeClientManager, targetNamespace)

	logger.Info("starting platform API", zap.String("grpc_addr", grpcAddr), zap.String("http_addr", httpAddr))

	ctx := context.Background()
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

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}
