package module

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

type RuntimeManager struct {
	mu       sync.RWMutex
	runtimes map[ExecutionMode]ModuleRuntime

	defaultRuntime ExecutionMode

	privilegedConfig *PrivilegedRuntimeConfig

	containerdAddress   string
	containerdNamespace string
}

type RuntimeManagerConfig struct {
	DefaultRuntime ExecutionMode

	EnablePrivileged bool

	PrivilegedConfig *PrivilegedRuntimeConfig

	EnableContainerd bool

	ContainerdAddress string

	ContainerdNamespace string
}

func NewRuntimeManager(cfg *RuntimeManagerConfig) (*RuntimeManager, error) {
	rm := &RuntimeManager{
		runtimes:       make(map[ExecutionMode]ModuleRuntime),
		defaultRuntime: cfg.DefaultRuntime,
	}

	if rm.defaultRuntime == "" {
		rm.defaultRuntime = ExecutionModeUserspace
	}

	rm.runtimes[ExecutionModeUserspace] = NewUserspaceRuntime()

	if cfg.EnablePrivileged {
		privCfg := cfg.PrivilegedConfig
		if privCfg == nil {
			privCfg = &PrivilegedRuntimeConfig{}
		}

		privRuntime, err := NewPrivilegedRuntime(privCfg)
		if err != nil {
			log.WithError(err).Warn("failed to initialize privileged runtime")
		} else {
			rm.runtimes[ExecutionModeRoot] = privRuntime
			rm.privilegedConfig = privCfg
		}
	}

	if cfg.EnableContainerd {
		rm.containerdAddress = cfg.ContainerdAddress
		rm.containerdNamespace = cfg.ContainerdNamespace

		if rm.containerdAddress == "" {
			rm.containerdAddress = "/run/containerd/containerd.sock"
		}

		containerdRuntime, err := NewContainerdRuntime(rm.containerdAddress, rm.containerdNamespace)
		if err != nil {
			log.WithError(err).Warn("failed to initialize containerd runtime")
		} else {
			rm.runtimes[ExecutionModeContainerd] = containerdRuntime
		}
	}

	log.WithFields(log.Fields{
		"default":  rm.defaultRuntime,
		"runtimes": rm.availableRuntimes(),
	}).Info("runtime manager initialized")

	return rm, nil
}

func (rm *RuntimeManager) availableRuntimes() []string {
	var names []string
	for mode := range rm.runtimes {
		names = append(names, string(mode))
	}
	return names
}

func (rm *RuntimeManager) GetRuntime(mode ExecutionMode) (ModuleRuntime, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if mode == "" {
		mode = rm.defaultRuntime
	}

	runtime, ok := rm.runtimes[mode]
	if !ok {
		return nil, fmt.Errorf("runtime %s not available", mode)
	}

	return runtime, nil
}

func (rm *RuntimeManager) SelectRuntime(spec *ModuleSpec) (ModuleRuntime, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if spec.ExecutionMode != "" {
		runtime, ok := rm.runtimes[spec.ExecutionMode]
		if !ok {
			return nil, fmt.Errorf("requested runtime %s not available", spec.ExecutionMode)
		}
		return runtime, nil
	}

	if spec.Capabilities.RequiresRoot || spec.Capabilities.CanHandleRawPackets {
		if runtime, ok := rm.runtimes[ExecutionModeRoot]; ok {
			return runtime, nil
		}
		log.WithField("module", spec.ID).Warn("module requires root but privileged runtime not available, using userspace")
	}

	return rm.runtimes[rm.defaultRuntime], nil
}

func (rm *RuntimeManager) StartModule(ctx context.Context, spec *ModuleSpec) (*ModuleInstance, error) {
	runtime, err := rm.SelectRuntime(spec)
	if err != nil {
		return nil, fmt.Errorf("selecting runtime: %w", err)
	}

	log.WithFields(log.Fields{
		"module":  spec.ID,
		"runtime": runtime.Type(),
	}).Debug("starting module with selected runtime")

	return runtime.Start(ctx, spec)
}

func (rm *RuntimeManager) StopModule(ctx context.Context, instance *ModuleInstance) error {
	runtime, err := rm.GetRuntime(instance.Spec.ExecutionMode)
	if err != nil {
		return fmt.Errorf("getting runtime: %w", err)
	}

	return runtime.Stop(ctx, instance)
}

func (rm *RuntimeManager) RegisterRuntime(mode ExecutionMode, runtime ModuleRuntime) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.runtimes[mode] = runtime
}

func (rm *RuntimeManager) IsRuntimeAvailable(mode ExecutionMode) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	_, ok := rm.runtimes[mode]
	return ok
}

func (rm *RuntimeManager) HealthCheck(ctx context.Context) map[ExecutionMode]error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	results := make(map[ExecutionMode]error)

	for mode, runtime := range rm.runtimes {
		if runtime == nil {
			results[mode] = fmt.Errorf("runtime is nil")
		} else {
			results[mode] = nil
		}
	}

	return results
}
