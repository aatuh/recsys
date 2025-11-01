package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"recsys/internal/app"
	"recsys/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(ctx, config.EnvSource{})
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application, err := app.New(ctx, app.Options{Config: cfg})
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		if logger := application.Logger(); logger != nil {
			logger.Fatal("application error", zap.Error(err))
		}
		log.Fatalf("application error: %v", err)
	}
}
