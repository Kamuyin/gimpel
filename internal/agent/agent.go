package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/agent/config"
	"gimpel/internal/agent/control"
	"gimpel/internal/agent/listener"
	"gimpel/internal/agent/module"
	"gimpel/internal/agent/modules"
	"gimpel/internal/agent/store"
	"gimpel/internal/agent/telemetry"
	gimpelv1 "gimpel/api/go/v1"
)

type Agent struct {
	cfg      *config.AgentConfig
	identity *Identity

	controlClient *control.Client
	supervisor    *module.Supervisor
	listeners     *listener.Manager
	emitter       *telemetry.Emitter

	store          *store.Store
	catalogSyncer  *modules.CatalogSyncer
	downloader     *modules.ModuleDownloader
	reconciler     *modules.Reconciler

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func New(cfg *config.AgentConfig) (*Agent, error) {
	identity, err := LoadIdentity(cfg)
	if err != nil {
		return nil, fmt.Errorf("loading identity: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	a := &Agent{
		cfg:      cfg,
		identity: identity,
		ctx:      ctx,
		cancel:   cancel,
	}

	if err := a.initComponents(); err != nil {
		cancel()
		return nil, err
	}

	return a, nil
}

func (a *Agent) initComponents() error {
	var err error

	a.store, err = store.New(&store.Config{
		DBPath:   filepath.Join(a.cfg.DataDir, "agent.db"),
		CacheDir: a.cfg.Runtime.ModuleCacheDir,
	})
	if err != nil {
		return fmt.Errorf("creating store: %w", err)
	}

	a.emitter, err = telemetry.NewEmitter(a.ctx, a.cfg, a.identity.ID)
	if err != nil {
		return fmt.Errorf("creating emitter: %w", err)
	}

	a.controlClient, err = control.NewClient(a.cfg, a.identity)
	if err != nil {
		return fmt.Errorf("creating control client: %w", err)
	}

	a.supervisor, err = module.NewSupervisor(a.cfg, a.emitter)
	if err != nil {
		return fmt.Errorf("creating supervisor: %w", err)
	}
	
	a.listeners = listener.NewManager(a.cfg, a.supervisor, a.controlClient)

	if err := a.initModuleLifecycle(); err != nil {
		return err
	}

	return nil
}

func (a *Agent) Run(ctx context.Context) error {
	log.WithFields(log.Fields{
		"agent_id": a.identity.ID,
		"hostname": a.identity.Hostname,
	}).Info("starting agent")

	if err := a.controlClient.Connect(ctx); err != nil {
		return fmt.Errorf("connecting to control plane: %w", err)
	}
	defer a.controlClient.Close()

	if !a.identity.Registered {
		if !a.cfg.PairingMode {
			return fmt.Errorf("pairing_mode is required to register a new agent")
		}
		if err := a.register(ctx); err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}
	}

	if a.catalogSyncer != nil {
		if err := a.catalogSyncer.Connect(ctx); err != nil {
			return fmt.Errorf("connecting to catalog service: %w", err)
		}
		defer a.catalogSyncer.Close()

		a.downloader = modules.NewModuleDownloader(
			a.catalogSyncer.GetCatalogClient(),
			a.store,
			a.cfg.Runtime.ModuleCacheDir,
			a.catalogSyncer.GetVerifier(),
		)
		a.reconciler = modules.NewReconciler(a.store, a.downloader, a.supervisor)
		
		a.reconciler.SetListenerStarter(a.listeners)

		log.Info("performing initial module sync")
		if err := a.syncModules(ctx); err != nil {
			log.WithError(err).Warn("initial module sync failed")
		}
	}

	if err := a.fetchConfig(ctx); err != nil {
		log.WithError(err).Warn("failed to fetch initial config, using local config")
	}

	errCh := make(chan error, 6)

	go func() {
		errCh <- a.controlClient.RunHeartbeatLoop(ctx, a.cfg.HeartbeatInterval, a.collectMetrics)
	}()

	go func() {
		errCh <- a.emitter.Run(ctx)
	}()

	go func() {
		errCh <- a.supervisor.Run(ctx)
	}()

	go func() {
		errCh <- a.listeners.Run(ctx)
	}()

	if a.catalogSyncer != nil {
		go func() {
			errCh <- a.runModuleSyncLoop(ctx)
		}()
	}

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *Agent) Shutdown(ctx context.Context) error {
	log.Info("shutting down agent")
	a.cancel()

	if a.listeners != nil {
		a.listeners.Stop()
	}

	if a.supervisor != nil {
		a.supervisor.StopAll(ctx)
	}

	if a.emitter != nil {
		a.emitter.Flush(ctx)
	}

	if a.store != nil {
		a.store.Close()
	}

	return nil
}

func (a *Agent) RunPairing(ctx context.Context) error {
	log.WithFields(log.Fields{
		"hostname": a.identity.Hostname,
		"master":   a.cfg.ControlPlane.Address,
	}).Info("starting pairing process")

	if a.identity.Registered {
		log.Warn("agent is already registered - re-pairing will replace existing credentials")
	}

	// Connect to control plane
	if err := a.controlClient.ConnectInsecure(ctx); err != nil {
		return fmt.Errorf("connecting to control plane: %w", err)
	}
	defer a.controlClient.Close()

	if err := a.register(ctx); err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	log.WithField("agent_id", a.identity.ID).Info("pairing completed successfully")
	return nil
}

func (a *Agent) register(ctx context.Context) error {
	log.Info("registering with control plane")

	resp, err := a.controlClient.Register(ctx, a.cfg.PairingToken, a.identity)
	if err != nil {
		return err
	}

	a.identity.ID = resp.AgentId

	if err := a.identity.SaveCredentials(a.cfg.DataDir, resp.Certificate, resp.PrivateKey, resp.CaCertificate); err != nil {
		return fmt.Errorf("saving credentials: %w", err)
	}

	if err := a.identity.Persist(a.cfg.DataDir); err != nil {
		return fmt.Errorf("persisting identity: %w", err)
	}

	log.WithField("agent_id", a.identity.ID).Info("registration complete")

	if a.catalogSyncer != nil {
		a.catalogSyncer.UpdateAgentID(a.identity.ID)
	}

	if a.catalogSyncer == nil {
		if err := a.initModuleLifecycle(); err != nil {
			log.WithError(err).Warn("failed to initialize module lifecycle after registration")
		}
	}

	return nil
}

func (a *Agent) initModuleLifecycle() error {
	if a.catalogSyncer != nil {
		return nil
	}

	if len(a.cfg.Runtime.TrustedKeys) == 0 {
		log.Warn("no trusted keys configured, module lifecycle disabled")
		return nil
	}

	validKeys := false
	for _, keyPath := range a.cfg.Runtime.TrustedKeys {
		if _, err := os.Stat(keyPath); err == nil {
			validKeys = true
			break
		}
	}

	if !validKeys {
		log.WithField("trusted_keys", a.cfg.Runtime.TrustedKeys).Warn("no valid trusted keys found on disk, module lifecycle disabled")
		return nil
	}

	cs, err := modules.NewCatalogSyncer(
		a.cfg,
		a.identity.ID,
		a.store,
		a.cfg.Runtime.TrustedKeys...,
	)
	if err != nil {
		return fmt.Errorf("creating catalog syncer: %w", err)
	}

	a.catalogSyncer = cs
	log.WithField("trusted_keys", len(a.cfg.Runtime.TrustedKeys)).Info("module lifecycle enabled")
	return nil
}

func (a *Agent) fetchConfig(ctx context.Context) error {
	resp, err := a.controlClient.GetConfig(ctx, "")
	if err != nil {
		return err
	}

	if !resp.Updated || resp.Config == nil {
		return nil
	}

	a.applyConfig(resp.Config)
	return nil
}

func (a *Agent) applyConfig(cfg *gimpelv1.AgentConfig) {
	if cfg == nil {
		return
	}

	log.WithFields(log.Fields{
		"version": cfg.Version,
		"modules": len(cfg.Modules),
	}).Debug("applying new configuration from control plane")

	// TODO: Apply heartbeat interval, event flush interval from config
	// TODO: Apply module deployment changes
}

func (a *Agent) collectMetrics() (cpuUsage, memUsage float64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	if m.Sys > 0 {
		memUsage = (float64(m.Alloc) / float64(m.Sys)) * 100
	}

	cpuUsage = 0

	return cpuUsage, memUsage
}

func (a *Agent) syncModules(ctx context.Context) error {
	if err := a.catalogSyncer.SyncCatalog(ctx); err != nil {
		return fmt.Errorf("syncing catalog: %w", err)
	}

	_, err := a.catalogSyncer.SyncAssignments(ctx)
	if err != nil {
		return fmt.Errorf("syncing assignments: %w", err)
	}

	if err := a.reconciler.Reconcile(ctx); err != nil {
		return fmt.Errorf("reconciling deployments: %w", err)
	}

	return nil
}

func (a *Agent) runModuleSyncLoop(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := a.syncModules(ctx); err != nil {
				log.WithError(err).Warn("module sync failed")
			}
		}
	}
}
