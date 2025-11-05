package mw

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/yourorg/aegis/services/platform-api/internal/config"
)

type contextKey struct {
	name string
}

var identityKey = &contextKey{name: "aegis-identity"}

var (
	metricAuthFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aegis_auth_failures_total",
			Help: "Authentication failures by reason",
		},
		[]string{"reason"},
	)
)

// Identity represents the authenticated subject derived from a validated access token.
type Identity struct {
	Subject           string
	Email             string
	PreferredUsername string
	ClientID          string
	Issuer            string
	Audience          []string
	Scope             []string
	AMR               []string
	ACR               string
	Roles             []string
	PhishingResistant bool
	TokenID           string
	KeyID             string
}

// TokenClaims captures the subset of OpenID Connect / Keycloak token claims we rely upon.
type TokenClaims struct {
	jwt.RegisteredClaims
	Scope             string   `json:"scope"`
	ClientID          string   `json:"client_id"`
	AuthorizedParty   string   `json:"azp"`
	Email             string   `json:"email"`
	PreferredUsername string   `json:"preferred_username"`
	AMR               []string `json:"amr"`
	ACR               string   `json:"acr"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
}

// RequestMeta contains contextual information about the inbound request for auditing.
type RequestMeta struct {
	Source     string
	Method     string
	RemoteAddr string
	RequestID  string
}

// Authenticator validates incoming Bearer tokens and exposes helpers for gRPC and HTTP middleware.
type Authenticator struct {
	cfg        *config.AuthConfig
	logger     *zap.Logger
	keys       *jwksCache
	allowedAMR map[string]struct{}
}

// NewAuthenticator constructs an authenticator backed by a JWKS cache.
func NewAuthenticator(cfg *config.AuthConfig, logger *zap.Logger) (*Authenticator, error) {
	if cfg == nil {
		return nil, errors.New("auth config is required")
	}
	cache := newJWKSCache(cfg, logger)
	allowed := make(map[string]struct{}, len(cfg.AllowedPhishingResistantAMR))
	for _, v := range cfg.AllowedPhishingResistantAMR {
		allowed[v] = struct{}{}
	}
	if logger != nil {
		logger.Info("OIDC auth configuration loaded",
			zap.String("issuer", cfg.IssuerURL),
			zap.String("jwks_url", cfg.JWKSURL),
			zap.String("ca_bundle", cfg.CABundlePath),
			zap.Bool("skip_tls_verify", cfg.SkipTLSVerify),
		)
	}
	return &Authenticator{
		cfg:        cfg,
		logger:     logger,
		keys:       cache,
		allowedAMR: allowed,
	}, nil
}

// UnaryServerInterceptor enforces authentication for unary RPCs.
func (a *Authenticator) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		identity, newCtx, err := a.authenticateUnary(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		if identity != nil {
			grpc.SetHeader(newCtx, metadata.Pairs("x-aegis-subject", identity.Subject))
		}
		return handler(newCtx, req)
	}
}

// StreamServerInterceptor enforces authentication for streaming RPCs.
func (a *Authenticator) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		identity, newCtx, err := a.authenticateUnary(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		if identity != nil {
			grpc.SetHeader(newCtx, metadata.Pairs("x-aegis-subject", identity.Subject))
		}
		wrapped := &authenticatedServerStream{ServerStream: ss, ctx: newCtx}
		return handler(srv, wrapped)
	}
}

// HTTPMiddleware enforces authentication for REST/HTTP traffic terminated by the gRPC gateway.
func (a *Authenticator) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearer(r.Header.Get("Authorization"))
		if token == "" {
			a.writeHTTPError(w, status.Error(codes.Unauthenticated, "missing bearer token"))
			metricAuthFailures.WithLabelValues("missing_token").Inc()
			return
		}
		meta := RequestMeta{
			Source:     "http",
			Method:     r.Method,
			RemoteAddr: clientIPFromRemoteAddr(r.RemoteAddr),
			RequestID:  r.Header.Get("X-Request-Id"),
		}
		identity, err := a.Authenticate(r.Context(), token, meta)
		if err != nil {
			a.writeHTTPError(w, err)
			return
		}
		md := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
		ctx := metadata.NewIncomingContext(r.Context(), md)
		ctx = ContextWithIdentity(ctx, identity)
		w.Header().Set("X-Aegis-Subject", identity.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authenticate validates the supplied bearer token and returns the enriched identity when successful.
func (a *Authenticator) Authenticate(ctx context.Context, token string, meta RequestMeta) (*Identity, error) {
	if strings.TrimSpace(token) == "" {
		metricAuthFailures.WithLabelValues("empty_token").Inc()
		return nil, status.Error(codes.Unauthenticated, "bearer token required")
	}

	claims := &TokenClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(tok *jwt.Token) (interface{}, error) {
		kid, _ := tok.Header["kid"].(string)
		key, keyErr := a.keys.getKey(ctx, kid)
		if keyErr != nil {
			return nil, keyErr
		}
		return key, nil
	},
		jwt.WithAudience(a.cfg.Audience),
		jwt.WithIssuer(a.cfg.IssuerURL),
		jwt.WithLeeway(30*time.Second),
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
	)
	if err != nil {
		a.logFailure(meta, "token_parse", err)
		metricAuthFailures.WithLabelValues("token_parse").Inc()
		return nil, status.Error(codes.Unauthenticated, "invalid bearer token")
	}
	if !t.Valid {
		a.logFailure(meta, "token_invalid", errors.New("token not valid"))
		metricAuthFailures.WithLabelValues("token_invalid").Inc()
		return nil, status.Error(codes.Unauthenticated, "invalid bearer token")
	}

	identity := buildIdentity(claims)
	if kid, ok := t.Header["kid"].(string); ok {
		identity.KeyID = strings.TrimSpace(kid)
	}

	if identity.Subject == "" {
		a.logFailure(meta, "missing_subject", errors.New("subject claim empty"))
		metricAuthFailures.WithLabelValues("missing_subject").Inc()
		return nil, status.Error(codes.PermissionDenied, "token missing subject claim")
	}

	if a.cfg.RequirePhishingResistantMFA && !hasAllowedFactor(identity.AMR, a.allowedAMR) {
		a.logFailure(meta, "mfa_required", errors.New("phishing resistant factor missing"))
		metricAuthFailures.WithLabelValues("mfa_required").Inc()
		return nil, status.Error(codes.PermissionDenied, "phishing-resistant MFA required")
	}

	identity.PhishingResistant = hasAllowedFactor(identity.AMR, a.allowedAMR)
	a.logSuccess(meta, identity)
	return identity, nil
}

func (a *Authenticator) authenticateUnary(ctx context.Context, method string) (*Identity, context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		metricAuthFailures.WithLabelValues("metadata_missing").Inc()
		return nil, ctx, status.Error(codes.Unauthenticated, "metadata missing")
	}
	var token string
	if authz := md.Get("authorization"); len(authz) > 0 {
		token = extractBearer(authz[0])
	}
	if token == "" {
		metricAuthFailures.WithLabelValues("missing_token").Inc()
		return nil, ctx, status.Error(codes.Unauthenticated, "missing bearer token")
	}

	var remote string
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		remote = clientIPFromRemoteAddr(p.Addr.String())
	}

	meta := RequestMeta{Source: "grpc", Method: method, RemoteAddr: remote}
	identity, err := a.Authenticate(ctx, token, meta)
	if err != nil {
		return nil, ctx, err
	}
	newCtx := ContextWithIdentity(ctx, identity)
	return identity, newCtx, nil
}

func (a *Authenticator) logSuccess(meta RequestMeta, id *Identity) {
	if a.logger == nil {
		return
	}
	fields := []zap.Field{
		zap.String("event", "auth_login"),
		zap.String("subject", id.Subject),
		zap.String("issuer", id.Issuer),
		zap.String("client_id", id.ClientID),
		zap.Strings("aud", id.Audience),
		zap.Strings("scope", id.Scope),
		zap.Strings("amr", id.AMR),
		zap.String("acr", id.ACR),
		zap.Strings("roles", id.Roles),
		zap.Bool("phishing_resistant", id.PhishingResistant),
		zap.String("token_id", id.TokenID),
		zap.String("key_id", id.KeyID),
		zap.String("decision", "allow"),
	}
	fields = append(fields, meta.zapFields()...)
	a.logger.Info("authentication succeeded", fields...)
}

func (a *Authenticator) logFailure(meta RequestMeta, reason string, err error) {
	if a.logger == nil {
		return
	}
	fields := []zap.Field{
		zap.String("event", "auth_failure"),
		zap.String("reason", reason),
		zap.String("decision", "deny"),
	}
	fields = append(fields, meta.zapFields()...)
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	a.logger.Warn("authentication failed", fields...)
}

func (meta RequestMeta) zapFields() []zap.Field {
	fields := []zap.Field{}
	if meta.Source != "" {
		fields = append(fields, zap.String("source", meta.Source))
	}
	if meta.Method != "" {
		fields = append(fields, zap.String("method", meta.Method))
	}
	if meta.RemoteAddr != "" {
		fields = append(fields, zap.String("source_ip", meta.RemoteAddr))
	}
	if meta.RequestID != "" {
		fields = append(fields, zap.String("request_id", meta.RequestID))
	}
	return fields
}

// ContextWithIdentity injects the authenticated identity into the context for downstream consumers.
func ContextWithIdentity(ctx context.Context, id *Identity) context.Context {
	if ctx == nil || id == nil {
		return ctx
	}
	return context.WithValue(ctx, identityKey, id)
}

// IdentityFromContext retrieves the stored identity from context.
func IdentityFromContext(ctx context.Context) *Identity {
	if ctx == nil {
		return nil
	}
	if raw := ctx.Value(identityKey); raw != nil {
		if id, ok := raw.(*Identity); ok {
			return id
		}
	}
	return nil
}

type authenticatedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authenticatedServerStream) Context() context.Context {
	return s.ctx
}

func buildIdentity(claims *TokenClaims) *Identity {
	if claims == nil {
		return &Identity{}
	}
	clientID := strings.TrimSpace(claims.ClientID)
	if clientID == "" {
		clientID = strings.TrimSpace(claims.AuthorizedParty)
	}
	aud := make([]string, len(claims.Audience))
	copy(aud, claims.Audience)
	rawScope := strings.Fields(strings.TrimSpace(claims.Scope))
	roles := aggregateRoles(claims)
	amr := normalizeStrings(claims.AMR)
	acr := strings.ToLower(strings.TrimSpace(claims.ACR))

	return &Identity{
		Subject:           strings.TrimSpace(claims.Subject),
		Email:             strings.TrimSpace(claims.Email),
		PreferredUsername: strings.TrimSpace(claims.PreferredUsername),
		ClientID:          clientID,
		Issuer:            strings.TrimSpace(claims.Issuer),
		Audience:          aud,
		Scope:             rawScope,
		AMR:               amr,
		ACR:               acr,
		Roles:             roles,
		TokenID:           strings.TrimSpace(claims.ID),
	}
}

func aggregateRoles(claims *TokenClaims) []string {
	if claims == nil {
		return nil
	}
	roleSet := map[string]struct{}{}
	for _, r := range claims.RealmAccess.Roles {
		if trimmed := strings.TrimSpace(strings.ToLower(r)); trimmed != "" {
			roleSet[trimmed] = struct{}{}
		}
	}
	for _, access := range claims.ResourceAccess {
		for _, r := range access.Roles {
			if trimmed := strings.TrimSpace(strings.ToLower(r)); trimmed != "" {
				roleSet[trimmed] = struct{}{}
			}
		}
	}
	if len(roleSet) == 0 {
		return nil
	}
	out := make([]string, 0, len(roleSet))
	for role := range roleSet {
		out = append(out, role)
	}
	sort.Strings(out)
	return out
}

func hasAllowedFactor(amr []string, allowed map[string]struct{}) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, v := range amr {
		if _, ok := allowed[v]; ok {
			return true
		}
	}
	return false
}

func normalizeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	set := map[string]struct{}{}
	for _, v := range values {
		trimmed := strings.TrimSpace(strings.ToLower(v))
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func extractBearer(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return strings.TrimSpace(parts[1])
	}
	if len(parts) == 1 && header != "" {
		return strings.TrimSpace(parts[0])
	}
	return ""
}

func clientIPFromRemoteAddr(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return strings.TrimSpace(addr)
	}
	return strings.TrimSpace(host)
}

func (a *Authenticator) writeHTTPError(w http.ResponseWriter, err error) {
	statusProto := status.Convert(err)
	code := statusProto.Code()
	var httpCode int
	switch code {
	case codes.PermissionDenied:
		httpCode = http.StatusForbidden
	case codes.InvalidArgument:
		httpCode = http.StatusBadRequest
	case codes.Unauthenticated:
		httpCode = http.StatusUnauthorized
	default:
		httpCode = http.StatusInternalServerError
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	payload := fmt.Sprintf("{\"error\":\"%s\",\"message\":\"%s\"}", code.String(), sanitize(statusProto.Message()))
	_, _ = w.Write([]byte(payload))
}

func sanitize(msg string) string {
	replacer := strings.NewReplacer("\"", "'", "\\", "")
	return replacer.Replace(msg)
}

// jwksCache maintains a cached JWKS document with simple TTL-based refresh.
type jwksCache struct {
	cfg     *config.AuthConfig
	logger  *zap.Logger
	client  *http.Client
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
}

func newJWKSCache(cfg *config.AuthConfig, logger *zap.Logger) *jwksCache {
	client := &http.Client{Timeout: 10 * time.Second}
	if customClient, err := buildJWKSHTTPClient(cfg, logger); err != nil {
		if logger != nil {
			logger.Warn("failed to build jwks http client, falling back to default", zap.Error(err))
		}
	} else if customClient != nil {
		client = customClient
	}
	return &jwksCache{
		cfg:    cfg,
		logger: logger,
		client: client,
		keys:   map[string]*rsa.PublicKey{},
	}
}

func (c *jwksCache) getKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	expired := time.Now().After(c.expires)
	c.mu.RUnlock()
	if ok && !expired {
		return key, nil
	}
	if err := c.refresh(ctx); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if kid != "" {
		if key, ok := c.keys[kid]; ok {
			return key, nil
		}
	}
	if len(c.keys) == 1 {
		for _, v := range c.keys {
			return v, nil
		}
	}
	return nil, fmt.Errorf("jwks key %q not found", kid)
}

func (c *jwksCache) refresh(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Now().Before(c.expires) && len(c.keys) > 0 {
		return nil
	}
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, c.cfg.JWKSURL, nil)
	if err != nil {
		return fmt.Errorf("failed to build JWKS request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks fetch returned status %d", resp.StatusCode)
	}
	var payload struct {
		Keys []json.RawMessage `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey, len(payload.Keys))
	for _, raw := range payload.Keys {
		var entry jwkEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			if c.logger != nil {
				c.logger.Debug("skip malformed jwk entry", zap.Error(err))
			}
			continue
		}
		if entry.Kty != "RSA" || entry.N == "" || entry.E == "" {
			continue
		}
		pub, err := entry.toPublicKey()
		if err != nil {
			if c.logger != nil {
				c.logger.Debug("skip jwk entry", zap.Error(err))
			}
			continue
		}
		kid := entry.Kid
		if kid == "" {
			kid = fmt.Sprintf("anon-%d", len(keys)+1)
		}
		keys[kid] = pub
	}
	if len(keys) == 0 {
		return fmt.Errorf("jwks contained no usable RSA keys")
	}
	c.keys = keys
	c.expires = time.Now().Add(c.cfg.JWKSCacheTTL)
	return nil
}

type jwkEntry struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (j jwkEntry) toPublicKey() (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}
	if len(eBytes) == 0 {
		return nil, errors.New("rsa exponent missing")
	}
	exponent := 0
	for _, b := range eBytes {
		exponent = exponent<<8 + int(b)
	}
	if exponent == 0 {
		return nil, errors.New("rsa exponent invalid")
	}
	modulus := new(big.Int).SetBytes(nBytes)
	return &rsa.PublicKey{N: modulus, E: exponent}, nil
}

func buildJWKSHTTPClient(cfg *config.AuthConfig, logger *zap.Logger) (*http.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("auth config required")
	}

	skipVerify := cfg.SkipTLSVerify
	caPath := strings.TrimSpace(cfg.CABundlePath)
	if !skipVerify && caPath == "" {
		return nil, nil
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	tlsConfig := transport.TLSClientConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}

	if skipVerify {
		if logger != nil {
			logger.Warn("OIDC_SKIP_TLS_VERIFY enabled; JWKS TLS verification disabled")
		}
		tlsConfig.InsecureSkipVerify = true
	}

	if caPath != "" {
		pool, err := x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		data, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read oidc ca bundle %q: %w", caPath, err)
		}
		if !pool.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("no certificates found in oidc ca bundle %q", caPath)
		}
		tlsConfig.RootCAs = pool
	}

	transport.TLSClientConfig = tlsConfig

	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}, nil
}
