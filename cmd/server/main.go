// Command server is the entrypoint for the rakepro-web HTTP server.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/rakepro/rakepro-web/internal/config"
	"github.com/rakepro/rakepro-web/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// Logger not built yet; use the zerolog default to surface the error.
		log.Fatal().Err(err).Msg("load config")
	}

	logger := server.NewLogger(cfg.Env, cfg.LogLevel)

	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("init server")
	}

	// Cancel the context on SIGINT/SIGTERM to trigger graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := srv.Run(ctx); err != nil {
		logger.Fatal().Err(err).Msg("server error")
		os.Exit(1)
	}
}
