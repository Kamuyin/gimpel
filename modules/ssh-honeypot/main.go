package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	gimpelsdk "github.com/NoHaxxJustLags/gimpel/sdk/go"
)

func main() {
	port := os.Getenv("SSH_PORT")
	if port == "" {
		port = "2222"
	}

	standaloneMode := os.Getenv("GIMPEL_SOCKET") == ""

	module := NewSSHHoneypot()

	if standaloneMode {
		log.Printf("Running in STANDALONE mode on port %s", port)
		runStandalone(module, port)
	} else {
		log.Printf("Running in AGENT mode")
		server := gimpelsdk.NewServer(module)
		if err := server.Run(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

func runStandalone(module *SSHHoneypot, port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}
	defer listener.Close()

	log.Printf("SSH honeypot listening on :%s", port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		info := &gimpelsdk.ConnectionInfo{
			ConnectionID: "standalone",
			SourceIP:     conn.RemoteAddr().(*net.TCPAddr).IP.String(),
			SourcePort:   uint32(conn.RemoteAddr().(*net.TCPAddr).Port),
			DestPort:     uint32(2222),
			Protocol:     "ssh",
		}

		go module.HandleConnection(ctx, conn, info)
	}
}
