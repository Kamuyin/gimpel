package module

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

type UserspaceRuntime struct{}

func NewUserspaceRuntime() *UserspaceRuntime {
	log.Warn("⚠️  Using USERSPACE runtime - NO ISOLATION! This is for development only.")
	return &UserspaceRuntime{}
}

func (r *UserspaceRuntime) Type() RuntimeType {
	return RuntimeTypeUserspace
}

func (r *UserspaceRuntime) Start(ctx context.Context, spec *RuntimeSpec) (*RuntimeInstance, error) {
	socketDir := filepath.Dir(spec.SocketPath)
	if err := os.MkdirAll(socketDir, 0700); err != nil {
		return nil, fmt.Errorf("creating socket dir: %w", err)
	}

	os.Remove(spec.SocketPath)

	cmd := exec.CommandContext(ctx, spec.Image)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIMPEL_SOCKET=%s", spec.SocketPath),
		fmt.Sprintf("GIMPEL_MODULE_ID=%s", spec.ID),
	)
	for k, v := range spec.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting process: %w", err)
	}

	if err := waitForSocket(spec.SocketPath, 10*time.Second); err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("waiting for socket: %w", err)
	}

	instance := &RuntimeInstance{
		ID:         spec.ID,
		Pid:        cmd.Process.Pid,
		SocketPath: spec.SocketPath,
		StopFunc: func() {
			cmd.Process.Signal(os.Interrupt)
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				cmd.Process.Kill()
			}
		},
	}

	log.WithFields(log.Fields{
		"module": spec.ID,
		"pid":    cmd.Process.Pid,
		"socket": spec.SocketPath,
	}).Info("userspace module started")

	return instance, nil
}

func (r *UserspaceRuntime) Stop(ctx context.Context, instance *RuntimeInstance) error {
	instance.Stop()
	log.WithField("module", instance.ID).Info("userspace module stopped")
	return nil
}
