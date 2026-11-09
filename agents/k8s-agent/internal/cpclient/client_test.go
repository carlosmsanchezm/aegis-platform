package cpclient

import (
	"context"
	"encoding/base64"
	"encoding/pem"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

func TestClientAttachesOIDCBearerToken(t *testing.T) {
	t.Setenv("AEGIS_CP_GRPC_INSECURE", "true")

	tokenRequests := 0
	tokenServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenRequests++
		w.Header().Set("Content-Type", "application/json")
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.Form.Get("client_id"); got != "agent-client" {
			t.Fatalf("unexpected client_id: %q", got)
		}
		if got := r.Form.Get("client_secret"); got != "super-secret" {
			t.Fatalf("unexpected client_secret: %q", got)
		}
		if got := r.Form.Get("grant_type"); got != "client_credentials" {
			t.Fatalf("unexpected grant_type: %q", got)
		}
		if got := r.Form.Get("audience"); got != "platform-api" {
			t.Fatalf("unexpected audience: %q", got)
		}
		_, _ = w.Write([]byte(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenServer.Close()

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: tokenServer.Certificate().Raw})
	t.Setenv("AEGIS_CP_OIDC_TOKEN_URL", tokenServer.URL)
	t.Setenv("AEGIS_CP_OIDC_CLIENT_ID", "agent-client")
	t.Setenv("AEGIS_CP_OIDC_CLIENT_SECRET", "super-secret")
	t.Setenv("AEGIS_CP_OIDC_AUDIENCE", "platform-api")
	t.Setenv("AEGIS_CP_OIDC_CA_B64", base64.StdEncoding.EncodeToString(certPEM))
	t.Setenv("AEGIS_CP_OIDC_SKIP_TLS_VERIFY", "true")

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer lis.Close()

	var observedAuth string
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	aegis.RegisterAegisPlatformServer(grpcServer, &stubPlatformServer{
		onRegister: func(ctx context.Context, _ *aegis.ClusterRegisterRequest) (*aegis.ClusterRegisterResponse, error) {
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				if vals := md.Get("authorization"); len(vals) > 0 {
					observedAuth = vals[0]
				}
			}
			return &aegis.ClusterRegisterResponse{}, nil
		},
	})
	go func() {
		_ = grpcServer.Serve(lis)
	}()

	client, err := New(lis.Addr().String())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Register(ctx, &aegis.ClusterRegisterRequest{ClusterId: "cluster-1"}); err != nil {
		t.Fatalf("register cluster: %v", err)
	}

	if observedAuth != "Bearer test-token" {
		t.Fatalf("authorization header not attached, got %q", observedAuth)
	}
	if tokenRequests == 0 {
		t.Fatalf("expected token endpoint to be called")
	}
}

func TestLoadOIDCConfigFromEnvRequiresClientCredentials(t *testing.T) {
	t.Setenv("AEGIS_CP_OIDC_TOKEN_URL", "https://issuer.example/token")
	t.Setenv("AEGIS_CP_OIDC_CLIENT_ID", "")
	t.Setenv("AEGIS_CP_OIDC_CLIENT_SECRET", "")

	if _, err := loadOIDCConfigFromEnv(); err == nil {
		t.Fatalf("expected validation error for missing client credentials")
	}
}

type stubPlatformServer struct {
	aegis.UnimplementedAegisPlatformServer
	onRegister func(ctx context.Context, req *aegis.ClusterRegisterRequest) (*aegis.ClusterRegisterResponse, error)
}

func (s *stubPlatformServer) RegisterCluster(ctx context.Context, req *aegis.ClusterRegisterRequest) (*aegis.ClusterRegisterResponse, error) {
	if s.onRegister != nil {
		return s.onRegister(ctx, req)
	}
	return &aegis.ClusterRegisterResponse{}, nil
}
