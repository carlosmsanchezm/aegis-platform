package ekstoken

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type mockPresigner struct {
	url string
	err error
}

func (m *mockPresigner) PresignGetCallerIdentity(_ context.Context, _ *sts.GetCallerIdentityInput, _ ...func(*sts.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &v4.PresignedHTTPRequest{URL: m.url}, nil
}

func TestTokenFormat(t *testing.T) {
	p := &Provider{Region: "us-east-1", ClusterName: "my-cluster"}
	mock := &mockPresigner{url: "https://sts.us-east-1.amazonaws.com/?Action=GetCallerIdentity&X-Amz-Algorithm=AWS4-HMAC-SHA256&x-k8s-aws-id=my-cluster"}

	token, expiry, err := p.TokenWithPresigner(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(token, tokenPrefix) {
		t.Errorf("token should start with %q, got %q", tokenPrefix, token[:20])
	}
	if expiry.IsZero() {
		t.Error("expiry should not be zero")
	}

	// Decode and verify the URL is embedded
	encoded := strings.TrimPrefix(token, tokenPrefix)
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}
	if string(decoded) != mock.url {
		t.Errorf("decoded URL mismatch: got %q, want %q", string(decoded), mock.url)
	}
}

func TestTokenRequiresRegion(t *testing.T) {
	p := &Provider{ClusterName: "my-cluster"}
	_, _, err := p.TokenWithPresigner(context.Background(), &mockPresigner{url: "https://example.com"})
	if err == nil || !strings.Contains(err.Error(), "region") {
		t.Errorf("expected region error, got: %v", err)
	}
}

func TestTokenRequiresClusterName(t *testing.T) {
	p := &Provider{Region: "us-east-1"}
	_, _, err := p.TokenWithPresigner(context.Background(), &mockPresigner{url: "https://example.com"})
	if err == nil || !strings.Contains(err.Error(), "cluster name") {
		t.Errorf("expected cluster name error, got: %v", err)
	}
}

func TestTokenPresignError(t *testing.T) {
	p := &Provider{Region: "us-east-1", ClusterName: "my-cluster"}
	mock := &mockPresigner{err: context.DeadlineExceeded}
	_, _, err := p.TokenWithPresigner(context.Background(), mock)
	if err == nil || !strings.Contains(err.Error(), "presign") {
		t.Errorf("expected presign error, got: %v", err)
	}
}
