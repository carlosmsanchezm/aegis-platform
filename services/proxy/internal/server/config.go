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

	return Config{
		ListenAddr:       listen,
		JWTSecret:        []byte(secret),
		ExpectedAudience: aud,
		DestSuffix:       suffix,
		TokenReuseTTL:    time.Duration(ttlSeconds) * time.Second,
		Cluster:          os.Getenv("AEGIS_PROXY_CLUSTER"),
		IngressHost:      os.Getenv("AEGIS_PROXY_PUBLIC_HOST"),
		TLSCertFile:      certFile,
		TLSKeyFile:       keyFile,
		ClientCertSanSuffixAllowList: sanSuffixAllowList,
	}, nil
}

//
// Configuration options:
//   - AEGIS_PROXY_CLIENT_CERT_SAN_SUFFIX: Comma-separated list of allowed SAN suffixes for client certificates. If set, only clients with at least one SAN ending with an allowed suffix may connect. Example: ".trusted.example.com,.corp.local"
//
