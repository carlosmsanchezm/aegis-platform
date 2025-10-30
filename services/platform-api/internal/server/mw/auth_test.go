package mw

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yourorg/aegis/services/platform-api/internal/config"
)

func TestAuthenticateSuccess(t *testing.T) {
	auth, key, cleanup := newTestAuthenticator(t, true)
	defer cleanup()

	token := signedToken(t, key, tokenClaims{
		Issuer:   auth.cfg.IssuerURL,
		Subject:  "alice@example.com",
		Audience: []string{auth.cfg.Audience},
		Expires:  time.Now().Add(1 * time.Hour),
		AMR:      []string{"piv"},
		TokenID:  "token-123",
	})
	meta := RequestMeta{Source: "test", RemoteAddr: "127.0.0.1"}
	identity, err := auth.Authenticate(context.Background(), token, meta)
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if identity.Subject != "alice@example.com" {
		t.Fatalf("expected subject alice@example.com, got %q", identity.Subject)
	}
	if !identity.PhishingResistant {
		t.Fatalf("expected phishing resistant token")
	}
	if identity.TokenID != "token-123" {
		t.Fatalf("expected token id to be propagated, got %q", identity.TokenID)
	}
	if identity.KeyID != "test-key" {
		t.Fatalf("expected key id to be propagated, got %q", identity.KeyID)
	}
}

func TestAuthenticateInvalidIssuer(t *testing.T) {
	auth, key, cleanup := newTestAuthenticator(t, false)
	defer cleanup()

	token := signedToken(t, key, tokenClaims{
		Issuer:   "https://evil.example.com",
		Subject:  "bob@example.com",
		Audience: []string{auth.cfg.Audience},
		Expires:  time.Now().Add(time.Hour),
		AMR:      []string{"hwk"},
	})
	_, err := auth.Authenticate(context.Background(), token, RequestMeta{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestAuthenticateInvalidAudience(t *testing.T) {
	auth, key, cleanup := newTestAuthenticator(t, false)
	defer cleanup()

	token := signedToken(t, key, tokenClaims{
		Issuer:   auth.cfg.IssuerURL,
		Subject:  "charlie@example.com",
		Audience: []string{"other"},
		Expires:  time.Now().Add(time.Hour),
		AMR:      []string{"hwk"},
	})
	_, err := auth.Authenticate(context.Background(), token, RequestMeta{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestAuthenticateExpired(t *testing.T) {
	auth, key, cleanup := newTestAuthenticator(t, false)
	defer cleanup()

	token := signedToken(t, key, tokenClaims{
		Issuer:   auth.cfg.IssuerURL,
		Subject:  "dana@example.com",
		Audience: []string{auth.cfg.Audience},
		Expires:  time.Now().Add(-1 * time.Hour),
		AMR:      []string{"hwk"},
	})
	_, err := auth.Authenticate(context.Background(), token, RequestMeta{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestAuthenticateMissingAMR(t *testing.T) {
	auth, key, cleanup := newTestAuthenticator(t, true)
	defer cleanup()

	token := signedToken(t, key, tokenClaims{
		Issuer:   auth.cfg.IssuerURL,
		Subject:  "erin@example.com",
		Audience: []string{auth.cfg.Audience},
		Expires:  time.Now().Add(time.Hour),
	})
	_, err := auth.Authenticate(context.Background(), token, RequestMeta{})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected permission denied, got %v", err)
	}
}

func TestAuthenticateUsesAzpFallback(t *testing.T) {
	auth, key, cleanup := newTestAuthenticator(t, false)
	defer cleanup()

	token := signedToken(t, key, tokenClaims{
		Issuer:          auth.cfg.IssuerURL,
		Subject:         "fallback@example.com",
		Audience:        []string{auth.cfg.Audience},
		Expires:         time.Now().Add(time.Hour),
		AMR:             []string{"hwk"},
		ClientID:        "",
		AuthorizedParty: "backstage",
	})
	identity, err := auth.Authenticate(context.Background(), token, RequestMeta{})
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if identity.ClientID != "backstage" {
		t.Fatalf("expected client id to fall back to azp, got %q", identity.ClientID)
	}
}

func TestUnaryInterceptorMissingToken(t *testing.T) {
	auth, _, cleanup := newTestAuthenticator(t, false)
	defer cleanup()

	interceptor := auth.UnaryServerInterceptor()
	ctx := context.Background()
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "test"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return struct{}{}, nil
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

// --- helpers ---

type tokenClaims struct {
	Issuer          string
	Subject         string
	Audience        []string
	Expires         time.Time
	AMR             []string
	TokenID         string
	ClientID        string
	AuthorizedParty string
}

func signedToken(t *testing.T, key *rsa.PrivateKey, claims tokenClaims) string {
	t.Helper()
	rc := jwt.RegisteredClaims{
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
		Audience:  claims.Audience,
		ExpiresAt: jwt.NewNumericDate(claims.Expires),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	if claims.TokenID != "" {
		rc.ID = claims.TokenID
	}
	tc := &TokenClaims{
		RegisteredClaims: rc,
		AMR:              claims.AMR,
		ClientID:         claims.ClientID,
		AuthorizedParty:  claims.AuthorizedParty,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, tc)
	token.Header["kid"] = "test-key"
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func newTestAuthenticator(t *testing.T, requireMFA bool) (*Authenticator, *rsa.PrivateKey, func()) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	jwksHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwks := jwksJSON(t, key.Public().(*rsa.PublicKey))
		_ = json.NewEncoder(w).Encode(jwks)
	})
	server := httptest.NewServer(jwksHandler)

	issuer := "https://issuer.test"
	cfg := &config.AuthConfig{
		IssuerURL:                   issuer,
		Audience:                    "aegis-client",
		JWKSURL:                     server.URL,
		JWKSCacheTTL:                time.Minute,
		JWKSRefreshInterval:         30 * time.Second,
		RequirePhishingResistantMFA: requireMFA,
		AllowedPhishingResistantAMR: []string{"hwk", "piv", "webauthn"},
	}
	logger := zap.NewNop()
	auth, err := NewAuthenticator(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	cleanup := func() {
		server.Close()
	}
	return auth, key, cleanup
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func jwksJSON(t *testing.T, key *rsa.PublicKey) jwksDocument {
	t.Helper()
	n := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(bigIntToBytes(key.E))
	return jwksDocument{Keys: []jwk{{
		Kty: "RSA",
		Kid: "test-key",
		Use: "sig",
		Alg: "RS256",
		N:   n,
		E:   e,
	}}}
}

func bigIntToBytes(e int) []byte {
	out := make([]byte, 0, 4)
	temp := e
	for temp > 0 {
		out = append([]byte{byte(temp & 0xff)}, out...)
		temp >>= 8
	}
	if len(out) == 0 {
		return []byte{0x01}
	}
	return out
}
