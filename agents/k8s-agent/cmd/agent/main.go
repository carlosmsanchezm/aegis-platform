package main

import (
	"context"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/yourorg/aegis/agents/k8s-agent/internal/cpclient"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/discovery"
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

	logger.Info("dialing control plane", zap.String("endpoint", cp))
	client, err := cpclient.New(cp)
	if err != nil {
		logger.Fatal("failed to create control-plane client", zap.Error(err))
	}

	ctx := context.Background()
	flavors := discovery.StaticFlavors()
	logger.Info("discovered static flavors", zap.String("cluster_id", clusterID), zap.String("flavors", flavorList(flavors)))

	logger.Info("registering cluster", zap.String("cluster_id", clusterID))
	err = client.Register(ctx, &aegis.ClusterRegisterRequest{
		ClusterId: clusterID, Provider: provider, Region: region, Labels: map[string]string{"gpu.chip": "A10"},
	})
	if err != nil {
		logger.Fatal("cluster registration failed", zap.Error(err))
	}

	logger.Info("cluster registered", zap.String("cluster_id", clusterID))

	client.HeartbeatLoop(ctx, logger, clusterID, flavors)
	logger.Info("heartbeat loop exited", zap.String("cluster_id", clusterID))
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
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
