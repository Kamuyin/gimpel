package module

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type UserspaceRuntime struct{}

func NewUserspaceRuntime() *UserspaceRuntime {
	log.Warn("⚠️  Using USERSPACE runtime - NO ISOLATION! This is for development only.")
	return &UserspaceRuntime{}
}

func (r *UserspaceRuntime) Name() string {
	return "userspace"
}

func (r *UserspaceRuntime) Type() ExecutionMode {
	return ExecutionModeUserspace
}

func (r *UserspaceRuntime) Start(ctx context.Context, spec *ModuleSpec) (*ModuleInstance, error) {
	socketDir := filepath.Dir(spec.SocketPath)
	if err := os.MkdirAll(socketDir, 0700); err != nil {
		return nil, fmt.Errorf("creating socket dir: %w", err)
	}

	os.Remove(spec.SocketPath)

	procCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(procCtx, spec.Image)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIMPEL_SOCKET=%s", spec.SocketPath),
		fmt.Sprintf("GIMPEL_MODULE_ID=%s", spec.ID),
		fmt.Sprintf("GIMPEL_EXECUTION_MODE=%s", spec.ExecutionMode),
		fmt.Sprintf("GIMPEL_CONNECTION_MODE=%s", spec.ConnectionMode),
	)
	for k, v := range spec.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	if spec.WorkingDir != "" {
		cmd.Dir = spec.WorkingDir
	}

	cmd.Stdout = &moduleLogger{moduleID: spec.ID, level: "info"}
	cmd.Stderr = &moduleLogger{moduleID: spec.ID, level: "error"}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("starting process: %w", err)
	}

	if err := waitForSocket(spec.SocketPath, 10*time.Second); err != nil {
		cmd.Process.Kill()
		cancel()
		return nil, fmt.Errorf("waiting for socket: %w", err)
	}

	instance := &ModuleInstance{
		ID:         spec.ID,
		Spec:       spec,
		PID:        cmd.Process.Pid,
		SocketPath: spec.SocketPath,
		StartedAt:  time.Now(),
		State:      ModuleStateRunning,
		Metrics:    &ModuleMetrics{},
		StopFunc: func() {
			cmd.Process.Signal(os.Interrupt)
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				cmd.Process.Kill()
			}
			cancel()
		},
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			instance.LastError = err
			instance.State = ModuleStateFailed
		} else {
			instance.State = ModuleStateStopped
		}
	}()

	log.WithFields(log.Fields{
		"module": spec.ID,
		"pid":    cmd.Process.Pid,
		"socket": spec.SocketPath,
	}).Info("userspace module started")

	return instance, nil
}

func (r *UserspaceRuntime) Stop(ctx context.Context, instance *ModuleInstance) error {
	if instance.StopFunc != nil {
		instance.StopFunc()
	}
	log.WithField("module", instance.ID).Info("userspace module stopped")
	return nil
}

func (r *UserspaceRuntime) Signal(ctx context.Context, instance *ModuleInstance, signal int) error {
	if instance.PID == 0 {
		return fmt.Errorf("no PID for module %s", instance.ID)
	}
	proc, err := os.FindProcess(instance.PID)
	if err != nil {
		return fmt.Errorf("finding process: %w", err)
	}
	return proc.Signal(syscall.Signal(signal))
}

func (r *UserspaceRuntime) IsRunning(ctx context.Context, instance *ModuleInstance) bool {
	if instance.PID == 0 {
		return false
	}
	proc, err := os.FindProcess(instance.PID)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

func (r *UserspaceRuntime) Logs(ctx context.Context, instance *ModuleInstance, lines int) ([]string, error) {
	return nil, fmt.Errorf("logs not available for userspace runtime")
}
