package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/agent"
	"gimpel/internal/agent/config"
)

func main() {
	configPath := flag.String("config", "/etc/gimpel/agent.yaml", "path to config file")
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}

	a, err := agent.New(cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to create agent")
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.WithField("signal", sig).Info("received shutdown signal")
		cancel()
	}()

	if err := a.Run(ctx); err != nil && err != context.Canceled {
		log.WithError(err).Error("agent run failed")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30)
	defer shutdownCancel()

	if err := a.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("shutdown error")
	}

	log.Info("agent stopped")
}
