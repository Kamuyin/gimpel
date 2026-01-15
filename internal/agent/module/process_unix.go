//go:build linux || darwin || freebsd || openbsd || netbsd

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

	"gimpel/internal/agent/config"
)

type Process struct {
	cfg     config.ModuleConfig
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	exited  chan struct{}
	exitErr error
}

func StartProcess(ctx context.Context, cfg config.ModuleConfig) (*Process, error) {
	socketDir := filepath.Dir(cfg.SocketPath)
	if err := os.MkdirAll(socketDir, 0700); err != nil {
		return nil, fmt.Errorf("creating socket dir: %w", err)
	}

	os.Remove(cfg.SocketPath)

	procCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(procCtx, cfg.Image)

	cmd.Env = append(os.Environ(), fmt.Sprintf("GIMPEL_SOCKET=%s", cfg.SocketPath))
	for k, v := range cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	cmd.Stdout = &moduleLogger{moduleID: cfg.ID, level: "info"}
	cmd.Stderr = &moduleLogger{moduleID: cfg.ID, level: "error"}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("starting command: %w", err)
	}

	p := &Process{
		cfg:    cfg,
		cmd:    cmd,
		cancel: cancel,
		exited: make(chan struct{}),
	}

	go func() {
		p.exitErr = cmd.Wait()
		close(p.exited)
	}()

	if err := waitForSocket(cfg.SocketPath, 10*time.Second); err != nil {
		p.Stop()
		return nil, fmt.Errorf("waiting for socket: %w", err)
	}

	return p, nil
}

func (p *Process) Stop() {
	if p.cmd == nil || p.cmd.Process == nil {
		return
	}

	p.cmd.Process.Signal(syscall.SIGTERM)

	select {
	case <-p.exited:
		return
	case <-time.After(5 * time.Second):
		p.cmd.Process.Kill()
		<-p.exited
	}
}

func (p *Process) Wait() error {
	<-p.exited
	return p.exitErr
}

func (p *Process) Running() bool {
	select {
	case <-p.exited:
		return false
	default:
		return true
	}
}

func waitForSocket(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if info, err := os.Stat(path); err == nil && (info.Mode()&os.ModeSocket) != 0 {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("socket not created within timeout")
}

type moduleLogger struct {
	moduleID string
	level    string
}

func (l *moduleLogger) Write(p []byte) (n int, err error) {
	entry := log.WithField("module", l.moduleID)
	switch l.level {
	case "error":
		entry.Error(string(p))
	default:
		entry.Info(string(p))
	}
	return len(p), nil
}
