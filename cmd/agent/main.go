package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"gimpel/internal/agent"
	"gimpel/internal/agent/config"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	lvl, err := logrus.ParseLevel(cfg.LogLevel)
	if err == nil {
		log.SetLevel(lvl)
	} else {
		log.Warnf("Invalid log level '%s', defaulting to info", cfg.LogLevel)
	}

	a, err := agent.New(cfg, log)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := a.Run(ctx); err != nil {
		if err != context.Canceled {
			log.Errorf("Agent stopped with error: %v", err)
			os.Exit(1)
		}
	}
	log.Info("Agent shutdown complete")
}
