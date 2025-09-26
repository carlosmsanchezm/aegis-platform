package server

import (
	"fmt"
	"os"
	"strconv"
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
	}, nil
}
