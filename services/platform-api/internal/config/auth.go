package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultJWKSCacheTTL        = 5 * time.Minute
	defaultJWKSRefreshInterval = 1 * time.Minute
	defaultRoleBindingsJSON    = `[
  {
    "clients": ["backstage", "vscode-extension"],
    "roles": ["workspace-admin"],
    "projects": ["*"],
    "queues": ["*"]
  }
]`
)

// AuthConfig captures the runtime configuration for verifying OIDC access tokens.
type AuthConfig struct {
	IssuerURL                   string
	Audience                    string
	JWKSURL                     string
	JWKSCacheTTL                time.Duration
	JWKSRefreshInterval         time.Duration
	RequirePhishingResistantMFA bool
	AllowedPhishingResistantAMR []string
	DefaultRoleBindingsJSON     string
}

type rawAuthConfig struct {
	IssuerURL                   string   `json:"issuerUrl"`
	Audience                    string   `json:"audience"`
	JWKSURL                     string   `json:"jwksUrl"`
	JWKSCacheTTL                string   `json:"jwksCacheTtl"`
	JWKSRefreshInterval         string   `json:"jwksRefreshInterval"`
	RequirePhishingResistantMFA *bool    `json:"requirePhishingResistantMfa"`
	AllowedPhishingResistantAMR []string `json:"allowedPhishingResistantAmr"`
}

// LoadAuthConfig returns the effective authentication configuration derived from
// environment variables with an optional JSON overlay supplied via AUTH_CONFIG_JSON.
func LoadAuthConfig() (*AuthConfig, error) {
	cfg := &AuthConfig{
		JWKSCacheTTL:                defaultJWKSCacheTTL,
		JWKSRefreshInterval:         defaultJWKSRefreshInterval,
		AllowedPhishingResistantAMR: []string{"hwk", "webauthn", "piv", "piv-cac"},
		DefaultRoleBindingsJSON:     defaultRoleBindingsJSON,
	}

	if raw := strings.TrimSpace(os.Getenv("AUTH_CONFIG_JSON")); raw != "" {
		var parsed rawAuthConfig
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return nil, fmt.Errorf("failed to parse AUTH_CONFIG_JSON: %w", err)
		}
		applyRawConfig(cfg, &parsed)
	}

	overlayFromEnv(cfg)

	if cfg.IssuerURL == "" {
		return nil, fmt.Errorf("OIDC issuer not configured (set OIDC_ISSUER_URL)")
	}
	if cfg.Audience == "" {
		return nil, fmt.Errorf("OIDC audience not configured (set OIDC_AUDIENCE)")
	}

	cfg.IssuerURL = strings.TrimRight(cfg.IssuerURL, "/")
	if cfg.JWKSURL == "" {
		cfg.JWKSURL = fmt.Sprintf("%s/protocol/openid-connect/certs", cfg.IssuerURL)
	}

	if cfg.JWKSCacheTTL <= 0 {
		cfg.JWKSCacheTTL = defaultJWKSCacheTTL
	}
	if cfg.JWKSRefreshInterval <= 0 {
		cfg.JWKSRefreshInterval = defaultJWKSRefreshInterval
	}

	if len(cfg.AllowedPhishingResistantAMR) == 0 {
		cfg.AllowedPhishingResistantAMR = []string{"hwk", "webauthn", "piv", "piv-cac"}
	} else {
		cfg.AllowedPhishingResistantAMR = normalizeStrings(cfg.AllowedPhishingResistantAMR)
	}

	return cfg, nil
}

// DefaultRoleBindingsJSON returns the built-in authorization bindings used when AUTHZ_ROLE_BINDINGS_JSON is unset.
func DefaultRoleBindingsJSON() string {
	return defaultRoleBindingsJSON
}

func applyRawConfig(cfg *AuthConfig, raw *rawAuthConfig) {
	if raw == nil {
		return
	}
	if raw.IssuerURL != "" {
		cfg.IssuerURL = raw.IssuerURL
	}
	if raw.Audience != "" {
		cfg.Audience = raw.Audience
	}
	if raw.JWKSURL != "" {
		cfg.JWKSURL = raw.JWKSURL
	}
	if raw.JWKSCacheTTL != "" {
		if ttl, err := time.ParseDuration(strings.TrimSpace(raw.JWKSCacheTTL)); err == nil {
			cfg.JWKSCacheTTL = ttl
		}
	}
	if raw.JWKSRefreshInterval != "" {
		if interval, err := time.ParseDuration(strings.TrimSpace(raw.JWKSRefreshInterval)); err == nil {
			cfg.JWKSRefreshInterval = interval
		}
	}
	if raw.RequirePhishingResistantMFA != nil {
		cfg.RequirePhishingResistantMFA = *raw.RequirePhishingResistantMFA
	}
	if len(raw.AllowedPhishingResistantAMR) > 0 {
		cfg.AllowedPhishingResistantAMR = normalizeStrings(raw.AllowedPhishingResistantAMR)
	}
}

func overlayFromEnv(cfg *AuthConfig) {
	if v := strings.TrimSpace(os.Getenv("OIDC_ISSUER_URL")); v != "" {
		cfg.IssuerURL = v
	}
	if v := strings.TrimSpace(os.Getenv("OIDC_AUDIENCE")); v != "" {
		cfg.Audience = v
	}
	if v := strings.TrimSpace(os.Getenv("OIDC_JWKS_URL")); v != "" {
		cfg.JWKSURL = v
	}
	if v := strings.TrimSpace(os.Getenv("OIDC_JWKS_CACHE_TTL")); v != "" {
		if ttl, err := time.ParseDuration(v); err == nil {
			cfg.JWKSCacheTTL = ttl
		}
	}
	if v := strings.TrimSpace(os.Getenv("OIDC_JWKS_REFRESH_INTERVAL")); v != "" {
		if interval, err := time.ParseDuration(v); err == nil {
			cfg.JWKSRefreshInterval = interval
		}
	}
	if v := strings.TrimSpace(os.Getenv("REQUIRE_PHISHING_RESISTANT_MFA")); v != "" {
		if parsed, err := parseBool(v); err == nil {
			cfg.RequirePhishingResistantMFA = parsed
		}
	}
	if v := strings.TrimSpace(os.Getenv("ALLOWED_PHISHING_RESISTANT_AMR")); v != "" {
		parts := strings.Split(v, ",")
		cfg.AllowedPhishingResistantAMR = normalizeStrings(parts)
	}
}

func parseBool(val string) (bool, error) {
	val = strings.TrimSpace(strings.ToLower(val))
	switch val {
	case "1", "t", "true", "y", "yes":
		return true, nil
	case "0", "f", "false", "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", val)
	}
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, v := range values {
		trimmed := strings.TrimSpace(strings.ToLower(v))
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
