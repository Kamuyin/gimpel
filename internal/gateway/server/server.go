package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/gateway/config"
	"gimpel/internal/gateway/ingest"
)

type Server struct {
	cfg *config.GatewayConfig

	grpcServer *grpc.Server
	listener   net.Listener
}

func New(cfg *config.GatewayConfig) (*Server, error) {
	return &Server{
		cfg: cfg,
	}, nil
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.cfg.ListenAddress)
	if err != nil {
		return fmt.Errorf("binding to %s: %w", s.cfg.ListenAddress, err)
	}
	s.listener = ln

	opts, err := s.buildServerOptions()
	if err != nil {
		return fmt.Errorf("building server options: %w", err)
	}

	s.grpcServer = grpc.NewServer(opts...)

	handler := ingest.NewHandler()
	gimpelv1.RegisterIngestionServiceServer(s.grpcServer, handler)

	log.WithField("address", s.cfg.ListenAddress).Info("gateway server starting")

	go func() {
		if err := s.grpcServer.Serve(ln); err != nil {
			log.WithError(err).Error("server error")
		}
	}()

	return nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	log.Info("gateway server stopped")
}

func (s *Server) buildServerOptions() ([]grpc.ServerOption, error) {
	var opts []grpc.ServerOption

	if s.cfg.TLS.CertFile != "" && s.cfg.TLS.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading TLS cert: %w", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.VerifyClientCertIfGiven,
		}

		if s.cfg.TLS.CAFile != "" {
			caCert, err := os.ReadFile(s.cfg.TLS.CAFile)
			if err != nil {
				return nil, fmt.Errorf("reading CA cert: %w", err)
			}
			caPool := x509.NewCertPool()
			if !caPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse CA cert")
			}
			tlsConfig.ClientCAs = caPool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	return opts, nil
}
