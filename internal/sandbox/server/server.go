package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/sandbox/config"
	"gimpel/internal/sandbox/manager"
)

type Server struct {
	cfg *config.SandboxConfig
	mgr *manager.Manager

	grpcServer *grpc.Server
	listener   net.Listener
}

func New(cfg *config.SandboxConfig, mgr *manager.Manager) (*Server, error) {
	return &Server{
		cfg: cfg,
		mgr: mgr,
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

	handler := &Handler{mgr: s.mgr, cfg: s.cfg}
	gimpelv1.RegisterSandboxServiceServer(s.grpcServer, handler)

	log.WithField("address", s.cfg.ListenAddress).Info("sandbox server starting")

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
	log.Info("sandbox server stopped")
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

type Handler struct {
	gimpelv1.UnimplementedSandboxServiceServer
	mgr *manager.Manager
	cfg *config.SandboxConfig
}

func (h *Handler) CreateSession(ctx context.Context, req *gimpelv1.CreateSessionRequest) (*gimpelv1.CreateSessionResponse, error) {
	log.WithField("session_id", req.SessionId).Info("received create session request")

	session, err := h.mgr.CreateSession(req.SessionId, req.Image, req.Env)
	if err != nil {
		return nil, err
	}

	return &gimpelv1.CreateSessionResponse{
		Endpoint:  fmt.Sprintf("%s:%d", h.cfg.PublicIP, session.Port),
		TunnelKey: session.TunnelKey,
	}, nil
}

func (h *Handler) StopSession(ctx context.Context, req *gimpelv1.StopSessionRequest) (*gimpelv1.StopSessionResponse, error) {
	log.WithField("session_id", req.SessionId).Info("received stop session request")

	if err := h.mgr.StopSession(req.SessionId); err != nil {
		return &gimpelv1.StopSessionResponse{Success: false}, err
	}

	return &gimpelv1.StopSessionResponse{Success: true}, nil
}
