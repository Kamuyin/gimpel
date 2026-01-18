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
	"gimpel/internal/master/ca"
	"gimpel/internal/master/config"
	"gimpel/internal/master/configstore"
	"gimpel/internal/master/modulestore"
	"gimpel/internal/master/registry"
	"gimpel/internal/master/session"
	"gimpel/pkg/signing"
)

type Server struct {
	cfg *config.MasterConfig

	grpcServer *grpc.Server
	listener   net.Listener

	Registry    registry.Registry
	CA          *ca.CA
	ConfigStore configstore.ConfigStore
	SessionMgr  *session.SessionManager
	ModuleStore modulestore.Store
}

func New(cfg *config.MasterConfig) (*Server, error) {
	caInstance, err := ca.New(&cfg.CA)
	if err != nil {
		return nil, fmt.Errorf("initializing CA: %w", err)
	}

	var moduleSigner *signing.ModuleSigner
	if cfg.ModuleStore.SigningKeyFile != "" {
		keyPair, err := signing.LoadPrivateKey(cfg.ModuleStore.SigningKeyFile)
		if err != nil {
			if os.IsNotExist(err) && cfg.ModuleStore.AutoSign {
				log.Info("generating new module signing key pair")
				keyPair, err = signing.GenerateKeyPair()
				if err != nil {
					return nil, fmt.Errorf("generating signing key: %w", err)
				}
				if err := keyPair.SavePrivateKey(cfg.ModuleStore.SigningKeyFile); err != nil {
					return nil, fmt.Errorf("saving private key: %w", err)
				}
				if err := keyPair.SavePublicKey(cfg.ModuleStore.PublicKeyFile); err != nil {
					return nil, fmt.Errorf("saving public key: %w", err)
				}
				log.WithField("key_id", keyPair.KeyID).Info("signing key pair generated")
			} else if !os.IsNotExist(err) {
				return nil, fmt.Errorf("loading signing key: %w", err)
			}
		}

		if keyPair != nil {
			moduleSigner, err = signing.NewModuleSigner(keyPair)
			if err != nil {
				return nil, fmt.Errorf("creating module signer: %w", err)
			}
			log.WithField("key_id", keyPair.KeyID).Info("module signing enabled")
		}
	}

	moduleStore, err := modulestore.NewFileStore(cfg.ModuleStore.DataDir, moduleSigner)
	if err != nil {
		return nil, fmt.Errorf("initializing module store: %w", err)
	}

	s := &Server{
		cfg:         cfg,
		CA:          caInstance,
		Registry:    registry.NewInMemoryRegistry(&cfg.Registry),
		ConfigStore: configstore.NewInMemoryConfigStore(),
		SessionMgr:  session.NewSessionManager(&cfg.Sandbox),
		ModuleStore: moduleStore,
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

	handler := NewHandler(s.cfg, s.Registry, s.CA, s.ConfigStore, s.SessionMgr)
	gimpelv1.RegisterAgentControlServer(s.grpcServer, handler)

	catalogHandler := NewModuleCatalogHandler(s.ModuleStore)
	gimpelv1.RegisterModuleCatalogServiceServer(s.grpcServer, catalogHandler)

	log.WithField("address", s.cfg.ListenAddress).Info("master server starting")

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
