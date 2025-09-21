package cpclient

import (
	"context"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

const (
	statusSucceeded = "SUCCEEDED"
)

type Client struct {
	api aegis.AegisPlatformClient
}

func New(endpoint string) (*Client, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials())) // TLS later
	if err != nil {
		return nil, err
	}
	return &Client{api: aegis.NewAegisPlatformClient(conn)}, nil
}

func (c *Client) Register(ctx context.Context, req *aegis.ClusterRegisterRequest) error {
	_, err := c.api.RegisterCluster(ctx, req)
	return err
}

func (c *Client) HeartbeatLoop(ctx context.Context, logger *zap.Logger, clusterID string, flavors []*aegis.Flavor) {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()

	leaseTicker := time.NewTicker(5 * time.Second)
	defer leaseTicker.Stop()

	ttf := 60.0
	if v := os.Getenv("AEGIS_TTFG_P50"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			ttf = f
		}
	}

	logger.Info("starting heartbeat loop", zap.String("cluster_id", clusterID), zap.Int("flavor_count", len(flavors)), zap.Float64("ttfg_p50_sec", ttf))

	for {
		select {
		case <-ctx.Done():
			logger.Info("heartbeat loop context canceled", zap.String("cluster_id", clusterID))
			return
		case <-t.C:
			logger.Info("sending heartbeat", zap.String("cluster_id", clusterID), zap.Float64("ttfg_p50_sec", ttf), zap.Int("flavor_count", len(flavors)))
			if _, err := c.api.Heartbeat(ctx, &aegis.ClusterHeartbeat{
				ClusterId:        clusterID,
				TtfGpuSecondsP50: ttf,
				AvailableFlavors: flavors,
			}); err != nil {
				logger.Warn("heartbeat failed", zap.String("cluster_id", clusterID), zap.Error(err))
				continue
			}
			logger.Info("heartbeat acknowledged", zap.String("cluster_id", clusterID))
		case <-leaseTicker.C:
			resp, err := c.api.LeaseWorkload(ctx, &aegis.LeaseWorkloadRequest{ClusterId: clusterID, Max: 1})
			if err != nil {
				logger.Warn("lease workload failed", zap.String("cluster_id", clusterID), zap.Error(err))
				continue
			}
			if resp == nil || len(resp.GetItems()) == 0 {
				logger.Debug("no workloads leased", zap.String("cluster_id", clusterID))
				continue
			}
			for _, w := range resp.GetItems() {
				logger.Info("leased workload", zap.String("cluster_id", clusterID), zap.String("workload_id", w.GetId()))
				if _, err := c.api.AckWorkload(ctx, &aegis.AckWorkloadRequest{Id: w.GetId(), Status: statusSucceeded}); err != nil {
					logger.Warn("ack workload failed", zap.String("cluster_id", clusterID), zap.String("workload_id", w.GetId()), zap.Error(err))
					continue
				}
				logger.Info("acked workload", zap.String("cluster_id", clusterID), zap.String("workload_id", w.GetId()), zap.String("status", statusSucceeded))
			}
		}
	}
}
