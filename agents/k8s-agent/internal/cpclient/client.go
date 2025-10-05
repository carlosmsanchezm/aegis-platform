package cpclient

import (
	"context"
	"crypto/tls"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

type Client struct {
	api aegis.AegisPlatformClient
}

func New(endpoint string) (*Client, error) {
	var opts []grpc.DialOption

	// Check if TLS should be used (default: insecure for backward compatibility)
	if os.Getenv("AEGIS_CP_GRPC_INSECURE") != "false" {
		// Use insecure connection (default for in-cluster communication)
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// Use TLS with system cert pool and skip verification for self-signed certs
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // For self-signed certificates
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{api: aegis.NewAegisPlatformClient(conn)}, nil
}

func (c *Client) Register(ctx context.Context, req *aegis.ClusterRegisterRequest) error {
	_, err := c.api.RegisterCluster(ctx, req)
	return err
}

// HeartbeatLoop continuously reports cluster health and advertised flavors.
func (c *Client) HeartbeatLoop(ctx context.Context, logger *zap.Logger, clusterID string, flavors []*aegis.Flavor) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

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
		case <-ticker.C:
			logger.Debug("sending heartbeat", zap.String("cluster_id", clusterID), zap.Float64("ttfg_p50_sec", ttf), zap.Int("flavor_count", len(flavors)))
			if _, err := c.api.Heartbeat(ctx, &aegis.ClusterHeartbeat{
				ClusterId:        clusterID,
				TtfGpuSecondsP50: ttf,
				AvailableFlavors: flavors,
			}); err != nil {
				logger.Warn("heartbeat failed", zap.String("cluster_id", clusterID), zap.Error(err))
				continue
			}
			logger.Debug("heartbeat acknowledged", zap.String("cluster_id", clusterID))
		}
	}
}

func (c *Client) Lease(ctx context.Context, clusterID string, max int32) ([]*aegis.Workload, error) {
	resp, err := c.api.LeaseWorkload(ctx, &aegis.LeaseWorkloadRequest{ClusterId: clusterID, Max: max})
	if err != nil {
		return nil, err
	}
	return resp.GetItems(), nil
}

func (c *Client) Ack(ctx context.Context, id, status, backend, url string) error {
	_, err := c.api.AckWorkload(ctx, &aegis.AckWorkloadRequest{Id: id, Status: status, Backend: backend, Url: url})
	return err
}

func (c *Client) Start(ctx context.Context, id, clusterID string) error {
	_, err := c.api.StartWorkload(ctx, &aegis.StartWorkloadRequest{Id: id, ClusterId: clusterID})
	return err
}
