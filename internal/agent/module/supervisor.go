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

type Supervisor struct {
	cfg     *config.AgentConfig
	emitter *telemetry.Emitter

	runtimeMgr *RuntimeManager
	forwarder  *ConnectionForwarder

	mu        sync.RWMutex
	instances map[string]*ModuleInstance
	clients   map[string]*Client

	healthInterval time.Duration
	healthTimeout  time.Duration
}

func NewSupervisor(cfg *config.AgentConfig, emitter *telemetry.Emitter) (*Supervisor, error) {
	runtimeMgrCfg := &RuntimeManagerConfig{
		DefaultRuntime:      ExecutionMode(cfg.Runtime.DefaultExecutionMode),
		EnablePrivileged:    cfg.Runtime.EnablePrivileged,
		EnableContainerd:    cfg.Runtime.EnableContainerd,
		ContainerdAddress:   cfg.Runtime.ContainerdAddress,
		ContainerdNamespace: cfg.Runtime.ContainerdNamespace,
	}

	if runtimeMgrCfg.DefaultRuntime == "" {
		runtimeMgrCfg.DefaultRuntime = ExecutionModeUserspace
	}

	runtimeMgr, err := NewRuntimeManager(runtimeMgrCfg)
	if err != nil {
		return nil, fmt.Errorf("creating runtime manager: %w", err)
	}

	defaultConnMode := ConnectionMode(cfg.Runtime.DefaultConnectionMode)
	if defaultConnMode == "" {
		defaultConnMode = ConnectionModeTCPRelay
	}
	forwarder := NewConnectionForwarder(defaultConnMode)

	s := &Supervisor{
		cfg:            cfg,
		emitter:        emitter,
		runtimeMgr:     runtimeMgr,
		forwarder:      forwarder,
		instances:      make(map[string]*ModuleInstance),
		clients:        make(map[string]*Client),
		healthInterval: 10 * time.Second,
		healthTimeout:  5 * time.Second,
	}

	return s, nil
}

func (s *Supervisor) Run(ctx context.Context) error {
	log.Info("starting module supervisor v2")

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

	if existing, ok := s.instances[cfg.ID]; ok && existing.State == ModuleStateRunning {
		log.WithField("module", cfg.ID).Debug("module already running")
		return nil
	}

	spec := s.configToSpec(cfg)

	log.WithFields(log.Fields{
		"module":         spec.ID,
		"execution_mode": spec.ExecutionMode,
		"connection_mode": spec.ConnectionMode,
		"image":          spec.Image,
	}).Info("starting module")

	instance, err := s.runtimeMgr.StartModule(ctx, spec)
	if err != nil {
		return fmt.Errorf("starting module: %w", err)
	}

	s.instances[cfg.ID] = instance

	client, err := NewClient(instance.SocketPath)
	if err != nil {
		if instance.StopFunc != nil {
			instance.StopFunc()
		}
		delete(s.instances, cfg.ID)
		return fmt.Errorf("creating module client: %w", err)
	}
	s.clients[cfg.ID] = client

	connMode := ConnectionMode(cfg.ConnectionMode)
	if connMode == "" {
		connMode = spec.ConnectionMode
	}
	if err := s.forwarder.RegisterModule(cfg.ID, instance.SocketPath, instance.DataPort, connMode); err != nil {
		log.WithError(err).WithField("module", cfg.ID).Warn("failed to register connection forwarder")
	}

	log.WithFields(log.Fields{
		"module": cfg.ID,
		"pid":    instance.PID,
		"socket": instance.SocketPath,
	}).Info("module started successfully")

	return nil
}

func (s *Supervisor) StopModule(ctx context.Context, moduleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	instance, ok := s.instances[moduleID]
	if !ok {
		return nil
	}

	if client, ok := s.clients[moduleID]; ok {
		client.Close()
		delete(s.clients, moduleID)
	}

	s.forwarder.UnregisterModule(moduleID)

	if err := s.runtimeMgr.StopModule(ctx, instance); err != nil {
		log.WithError(err).WithField("module", moduleID).Warn("error stopping module")
	}

	delete(s.instances, moduleID)

	log.WithField("module", moduleID).Info("module stopped")
	return nil
}

func (s *Supervisor) StopAll(ctx context.Context) {
	s.mu.RLock()
	ids := make([]string, 0, len(s.instances))
	for id := range s.instances {
		ids = append(ids, id)
	}
	s.mu.RUnlock()

	for _, id := range ids {
		if err := s.StopModule(ctx, id); err != nil {
			log.WithError(err).WithField("module", id).Warn("error stopping module")
		}
	}
}

func (s *Supervisor) GetInstance(moduleID string) *ModuleInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.instances[moduleID]
}

func (s *Supervisor) GetClient(moduleID string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[moduleID]
}

func (s *Supervisor) HandleConnection(ctx context.Context, moduleID string, conn *ConnectionInfo) (int32, error) {
	client := s.GetClient(moduleID)
	if client == nil {
		return 0, fmt.Errorf("module %s not found or not running", moduleID)
	}

	return client.HandleConnection(ctx, conn)
}

func (s *Supervisor) ForwardConnection(ctx context.Context, req *ConnectionRequest) error {
	return s.forwarder.Forward(ctx, req)
}

func (s *Supervisor) configToSpec(cfg config.ModuleConfig) *ModuleSpec {
	execMode := ExecutionMode(cfg.ExecutionMode)
	if execMode == "" {
		execMode = ExecutionMode(s.cfg.Runtime.DefaultExecutionMode)
	}
	if execMode == "" {
		execMode = ExecutionModeUserspace
	}

	if cfg.RequiresRoot && execMode == ExecutionModeUserspace {
		if s.runtimeMgr.IsRuntimeAvailable(ExecutionModeRoot) {
			execMode = ExecutionModeRoot
		}
	}

	connMode := ConnectionMode(cfg.ConnectionMode)
	if connMode == "" {
		connMode = ConnectionMode(s.cfg.Runtime.DefaultConnectionMode)
	}
	if connMode == "" {
		connMode = ConnectionModeTCPRelay
	}

	socketPath := cfg.SocketPath
	if socketPath == "" {
		socketPath = fmt.Sprintf("%s/modules/%s.sock", s.cfg.DataDir, cfg.ID)
	}

	spec := &ModuleSpec{
		ID:             cfg.ID,
		Name:           cfg.Name,
		Image:          cfg.Image,
		ExecutionMode:  execMode,
		ConnectionMode: connMode,
		SocketPath:     socketPath,
		Env:            cfg.Env,
		WorkingDir:     cfg.WorkingDir,
		Capabilities: ModuleCapabilities{
			RequiresRoot:        cfg.RequiresRoot,
			CanHandleRawPackets: cfg.CanHandleRawPackets,
		},
		ResourceLimits: ResourceLimits{
			MaxMemoryMB:   cfg.ResourceLimits.MaxMemoryMB,
			MaxCPUPercent: cfg.ResourceLimits.MaxCPUPercent,
			MaxOpenFiles:  cfg.ResourceLimits.MaxOpenFiles,
			MaxProcesses:  cfg.ResourceLimits.MaxProcesses,
		},
		RestartPolicy: RestartPolicy{
			Policy:            cfg.RestartPolicy.Policy,
			MaxRestarts:       cfg.RestartPolicy.MaxRestarts,
			RestartDelay:      cfg.RestartPolicy.RestartDelay,
			BackoffMultiplier: cfg.RestartPolicy.BackoffMultiplier,
			MaxBackoffDelay:   cfg.RestartPolicy.MaxBackoffDelay,
		},
		HealthCheck: HealthCheckConfig{
			Enabled:  cfg.HealthCheck.Enabled,
			Interval: cfg.HealthCheck.Interval,
			Timeout:  cfg.HealthCheck.Timeout,
			Retries:  cfg.HealthCheck.Retries,
		},
	}

	if spec.HealthCheck.Interval == 0 {
		spec.HealthCheck.Interval = 10 * time.Second
	}
	if spec.HealthCheck.Timeout == 0 {
		spec.HealthCheck.Timeout = 5 * time.Second
	}
	if spec.HealthCheck.Retries == 0 {
		spec.HealthCheck.Retries = 3
	}

	if spec.RestartPolicy.Policy == "" {
		spec.RestartPolicy.Policy = "on-failure"
	}
	if spec.RestartPolicy.RestartDelay == 0 {
		spec.RestartPolicy.RestartDelay = 1 * time.Second
	}
	if spec.RestartPolicy.BackoffMultiplier == 0 {
		spec.RestartPolicy.BackoffMultiplier = 2.0
	}
	if spec.RestartPolicy.MaxBackoffDelay == 0 {
		spec.RestartPolicy.MaxBackoffDelay = 5 * time.Minute
	}

	return spec
}

func (s *Supervisor) checkHealth(ctx context.Context) {
	s.mu.RLock()
	instances := make([]*ModuleInstance, 0, len(s.instances))
	for _, inst := range s.instances {
		instances = append(instances, inst)
	}
	s.mu.RUnlock()

	for _, inst := range instances {
		if inst.State != ModuleStateRunning {
			continue
		}

		client := s.GetClient(inst.ID)
		if client == nil {
			continue
		}

		healthCtx, cancel := context.WithTimeout(ctx, s.healthTimeout)
		healthy, status, err := client.HealthCheck(healthCtx)
		cancel()

		if err != nil || !healthy {
			log.WithFields(log.Fields{
				"module": inst.ID,
				"status": status,
				"error":  err,
			}).Warn("module health check failed")

			if inst.Metrics != nil {
				inst.Metrics.HealthChecksFailed++
			}

			s.handleUnhealthyModule(ctx, inst)
		} else {
			if inst.Metrics != nil {
				inst.Metrics.HealthChecksPassed++
				inst.Metrics.LastHealthCheck = time.Now()
			}
		}
	}
}

func (s *Supervisor) handleUnhealthyModule(ctx context.Context, inst *ModuleInstance) {
	spec := inst.Spec
	if spec == nil {
		return
	}

	switch spec.RestartPolicy.Policy {
	case "always", "on-failure":
		if spec.RestartPolicy.MaxRestarts > 0 && inst.RestartCount >= spec.RestartPolicy.MaxRestarts {
			log.WithField("module", inst.ID).Error("max restarts exceeded, not restarting")
			return
		}

		delay := spec.RestartPolicy.RestartDelay
		for i := 0; i < inst.RestartCount; i++ {
			delay = time.Duration(float64(delay) * spec.RestartPolicy.BackoffMultiplier)
			if delay > spec.RestartPolicy.MaxBackoffDelay {
				delay = spec.RestartPolicy.MaxBackoffDelay
				break
			}
		}

		log.WithFields(log.Fields{
			"module":        inst.ID,
			"restart_count": inst.RestartCount + 1,
			"delay":         delay,
		}).Info("scheduling module restart")

		go func() {
			time.Sleep(delay)
			s.restartModule(ctx, inst.ID)
		}()

	case "never":
		log.WithField("module", inst.ID).Warn("module unhealthy but restart policy is 'never'")

	default:
		log.WithFields(log.Fields{
			"module": inst.ID,
			"policy": spec.RestartPolicy.Policy,
		}).Warn("unknown restart policy")
	}
}

func (s *Supervisor) restartModule(ctx context.Context, moduleID string) {
	s.mu.RLock()
	inst, ok := s.instances[moduleID]
	if !ok {
		s.mu.RUnlock()
		return
	}
	spec := inst.Spec
	restartCount := inst.RestartCount
	s.mu.RUnlock()

	var cfg config.ModuleConfig
	for _, c := range s.cfg.Modules {
		if c.ID == moduleID {
			cfg = c
			break
		}
	}

	if err := s.StopModule(ctx, moduleID); err != nil {
		log.WithError(err).WithField("module", moduleID).Error("failed to stop module for restart")
		return
	}

	if err := s.StartModule(ctx, cfg); err != nil {
		log.WithError(err).WithField("module", moduleID).Error("failed to restart module")
		return
	}

	s.mu.Lock()
	if newInst, ok := s.instances[moduleID]; ok {
		newInst.RestartCount = restartCount + 1
		newInst.Spec = spec
	}
	s.mu.Unlock()
}

func (s *Supervisor) GetMetrics(moduleID string) *ModuleMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if inst, ok := s.instances[moduleID]; ok {
		return inst.Metrics
	}
	return nil
}

func (s *Supervisor) ListModules() []ModuleInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var infos []ModuleInfo
	for id, inst := range s.instances {
		info := ModuleInfo{
			ID:           id,
			Name:         inst.Spec.Name,
			State:        inst.State,
			PID:          inst.PID,
			StartedAt:    inst.StartedAt,
			RestartCount: inst.RestartCount,
		}
		if inst.LastError != nil {
			info.LastError = inst.LastError.Error()
		}
		infos = append(infos, info)
	}

	return infos
}

type ModuleInfo struct {
	ID           string
	Name         string
	State        ModuleState
	PID          int
	StartedAt    time.Time
	RestartCount int
	LastError    string
}
