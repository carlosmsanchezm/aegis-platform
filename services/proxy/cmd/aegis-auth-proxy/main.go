package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/proxy/internal/server"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("unable to create logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	cfg, err := server.LoadConfig()
	if err != nil {
		logger.Fatal("invalid configuration", zap.Error(err))
	}

	proxy := server.NewProxyServer(logger, cfg)
	if err := proxy.Start(); err != nil {
		logger.Fatal("proxy exited", zap.Error(err))
	}
}
