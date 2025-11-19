package cpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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

	if enableTLS() {
		creds, err := buildTLSCredentials(endpoint)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{api: aegis.NewAegisPlatformClient(conn)}, nil
}

func enableTLS() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("AEGIS_CP_GRPC_INSECURE")), "false")
}

func buildTLSCredentials(endpoint string) (credentials.TransportCredentials, error) {
	tlsConfig := &tls.Config{}

	if skip := parseEnvBool("AEGIS_CP_GRPC_SKIP_VERIFY", false); skip {
		tlsConfig.InsecureSkipVerify = true
	} else {
		caPool, err := loadCustomCAPool()
		if err != nil {
			return nil, fmt.Errorf("load control-plane CA: %w", err)
		}
		if caPool != nil {
			tlsConfig.RootCAs = caPool
		}
	}

	if serverName := strings.TrimSpace(os.Getenv("AEGIS_CP_GRPC_SERVER_NAME")); serverName != "" {
		tlsConfig.ServerName = serverName
	} else if host, _, err := net.SplitHostPort(endpoint); err == nil && host != "" {
		tlsConfig.ServerName = host
	}

	return credentials.NewTLS(tlsConfig), nil
}

func loadCustomCAPool() (*x509.CertPool, error) {
	pemData, err := resolveCAPEM()
	if err != nil {
		return nil, err
	}
	if len(pemData) == 0 {
		return nil, nil
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemData) {
		return nil, fmt.Errorf("unable to parse CA PEM data")
	}
	return pool, nil
}

func resolveCAPEM() ([]byte, error) {
	if b64 := strings.TrimSpace(os.Getenv("AEGIS_PLATFORM_CA_B64")); b64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("decode AEGIS_PLATFORM_CA_B64: %w", err)
		}
		return decoded, nil
	}
	if pem := strings.TrimSpace(os.Getenv("AEGIS_PLATFORM_CA_PEM")); pem != "" {
		return []byte(pem), nil
	}
	if path := strings.TrimSpace(os.Getenv("AEGIS_PLATFORM_CA_FILE")); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read AEGIS_PLATFORM_CA_FILE: %w", err)
		}
		return data, nil
	}
	return nil, nil
}

func parseEnvBool(key string, def bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	val, err := strconv.ParseBool(raw)
	if err != nil {
		return def
	}
	return val
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

func (c *Client) SubmitWorkload(ctx context.Context, workload *aegis.Workload) (*aegis.Workload, error) {
	if workload == nil {
		return nil, fmt.Errorf("workload payload required")
	}
	return c.api.SubmitWorkload(ctx, &aegis.SubmitWorkloadRequest{Workload: workload})
}

func (c *Client) GetWorkload(ctx context.Context, id string) (*aegis.Workload, error) {
	if id == "" {
		return nil, fmt.Errorf("workload id required")
	}
	return c.api.GetWorkload(ctx, &aegis.GetWorkloadRequest{Id: id})
}
