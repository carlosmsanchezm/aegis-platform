package main

import (
	"context"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yourorg/aegis/agents/k8s-agent/internal/cpclient"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/discovery"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/kube"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/runtime"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	cp := getenv("AEGIS_CP_GRPC", "localhost:8081")
	clusterID := getenv("AEGIS_CLUSTER_ID", "dev-1")
	region := getenv("AEGIS_REGION", "us-local")
	provider := getenv("AEGIS_PROVIDER", "DEV")

	logger.Info("starting k8s agent", zap.String("cluster_id", clusterID), zap.String("region", region), zap.String("provider", provider), zap.String("endpoint", cp))

	client, err := cpclient.New(cp)
	if err != nil {
		logger.Fatal("failed to create control-plane client", zap.Error(err))
	}

	ctx := context.Background()
	if err := client.Register(ctx, &aegis.ClusterRegisterRequest{
		ClusterId: clusterID,
		Provider:  provider,
		Region:    region,
		Labels:    map[string]string{},
	}); err != nil {
		logger.Fatal("cluster registration failed", zap.Error(err))
	}

	flavors := discoverFlavors(ctx, logger)
	logger.Info("cluster registered", zap.String("cluster_id", clusterID), zap.String("flavors", flavorList(flavors)))

	go client.HeartbeatLoop(ctx, logger, clusterID, flavors)

	if getenv("AEGIS_EXECUTOR", "job") == "job" {
		orch, err := runtime.NewOrchestrator(logger, clusterID, client)
		if err != nil {
			logger.Fatal("orchestrator init failed", zap.Error(err))
		}
		orch.Run(ctx)
	} else {
		logger.Info("executor disabled; agent will only heartbeat")
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
		}
	}
}

func discoverFlavors(ctx context.Context, logger *zap.Logger) []*aegis.Flavor {
	cs, _, err := kube.New()
	if err != nil {
		logger.Warn("kube client init failed; using static flavors", zap.Error(err))
		return discovery.StaticFlavors()
	}
	flavors, err := discovery.DiscoverFlavors(ctx, cs)
	if err != nil || len(flavors) == 0 {
		if err != nil {
			logger.Warn("dynamic discovery failed; using static flavors", zap.Error(err))
		}
		return discovery.StaticFlavors()
	}
	return flavors
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func flavorList(f []*aegis.Flavor) string {
	if len(f) == 0 {
		return ""
	}
	names := make([]string, 0, len(f))
	for _, fl := range f {
		names = append(names, fl.GetName())
	}
	return strings.Join(names, ",")
}
