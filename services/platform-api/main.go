package main

import (
	"context"
	"log"
	"os"

	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/platform-api/internal/server"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	grpcAddr := getenv("GRPC_ADDR", ":8081")
	httpAddr := getenv("HTTP_ADDR", ":8080")

	st := store.NewMemStore()
	svc := server.New(logger, st)

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
