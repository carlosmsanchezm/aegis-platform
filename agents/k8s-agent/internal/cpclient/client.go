package cpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

type Client struct {
	api  aegis.AegisPlatformClient
	conn *grpc.ClientConn
}

func New(endpoint string) (*Client, error) {
	var opts []grpc.DialOption

	tlsEnabled := enableTLS()
	if tlsEnabled {
		creds, err := buildTLSCredentials(endpoint)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if oidcCfg, err := loadOIDCConfigFromEnv(); err != nil {
		return nil, err
	} else if oidcCfg != nil {
		tokenSource, err := buildOIDCTokenSource(oidcCfg)
		if err != nil {
			return nil, err
		}
		log.Printf("cpclient: enabling OIDC client credentials for %s (skip_tls_verify=%t, custom_ca=%t)", oidcCfg.TokenURL, oidcCfg.SkipTLS, len(oidcCfg.CAPEM) > 0)
		opts = append(opts, grpc.WithPerRPCCredentials(&bearerTokenCredentials{
			source:            oauth2.ReuseTokenSource(nil, tokenSource),
			requireTLS:        tlsEnabled,
			headerKey:         "authorization",
			scheme:            "Bearer",
			metadataFormatter: strings.ToLower,
		}))
	}

	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		api:  aegis.NewAegisPlatformClient(conn),
		conn: conn,
	}, nil
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

// Close releases the underlying connection when the client is no longer needed.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// HeartbeatLoop continuously reports cluster health, advertised flavors, and spoke proxy URL.
func (c *Client) HeartbeatLoop(ctx context.Context, logger *zap.Logger, clusterID string, flavors []*aegis.Flavor, proxyURL string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	ttf := 60.0
	if v := os.Getenv("AEGIS_TTFG_P50"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			ttf = f
		}
	}

	logger.Info("starting heartbeat loop", zap.String("cluster_id", clusterID), zap.Int("flavor_count", len(flavors)), zap.Float64("ttfg_p50_sec", ttf), zap.String("proxy_url", proxyURL))

	for {
		select {
		case <-ctx.Done():
			logger.Info("heartbeat loop context canceled", zap.String("cluster_id", clusterID))
			return
		case <-ticker.C:
			logger.Debug("sending heartbeat", zap.String("cluster_id", clusterID), zap.Float64("ttfg_p50_sec", ttf), zap.Int("flavor_count", len(flavors)), zap.String("proxy_url", proxyURL))
			if _, err := c.api.Heartbeat(ctx, &aegis.ClusterHeartbeat{
				ClusterId:        clusterID,
				TtfGpuSecondsP50: ttf,
				AvailableFlavors: flavors,
				ProxyUrl:         proxyURL,
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

func (c *Client) ListClusterWorkloadIDs(ctx context.Context, clusterID string) ([]string, error) {
	if strings.TrimSpace(clusterID) == "" {
		return nil, fmt.Errorf("cluster id required")
	}
	resp, err := c.api.ListClusterWorkloadIDs(ctx, &aegis.ListClusterWorkloadIDsRequest{ClusterId: clusterID})
	if err != nil {
		return nil, err
	}
	return resp.GetWorkloadIds(), nil
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

type oidcConfig struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Audience     string
	SkipTLS      bool
	CAPEM        []byte
}

func loadOIDCConfigFromEnv() (*oidcConfig, error) {
	tokenURL := strings.TrimSpace(os.Getenv("AEGIS_CP_OIDC_TOKEN_URL"))
	if tokenURL == "" {
		return nil, nil
	}

	clientID := strings.TrimSpace(os.Getenv("AEGIS_CP_OIDC_CLIENT_ID"))
	if clientID == "" {
		return nil, fmt.Errorf("AEGIS_CP_OIDC_CLIENT_ID is required when AEGIS_CP_OIDC_TOKEN_URL is set")
	}
	clientSecret := strings.TrimSpace(os.Getenv("AEGIS_CP_OIDC_CLIENT_SECRET"))
	if clientSecret == "" {
		return nil, fmt.Errorf("AEGIS_CP_OIDC_CLIENT_SECRET is required when AEGIS_CP_OIDC_TOKEN_URL is set")
	}

	caPEM, err := resolveOIDCCAPEM()
	if err != nil {
		return nil, err
	}

	return &oidcConfig{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Audience:     strings.TrimSpace(os.Getenv("AEGIS_CP_OIDC_AUDIENCE")),
		SkipTLS:      parseEnvBool("AEGIS_CP_OIDC_SKIP_TLS_VERIFY", false),
		CAPEM:        caPEM,
	}, nil
}

func buildOIDCTokenSource(cfg *oidcConfig) (oauth2.TokenSource, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: cfg.SkipTLS}
	if len(cfg.CAPEM) > 0 {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(cfg.CAPEM) {
			return nil, fmt.Errorf("parse AEGIS_CP_OIDC_CA_B64: invalid PEM data")
		}
		tlsConfig.RootCAs = pool
	}

	httpClient := &http.Client{
		Timeout:   15 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	clientCfg := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenURL,
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	if cfg.Audience != "" {
		clientCfg.EndpointParams = url.Values{"audience": {cfg.Audience}}
	}

	return &clientCredentialsTokenSource{
		config:     clientCfg,
		httpClient: httpClient,
		timeout:    15 * time.Second,
	}, nil
}

type clientCredentialsTokenSource struct {
	config     *clientcredentials.Config
	httpClient *http.Client
	timeout    time.Duration
}

func (s *clientCredentialsTokenSource) Token() (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, s.httpClient)
	if s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}
	return s.config.Token(ctx)
}

type bearerTokenCredentials struct {
	source            oauth2.TokenSource
	requireTLS        bool
	headerKey         string
	scheme            string
	metadataFormatter func(string) string
	logOnce           sync.Once
}

func (c *bearerTokenCredentials) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	token, err := c.source.Token()
	if err != nil {
		c.logOnce.Do(func() {
			log.Printf("cpclient: OIDC token fetch failed: %v", err)
		})
		return nil, err
	}
	if token == nil || token.AccessToken == "" {
		return nil, fmt.Errorf("oidc token missing access token")
	}
	key := c.headerKey
	if c.metadataFormatter != nil {
		key = c.metadataFormatter(key)
	}
	return map[string]string{
		key: fmt.Sprintf("%s %s", c.scheme, token.AccessToken),
	}, nil
}

func (c *bearerTokenCredentials) RequireTransportSecurity() bool {
	return c.requireTLS
}

func resolveOIDCCAPEM() ([]byte, error) {
	if b64 := strings.TrimSpace(os.Getenv("AEGIS_CP_OIDC_CA_B64")); b64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("decode AEGIS_CP_OIDC_CA_B64: %w", err)
		}
		return decoded, nil
	}
	if path := strings.TrimSpace(os.Getenv("AEGIS_CP_OIDC_CA_FILE")); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read AEGIS_CP_OIDC_CA_FILE: %w", err)
		}
		return data, nil
	}
	return resolveCAPEM()
}
