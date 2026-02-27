package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/master/ca"
	"gimpel/internal/master/config"
	"gimpel/internal/master/session"
	"gimpel/internal/master/store"
)

type Server struct {
	cfg *config.MasterConfig

	grpcServer *grpc.Server
	listener   net.Listener

	Store      *store.Store
	CA         *ca.CA
	SessionMgr *session.SessionManager
}

func New(cfg *config.MasterConfig) (*Server, error) {
	caInstance, err := ca.New(&cfg.CA)
	if err != nil {
		return nil, fmt.Errorf("initializing CA: %w", err)
	}

	masterStore, err := store.New(&store.Config{
		DBPath:   filepath.Join(cfg.DataDir, "master.db"),
		ImageDir: filepath.Join(cfg.DataDir, "images"),
	})
	if err != nil {
		return nil, fmt.Errorf("initializing store: %w", err)
	}

	s := &Server{
		cfg:        cfg,
		CA:         caInstance,
		SessionMgr: session.NewSessionManager(&cfg.Sandbox),
		Store:      masterStore,
	}

	return s, nil
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

	handler := NewHandler(s.cfg, s.Store, s.CA, s.SessionMgr)
	gimpelv1.RegisterAgentControlServer(s.grpcServer, handler)

	catalogHandler := NewModuleCatalogHandler(s.Store)
	gimpelv1.RegisterModuleCatalogServiceServer(s.grpcServer, catalogHandler)

	log.WithField("address", s.cfg.ListenAddress).Info("master server starting")

	go func() {
		if err := s.grpcServer.Serve(ln); err != nil {
			log.WithError(err).Error("server error")
		}
	}()

	restAddr := ":8080"
	if s.cfg.RESTAddress != "" {
		restAddr = s.cfg.RESTAddress
	}
	if err := s.StartRESTServer(restAddr); err != nil {
		return fmt.Errorf("starting REST server: %w", err)
	}

	return nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.Store != nil {
		s.Store.Close()
	}
	log.Info("master server stopped")
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
