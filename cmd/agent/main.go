package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/agent"
	"gimpel/internal/agent/config"
)

func main() {
	configPath := flag.String("config", "/etc/gimpel/agent.yaml", "path to config file")
	debug := flag.Bool("debug", false, "enable debug logging")
	
	// Pairing mode flags
	pairMode := flag.Bool("pair", false, "enter pairing mode to register with master")
	pairToken := flag.String("pair-token", "", "pairing token (if not provided, will prompt interactively)")
	
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

	// Handle pairing mode
	if *pairMode {
		cfg.PairingMode = true
		if *pairToken != "" {
			cfg.PairingToken = *pairToken
		} else {
			token, err := promptForToken()
			if err != nil {
				log.WithError(err).Fatal("failed to read pairing token")
			}
			cfg.PairingToken = token
		}
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

	if cfg.PairingMode {
		if err := a.RunPairing(ctx); err != nil {
			log.WithError(err).Fatal("pairing failed")
		}
		printPairingSuccess()
		return
	}

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

func printPairingSuccess() {
	fmt.Println("pairing succeeded")
	fmt.Println("agent credentials saved")
}

// promptForToken prompts the user to enter the pairing token interactively
func promptForToken() (string, error) {
	fmt.Println("pairing mode")
	fmt.Print("pairing code: ")

	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Normalize: trim whitespace, remove dashes, uppercase
	token = strings.TrimSpace(token)
	token = strings.ToUpper(strings.ReplaceAll(token, "-", ""))

	if len(token) != 8 {
		return "", fmt.Errorf("invalid pairing code format (expected 8 characters, got %d)", len(token))
	}

	fmt.Println()
	return token, nil
}
