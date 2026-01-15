package module

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/agent/config"
	"gimpel/internal/agent/telemetry"
)

type ModuleState int

const (
	ModuleStateStopped ModuleState = iota
	ModuleStateStarting
	ModuleStateRunning
	ModuleStateFailed
)

type ManagedModule struct {
	Config  config.ModuleConfig
	State   ModuleState
	Process *Process
	Client  *Client
	LastErr error
}

type Supervisor struct {
	cfg     *config.AgentConfig
	emitter *telemetry.Emitter

	mu      sync.RWMutex
	modules map[string]*ManagedModule

	healthInterval time.Duration
}

func NewSupervisor(cfg *config.AgentConfig, emitter *telemetry.Emitter) *Supervisor {
	return &Supervisor{
		cfg:            cfg,
		emitter:        emitter,
		modules:        make(map[string]*ManagedModule),
		healthInterval: 10 * time.Second,
	}
}

func (s *Supervisor) Run(ctx context.Context) error {
	for _, modCfg := range s.cfg.Modules {
		if err := s.StartModule(ctx, modCfg); err != nil {
			log.WithError(err).WithField("module", modCfg.ID).Error("failed to start module")
		}
	}

	ticker := time.NewTicker(s.healthInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.checkHealth(ctx)
		}
	}
}

func (s *Supervisor) StartModule(ctx context.Context, cfg config.ModuleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.modules[cfg.ID]; ok && existing.State == ModuleStateRunning {
		return nil
	}

	mod := &ManagedModule{
		Config: cfg,
		State:  ModuleStateStarting,
	}
	s.modules[cfg.ID] = mod

	proc, err := StartProcess(ctx, cfg)
	if err != nil {
		mod.State = ModuleStateFailed
		mod.LastErr = err
		return fmt.Errorf("starting process: %w", err)
	}
	mod.Process = proc

	client, err := NewClient(cfg.SocketPath)
	if err != nil {
		proc.Stop()
		mod.State = ModuleStateFailed
		mod.LastErr = err
		return fmt.Errorf("creating client: %w", err)
	}
	mod.Client = client

	mod.State = ModuleStateRunning
	log.WithField("module", cfg.ID).Info("module started")
	return nil
}

func (s *Supervisor) StopModule(ctx context.Context, moduleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mod, ok := s.modules[moduleID]
	if !ok {
		return nil
	}

	if mod.Client != nil {
		mod.Client.Close()
	}

	if mod.Process != nil {
		mod.Process.Stop()
	}

	mod.State = ModuleStateStopped
	log.WithField("module", moduleID).Info("module stopped")
	return nil
}

func (s *Supervisor) StopAll(ctx context.Context) {
	s.mu.RLock()
	ids := make([]string, 0, len(s.modules))
	for id := range s.modules {
		ids = append(ids, id)
	}
	s.mu.RUnlock()

	for _, id := range ids {
		if err := s.StopModule(ctx, id); err != nil {
			log.WithError(err).WithField("module", id).Warn("error stopping module")
		}
	}
}

func (s *Supervisor) GetModule(moduleID string) *ManagedModule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modules[moduleID]
}

func (s *Supervisor) GetModuleForListener(listenerID string) *ManagedModule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, mod := range s.modules {
		for _, l := range mod.Config.Listeners {
			if l.ID == listenerID {
				return mod
			}
		}
	}
	return nil
}

func (s *Supervisor) checkHealth(ctx context.Context) {
	s.mu.RLock()
	modules := make([]*ManagedModule, 0, len(s.modules))
	for _, m := range s.modules {
		modules = append(modules, m)
	}
	s.mu.RUnlock()

	for _, mod := range modules {
		if mod.State != ModuleStateRunning || mod.Client == nil {
			continue
		}

		healthy, status, err := mod.Client.HealthCheck(ctx)
		if err != nil || !healthy {
			log.WithFields(log.Fields{
				"module": mod.Config.ID,
				"status": status,
				"error":  err,
			}).Warn("module health check failed, restarting")

			s.restartModule(ctx, mod.Config.ID)
		}
	}
}

func (s *Supervisor) restartModule(ctx context.Context, moduleID string) {
	s.mu.RLock()
	mod, ok := s.modules[moduleID]
	if !ok {
		s.mu.RUnlock()
		return
	}
	cfg := mod.Config
	s.mu.RUnlock()

	s.StopModule(ctx, moduleID)
	if err := s.StartModule(ctx, cfg); err != nil {
		log.WithError(err).WithField("module", moduleID).Error("failed to restart module")
	}
}

func (s *Supervisor) HandleConnection(ctx context.Context, moduleID string, conn *ConnectionInfo) (int32, error) {
	mod := s.GetModule(moduleID)
	if mod == nil {
		return 0, fmt.Errorf("module %s not found", moduleID)
	}
	if mod.Client == nil {
		return 0, fmt.Errorf("module %s has no client", moduleID)
	}
	return mod.Client.HandleConnection(ctx, conn)
}

type ConnectionInfo struct {
	ConnectionID string
	SourceIP     string
	SourcePort   uint32
	DestIP       string
	DestPort     uint32
	Protocol     string
	FD           int
}
