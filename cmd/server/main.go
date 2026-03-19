package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"remote-server/config"
	"remote-server/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	cfg.ServerAddr = "0.0.0.0:8080"

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := server.New(ctx, cfg, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize application")
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("application stopped with error")
	}

	log.Info().Msg("application shutdown complete")
}
