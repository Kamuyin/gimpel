package agent

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/agent/config"
	"gimpel/internal/agent/control"
	"gimpel/internal/agent/listener"
	"gimpel/internal/agent/module"
	"gimpel/internal/agent/telemetry"
)

type Agent struct {
	cfg      *config.AgentConfig
	identity *Identity

	controlClient *control.Client
	supervisor    *module.Supervisor
	listeners     *listener.Manager
	emitter       *telemetry.Emitter

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

	a.emitter, err = telemetry.NewEmitter(a.ctx, a.cfg, a.identity.ID)
	if err != nil {
		return fmt.Errorf("creating emitter: %w", err)
	}

	a.controlClient, err = control.NewClient(a.cfg, a.identity)
	if err != nil {
		return fmt.Errorf("creating control client: %w", err)
	}

	a.supervisor = module.NewSupervisor(a.cfg, a.emitter)
	a.listeners = listener.NewManager(a.cfg, a.supervisor, a.controlClient)

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
		if err := a.register(ctx); err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}
	}

	if err := a.fetchConfig(ctx); err != nil {
		log.WithError(err).Warn("failed to fetch initial config, using local config")
	}

	errCh := make(chan error, 4)

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

	return nil
}

func (a *Agent) register(ctx context.Context) error {
	log.Info("registering with control plane")

	resp, err := a.controlClient.Register(ctx, a.cfg.RegistrationToken, a.identity)
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

func (a *Agent) applyConfig(cfg interface{}) {
	log.Debug("applying new configuration from control plane")
}

func (a *Agent) collectMetrics() (cpuUsage, memUsage float64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memUsage = float64(m.Alloc) / float64(m.Sys)
	return 0, memUsage
}
