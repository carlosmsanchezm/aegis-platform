package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadAuthConfigFromEnv(t *testing.T) {
	t.Setenv("AUTH_CONFIG_JSON", "")
	t.Setenv("OIDC_ISSUER_URL", " https://keycloak.localtest.me/realms/aegis/ ")
	t.Setenv("OIDC_AUDIENCE", " backstage ")
	t.Setenv("OIDC_JWKS_CACHE_TTL", "90s")
	t.Setenv("OIDC_JWKS_REFRESH_INTERVAL", " 45s ")
	t.Setenv("ALLOWED_PHISHING_RESISTANT_AMR", " piv , HWK ,webauthn , piv ")
	t.Setenv("REQUIRE_PHISHING_RESISTANT_MFA", "YeS")

	cfg, err := LoadAuthConfig()
	if err != nil {
		t.Fatalf("LoadAuthConfig returned error: %v", err)
	}

	if cfg.IssuerURL != "https://keycloak.localtest.me/realms/aegis" {
		t.Fatalf("unexpected issuer url %q", cfg.IssuerURL)
	}
	if cfg.Audience != "backstage" {
		t.Fatalf("unexpected audience %q", cfg.Audience)
	}
	if cfg.JWKSURL != "https://keycloak.localtest.me/realms/aegis/protocol/openid-connect/certs" {
		t.Fatalf("unexpected jwks url %q", cfg.JWKSURL)
	}
	if cfg.JWKSCacheTTL != 90*time.Second {
		t.Fatalf("unexpected cache ttl %v", cfg.JWKSCacheTTL)
	}
	if cfg.JWKSRefreshInterval != 45*time.Second {
		t.Fatalf("unexpected refresh interval %v", cfg.JWKSRefreshInterval)
	}
	if !cfg.RequirePhishingResistantMFA {
		t.Fatal("expected phishing resistant MFA requirement to be enabled")
	}
	expectedAMR := []string{"hwk", "piv", "webauthn"}
	if len(cfg.AllowedPhishingResistantAMR) != len(expectedAMR) {
		t.Fatalf("unexpected amr list %v", cfg.AllowedPhishingResistantAMR)
	}
	for _, factor := range expectedAMR {
		if !contains(cfg.AllowedPhishingResistantAMR, factor) {
			t.Fatalf("amr factor %q missing from %v", factor, cfg.AllowedPhishingResistantAMR)
		}
	}
}

func TestLoadAuthConfigJSONOverlay(t *testing.T) {
	t.Setenv("OIDC_ISSUER_URL", "")
	t.Setenv("OIDC_AUDIENCE", "")
	t.Setenv("AUTH_CONFIG_JSON", `{
		"issuerUrl": "https://sso.example.mil/realms/aegis",
		"audience": "aegis-platform",
		"jwksUrl": "https://jwks.example.mil/aegis.json",
		"jwksCacheTtl": "2m",
		"jwksRefreshInterval": "30s",
		"requirePhishingResistantMfa": false,
		"allowedPhishingResistantAmr": ["PIV", "webauthn"]
	}`)

	cfg, err := LoadAuthConfig()
	if err != nil {
		t.Fatalf("LoadAuthConfig returned error: %v", err)
	}

	if cfg.IssuerURL != "https://sso.example.mil/realms/aegis" {
		t.Fatalf("unexpected issuer url %q", cfg.IssuerURL)
	}
	if cfg.Audience != "aegis-platform" {
		t.Fatalf("unexpected audience %q", cfg.Audience)
	}
	if cfg.JWKSURL != "https://jwks.example.mil/aegis.json" {
		t.Fatalf("unexpected jwks url %q", cfg.JWKSURL)
	}
	if cfg.JWKSCacheTTL != 2*time.Minute {
		t.Fatalf("unexpected cache ttl %v", cfg.JWKSCacheTTL)
	}
	if cfg.JWKSRefreshInterval != 30*time.Second {
		t.Fatalf("unexpected refresh interval %v", cfg.JWKSRefreshInterval)
	}
	if cfg.RequirePhishingResistantMFA {
		t.Fatal("expected phishing resistant MFA requirement to be disabled via JSON")
	}
	if len(cfg.AllowedPhishingResistantAMR) != 2 ||
		cfg.AllowedPhishingResistantAMR[0] != "piv" ||
		cfg.AllowedPhishingResistantAMR[1] != "webauthn" {
		t.Fatalf("unexpected amr values %v", cfg.AllowedPhishingResistantAMR)
	}
}

func TestLoadAuthConfigMissingIssuerFails(t *testing.T) {
	t.Setenv("AUTH_CONFIG_JSON", "")
	unset := []string{
		"OIDC_ISSUER_URL",
		"OIDC_AUDIENCE",
		"OIDC_JWKS_URL",
		"OIDC_JWKS_CACHE_TTL",
		"OIDC_JWKS_REFRESH_INTERVAL",
	}
	for _, key := range unset {
		t.Setenv(key, "")
	}
	t.Setenv("OIDC_AUDIENCE", "aegis-api")

	_, err := LoadAuthConfig()
	if err == nil || !strings.Contains(err.Error(), "issuer") {
		t.Fatalf("expected issuer error, got %v", err)
	}

	t.Setenv("OIDC_ISSUER_URL", "https://issuer.example")
	t.Setenv("OIDC_AUDIENCE", "")
	_, err = LoadAuthConfig()
	if err == nil || !strings.Contains(err.Error(), "audience") {
		t.Fatalf("expected audience error, got %v", err)
	}
}

func contains(values []string, candidate string) bool {
	for _, v := range values {
		if v == candidate {
			return true
		}
	}
	return false
}

func TestNormalizeStringsDeduplication(t *testing.T) {
	clean := normalizeStrings([]string{" PIV ", "piv", "WEBAuthn", "", " "})
	if len(clean) != 2 {
		t.Fatalf("expected 2 normalized values, got %v", clean)
	}
	if !contains(clean, "piv") || !contains(clean, "webauthn") {
		t.Fatalf("unexpected normalized values %v", clean)
	}
}

func TestParseBoolErrors(t *testing.T) {
	t.Setenv("REQUIRE_PHISHING_RESISTANT_MFA", "definitely")
	t.Setenv("OIDC_ISSUER_URL", "https://issuer")
	t.Setenv("OIDC_AUDIENCE", "aud")

	cfg, err := LoadAuthConfig()
	if err != nil {
		t.Fatalf("LoadAuthConfig returned error: %v", err)
	}
	if cfg.RequirePhishingResistantMFA {
		t.Fatal("expected default phishing resistant flag when parse fails")
	}
}

func TestMain(m *testing.M) {
	// Clear any inherited environment to make tests deterministic.
	os.Clearenv()
	code := m.Run()
	os.Exit(code)
}
