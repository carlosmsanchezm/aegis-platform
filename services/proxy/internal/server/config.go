package server

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddr       string
	JWTSecret        []byte
	ExpectedAudience string
	DestSuffix       string
	TokenReuseTTL    time.Duration
	Cluster          string
	IngressHost      string
	TLSCertFile      string
	TLSKeyFile       string
	// ClientCertSanSuffixAllowList is a list of allowed SAN suffixes for client certificates.
	// If non-empty, at least one SAN in the client certificate must end with one of these suffixes.
	// Controlled by the AEGIS_PROXY_CLIENT_CERT_SAN_SUFFIX environment variable (comma-separated).
	ClientCertSanSuffixAllowList []string
	// PrivilegedSessionTimeout is the inactivity timeout for privileged user sessions (SC-10).
	PrivilegedSessionTimeout time.Duration
	// StandardSessionTimeout is the inactivity timeout for standard user sessions (SC-10).
	StandardSessionTimeout time.Duration
}

func LoadConfig() (Config, error) {
	secret := os.Getenv("AEGIS_PROXY_JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("AEGIS_PROXY_JWT_SECRET must be set")
	}
	listen := os.Getenv("AEGIS_PROXY_LISTEN")
	if listen == "" {
		listen = ":8080"
	}
	aud := os.Getenv("AEGIS_PROXY_EXPECTED_AUDIENCE")
	if aud == "" {
		aud = "aegis-proxy"
	}
	suffix := os.Getenv("AEGIS_PROXY_ALLOWED_SUFFIX")
	if suffix == "" {
		suffix = ".svc.cluster.local"
	}
	ttlSeconds := 600
	if txt := os.Getenv("AEGIS_PROXY_JTI_TTL_SECONDS"); txt != "" {
		if val, err := strconv.Atoi(txt); err == nil && val > 0 {
			ttlSeconds = val
		}
	}
	certFile := os.Getenv("AEGIS_PROXY_TLS_CERT")
	keyFile := os.Getenv("AEGIS_PROXY_TLS_KEY")
	if certFile == "" || keyFile == "" {
		return Config{}, fmt.Errorf("AEGIS_PROXY_TLS_CERT and AEGIS_PROXY_TLS_KEY must be set")
	}

	// Load client cert SAN suffix allow-list from env var
	var sanSuffixAllowList []string
	if v := os.Getenv("AEGIS_PROXY_CLIENT_CERT_SAN_SUFFIX"); v != "" {
		for _, s := range strings.Split(v, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				sanSuffixAllowList = append(sanSuffixAllowList, s)
			}
		}
	}

	// Load session inactivity timeouts (SC-10 Network Disconnect)
	privilegedTimeout := 10 * time.Minute
	if txt := os.Getenv("AEGIS_PROXY_PRIVILEGED_TIMEOUT"); txt != "" {
		if d, err := time.ParseDuration(txt); err == nil && d > 0 {
			privilegedTimeout = d
		} else {
			return Config{}, fmt.Errorf("invalid AEGIS_PROXY_PRIVILEGED_TIMEOUT: %v", err)
		}
	}

	standardTimeout := 15 * time.Minute
	if txt := os.Getenv("AEGIS_PROXY_STANDARD_TIMEOUT"); txt != "" {
		if d, err := time.ParseDuration(txt); err == nil && d > 0 {
			standardTimeout = d
		} else {
			return Config{}, fmt.Errorf("invalid AEGIS_PROXY_STANDARD_TIMEOUT: %v", err)
		}
	}

	// Validate that timeouts are positive
	if privilegedTimeout <= 0 {
		return Config{}, fmt.Errorf("AEGIS_PROXY_PRIVILEGED_TIMEOUT must be positive")
	}
	if standardTimeout <= 0 {
		return Config{}, fmt.Errorf("AEGIS_PROXY_STANDARD_TIMEOUT must be positive")
	}

	return Config{
		ListenAddr:                   listen,
		JWTSecret:                    []byte(secret),
		ExpectedAudience:             aud,
		DestSuffix:                   suffix,
		TokenReuseTTL:                time.Duration(ttlSeconds) * time.Second,
		Cluster:                      os.Getenv("AEGIS_PROXY_CLUSTER"),
		IngressHost:                  os.Getenv("AEGIS_PROXY_PUBLIC_HOST"),
		TLSCertFile:                  certFile,
		TLSKeyFile:                   keyFile,
		ClientCertSanSuffixAllowList: sanSuffixAllowList,
		PrivilegedSessionTimeout:     privilegedTimeout,
		StandardSessionTimeout:       standardTimeout,
	}, nil
}

//
// Configuration options:
//   - AEGIS_PROXY_CLIENT_CERT_SAN_SUFFIX: Comma-separated list of allowed SAN suffixes for client certificates. If set, only clients with at least one SAN ending with an allowed suffix may connect. Example: ".trusted.example.com,.corp.local"
//   - AEGIS_PROXY_PRIVILEGED_TIMEOUT: Inactivity timeout for privileged user sessions (default: 10m). Enforces SC-10 Network Disconnect requirements for elevated-privilege sessions.
//   - AEGIS_PROXY_STANDARD_TIMEOUT: Inactivity timeout for standard user sessions (default: 15m). Enforces SC-10 Network Disconnect requirements for standard sessions.
//
