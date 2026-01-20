package modules

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/agent/config"
	"gimpel/internal/agent/control"
	"gimpel/internal/agent/store"
	"gimpel/pkg/signing"
)

type CatalogSyncer struct {
	cfg      *config.AgentConfig
	agentID  string
	store    Store
	verifier *signing.ModuleVerifier

	mu             sync.RWMutex
	conn           *grpc.ClientConn
	catalogClient  gimpelv1.ModuleCatalogServiceClient
	catalogVersion int64
	configVersion  int64
}

func (cs *CatalogSyncer) GetCatalogClient() gimpelv1.ModuleCatalogServiceClient {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.catalogClient
}

func (cs *CatalogSyncer) GetVerifier() *signing.ModuleVerifier {
	return cs.verifier
}

func (cs *CatalogSyncer) UpdateAgentID(agentID string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.agentID = agentID
}

type Store interface {
	GetAgentState() (*store.AgentState, error)
	SaveAgentState(*store.AgentState) error
	GetModuleCache(moduleID, version string) (*store.ModuleCache, error)
	SaveModuleCache(*store.ModuleCache) error
	GetDeploymentConfig() (*store.DeploymentConfig, error)
	SaveDeploymentConfig(*store.DeploymentConfig) error
}

func NewCatalogSyncer(cfg *config.AgentConfig, agentID string, s Store, trustedKeyPath string) (*CatalogSyncer, error) {
	keyPair, err := signing.LoadPublicKey(trustedKeyPath)
	if err != nil {
		return nil, fmt.Errorf("loading trusted key: %w", err)
	}

	verifier := signing.NewModuleVerifier(keyPair)

	cs := &CatalogSyncer{
		cfg:      cfg,
		agentID:  agentID,
		store:    s,
		verifier: verifier,
	}

	state, err := s.GetAgentState()
	if err == nil && state != nil {
		cs.catalogVersion = state.CatalogVersion
		cs.configVersion = state.ConfigVersion
	}

	return cs, nil
}

func (cs *CatalogSyncer) Connect(ctx context.Context) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.conn != nil {
		return nil
	}

	var opts []grpc.DialOption

	tlsCfg := cs.cfg.ControlPlane.TLS
	if tlsCfg.CertFile != "" && tlsCfg.KeyFile != "" {
		creds, err := control.LoadClientCredentials(tlsCfg.CertFile, tlsCfg.KeyFile, tlsCfg.CAFile)
		if err != nil {
			return fmt.Errorf("loading TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cs.cfg.ControlPlane.Address, opts...)
	if err != nil {
		return fmt.Errorf("dialing master: %w", err)
	}

	cs.conn = conn
	cs.catalogClient = gimpelv1.NewModuleCatalogServiceClient(conn)

	log.Info("connected to module catalog service")
	return nil
}

func (cs *CatalogSyncer) Close() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.conn != nil {
		return cs.conn.Close()
	}
	return nil
}

func (cs *CatalogSyncer) SyncCatalog(ctx context.Context) error {
	cs.mu.RLock()
	client := cs.catalogClient
	currentVersion := cs.catalogVersion
	cs.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("not connected")
	}

	resp, err := client.GetCatalog(ctx, &gimpelv1.GetCatalogRequest{
		CurrentVersion: currentVersion,
	})
	if err != nil {
		return fmt.Errorf("fetching catalog: %w", err)
	}

	if !resp.Updated {
		log.Debug("catalog is up to date")
		return nil
	}

	catalog := resp.Catalog
	if catalog == nil {
		return fmt.Errorf("empty catalog response")
	}

	if catalog.Signature != nil && catalog.SignedBy != "" {
		if err := cs.verifier.VerifyCatalog(catalog); err != nil {
			return fmt.Errorf("catalog signature verification failed: %w", err)
		}
	} else {
		log.Warn("catalog is unsigned; skipping catalog signature verification")
	}

	log.WithFields(log.Fields{
		"version":      catalog.Version,
		"module_count": len(catalog.Modules),
		"signed_by":    catalog.SignedBy,
	}).Info("catalog updated and verified")

	cs.mu.Lock()
	cs.catalogVersion = catalog.Version
	cs.mu.Unlock()

	if err := cs.store.SaveAgentState(&store.AgentState{
		AgentID:        cs.agentID,
		CatalogVersion: catalog.Version,
		ConfigVersion:  cs.configVersion,
	}); err != nil {
		log.WithError(err).Warn("failed to save agent state")
	}

	return nil
}

func (cs *CatalogSyncer) SyncAssignments(ctx context.Context) (*store.DeploymentConfig, error) {
	cs.mu.RLock()
	client := cs.catalogClient
	currentVersion := cs.configVersion
	cs.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("not connected")
	}

	resp, err := client.GetModuleAssignments(ctx, &gimpelv1.GetModuleAssignmentsRequest{
		AgentId:        cs.agentID,
		CurrentVersion: currentVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching assignments: %w", err)
	}

	if !resp.Updated {
		log.Debug("assignments are up to date")
		return nil, nil
	}

	agentConfig := resp.Config
	if agentConfig == nil {
		return nil, fmt.Errorf("empty assignments response")
	}

	if agentConfig.Signature != nil && len(agentConfig.Signature) > 0 {
		if err := cs.verifier.VerifyAgentConfig(agentConfig); err != nil {
			return nil, fmt.Errorf("assignment signature verification failed: %w", err)
		}
	} else {
		log.Warn("assignments are unsigned; skipping assignment signature verification")
	}

	deployment := &store.DeploymentConfig{
		Version:    agentConfig.Version,
		Signature:  agentConfig.Signature,
		ReceivedAt: time.Now(),
		Modules:    make([]store.ModuleDeployment, 0, len(agentConfig.Assignments)),
	}

	for _, assignment := range agentConfig.Assignments {
		listeners := make([]store.ListenerConfig, 0, len(assignment.Listeners))
		for _, l := range assignment.Listeners {
			listeners = append(listeners, store.ListenerConfig{
				ID:              l.Id,
				Protocol:        l.Protocol,
				Port:            l.Port,
				HighInteraction: l.HighInteraction,
			})
		}

		deployment.Modules = append(deployment.Modules, store.ModuleDeployment{
			ModuleID:      assignment.ModuleId,
			ModuleVersion: assignment.Version,
			Enabled:       true,
			ExecutionMode: assignment.ExecutionMode,
			Listeners:     listeners,
			Env:           assignment.Env,
		})
	}

	log.WithFields(log.Fields{
		"version":     deployment.Version,
		"assignments": len(deployment.Modules),
	}).Info("assignments updated and verified")

	cs.mu.Lock()
	cs.configVersion = deployment.Version
	cs.mu.Unlock()

	if err := cs.store.SaveDeploymentConfig(deployment); err != nil {
		return nil, fmt.Errorf("saving deployment config: %w", err)
	}

	if err := cs.store.SaveAgentState(&store.AgentState{
		AgentID:        cs.agentID,
		CatalogVersion: cs.catalogVersion,
		ConfigVersion:  deployment.Version,
	}); err != nil {
		log.WithError(err).Warn("failed to save agent state")
	}

	return deployment, nil
}

func (cs *CatalogSyncer) CurrentVersions() (catalogVersion, configVersion int64) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.catalogVersion, cs.configVersion
}
