package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/master/config"
	"gimpel/internal/master/registry"
	"gimpel/internal/master/server"
)

func main() {
	configPath := flag.String("config", "/etc/gimpel/master.yaml", "path to config file")
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

	srv, err := server.New(cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to create server")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if inMemReg, ok := srv.Registry.(*registry.InMemoryRegistry); ok {
		go inMemReg.RunHealthChecker(ctx)
	}

	if err := srv.Start(); err != nil {
		log.WithError(err).Fatal("failed to start server")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.WithField("signal", sig).Info("received shutdown signal")

	cancel()
	srv.Stop()

	log.Info("master stopped")
}
