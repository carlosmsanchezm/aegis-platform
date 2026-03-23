package main

import (
	"log"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/proxy/internal/server"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("unable to create logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	// FIPS 140-2 startup verification (SC-13)
	checkFIPSMode(logger)

	cfg, err := server.LoadConfig()
	if err != nil {
		logger.Fatal("invalid configuration", zap.Error(err))
	}

	proxy := server.NewProxyServer(logger, cfg)
	if err := proxy.Start(); err != nil {
		logger.Fatal("proxy exited", zap.Error(err))
	}
}

// checkFIPSMode verifies BoringCrypto is active at startup (SC-13 FIPS 140-2).
func checkFIPSMode(logger *zap.Logger) {
	boringActive := strings.Contains(runtime.Version(), "boringcrypto")
	if boringActive {
		logger.Info("FIPS mode: enabled (BoringCrypto)", zap.String("go_version", runtime.Version()))
	} else {
		logger.Info("FIPS mode: disabled (standard Go crypto)", zap.String("go_version", runtime.Version()))
	}
	if os.Getenv("AEGIS_FIPS_ENABLED") == "true" && !boringActive {
		logger.Fatal("AEGIS_FIPS_ENABLED=true but BoringCrypto is not active; binary must be built with GOEXPERIMENT=boringcrypto")
	}
}
