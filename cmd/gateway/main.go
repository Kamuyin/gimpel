package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/gateway/config"
	"gimpel/internal/gateway/server"
)

func main() {
	configFile := flag.String("config", "gateway.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Warnf("invalid log level '%s', defaulting to info", cfg.LogLevel)
		level = log.InfoLevel
	}
	log.SetLevel(level)
	log.SetFormatter(&log.JSONFormatter{})

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Info("shutting down gateway...")
	srv.Stop()
}
