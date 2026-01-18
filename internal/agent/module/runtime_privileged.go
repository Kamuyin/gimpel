//go:build linux || darwin || freebsd || openbsd || netbsd

package module

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type PrivilegedRuntime struct {
	sudoPath string

	dropPrivileges bool

	targetUser string

	targetGroup string

	useCapabilities bool

	requiredCapabilities []string
}

type PrivilegedRuntimeConfig struct {
	SudoPath string

	DropPrivileges bool

	TargetUser string

	TargetGroup string

	UseCapabilities bool

	RequiredCapabilities []string
}

func NewPrivilegedRuntime(cfg *PrivilegedRuntimeConfig) (*PrivilegedRuntime, error) {
	sudoPath := cfg.SudoPath
	if sudoPath == "" {
		var err error
		sudoPath, err = exec.LookPath("sudo")
		if err != nil {
			return nil, fmt.Errorf("sudo not found in PATH: %w", err)
		}
	}

	if _, err := os.Stat(sudoPath); err != nil {
		return nil, fmt.Errorf("sudo not accessible: %w", err)
	}

	targetUser := cfg.TargetUser
	if targetUser == "" {
		currentUser, err := user.Current()
		if err == nil {
			targetUser = currentUser.Username
		}
	}

	log.WithFields(log.Fields{
		"sudo_path":        sudoPath,
		"drop_privileges":  cfg.DropPrivileges,
		"target_user":      targetUser,
		"use_capabilities": cfg.UseCapabilities,
	}).Warn("⚠️  Privileged runtime initialized - modules will run with elevated privileges")

	return &PrivilegedRuntime{
		sudoPath:             sudoPath,
		dropPrivileges:       cfg.DropPrivileges,
		targetUser:           targetUser,
		targetGroup:          cfg.TargetGroup,
		useCapabilities:      cfg.UseCapabilities,
		requiredCapabilities: cfg.RequiredCapabilities,
	}, nil
}

func (r *PrivilegedRuntime) Name() string {
	return "privileged"
}

func (r *PrivilegedRuntime) Type() ExecutionMode {
	return ExecutionModeRoot
}

func (r *PrivilegedRuntime) Start(ctx context.Context, spec *ModuleSpec) (*ModuleInstance, error) {
	if spec.Image == "" {
		return nil, fmt.Errorf("module image path is required")
	}

	socketDir := filepath.Dir(spec.SocketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return nil, fmt.Errorf("creating socket dir: %w", err)
	}

	os.Remove(spec.SocketPath)

	args := r.buildCommandArgs(spec)

	log.WithFields(log.Fields{
		"module":  spec.ID,
		"command": args,
	}).Debug("starting privileged module")

	procCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(procCtx, args[0], args[1:]...)

	cmd.Env = r.buildEnvironment(spec)

	if spec.WorkingDir != "" {
		cmd.Dir = spec.WorkingDir
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	cmd.Stdout = &moduleLogger{moduleID: spec.ID, level: "info"}
	cmd.Stderr = &moduleLogger{moduleID: spec.ID, level: "error"}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("starting privileged process: %w", err)
	}

	if err := waitForSocket(spec.SocketPath, 30*time.Second); err != nil {
		cmd.Process.Kill()
		cancel()
		return nil, fmt.Errorf("waiting for module socket: %w", err)
	}

	if err := os.Chmod(spec.SocketPath, 0666); err != nil {
		log.WithError(err).Warn("failed to change socket permissions")
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
			cmd.Process.Signal(syscall.SIGTERM)
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case <-done:
			case <-time.After(10 * time.Second):
				cmd.Process.Kill()
				cmd.Process.Signal(syscall.SIGKILL)
			}
			cancel()
		},
	}

	go r.monitorProcess(cmd, instance)

	log.WithFields(log.Fields{
		"module": spec.ID,
		"pid":    cmd.Process.Pid,
		"socket": spec.SocketPath,
	}).Info("privileged module started")

	return instance, nil
}

func (r *PrivilegedRuntime) buildCommandArgs(spec *ModuleSpec) []string {
	args := []string{r.sudoPath}

	args = append(args, "-E")

	if r.useCapabilities && len(r.requiredCapabilities) > 0 {
		args = append(args, "capsh")
		caps := strings.Join(r.requiredCapabilities, ",")
		args = append(args, "--caps="+caps+"+eip")
		args = append(args, "--")
	}

	if r.dropPrivileges && r.targetUser != "" {
		args = append(args, "-u", r.targetUser)
		if r.targetGroup != "" {
			args = append(args, "-g", r.targetGroup)
		}
	}

	args = append(args, spec.Image)

	return args
}

func (r *PrivilegedRuntime) buildEnvironment(spec *ModuleSpec) []string {
	env := os.Environ()

	env = append(env,
		fmt.Sprintf("GIMPEL_SOCKET=%s", spec.SocketPath),
		fmt.Sprintf("GIMPEL_MODULE_ID=%s", spec.ID),
		fmt.Sprintf("GIMPEL_EXECUTION_MODE=%s", spec.ExecutionMode),
		fmt.Sprintf("GIMPEL_CONNECTION_MODE=%s", spec.ConnectionMode),
	)

	for k, v := range spec.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

func (r *PrivilegedRuntime) monitorProcess(cmd *exec.Cmd, instance *ModuleInstance) {
	err := cmd.Wait()
	if err != nil {
		instance.LastError = err
		instance.State = ModuleStateFailed
		log.WithError(err).WithField("module", instance.ID).Error("privileged module exited with error")
	} else {
		instance.State = ModuleStateStopped
		log.WithField("module", instance.ID).Info("privileged module exited normally")
	}
}

func (r *PrivilegedRuntime) Stop(ctx context.Context, instance *ModuleInstance) error {
	if instance.StopFunc != nil {
		instance.StopFunc()
	}
	log.WithField("module", instance.ID).Info("privileged module stopped")
	return nil
}

func (r *PrivilegedRuntime) Signal(ctx context.Context, instance *ModuleInstance, signal int) error {
	if instance.PID == 0 {
		return fmt.Errorf("no PID for module %s", instance.ID)
	}

	proc, err := os.FindProcess(instance.PID)
	if err != nil {
		return fmt.Errorf("finding process: %w", err)
	}

	return proc.Signal(syscall.Signal(signal))
}

func (r *PrivilegedRuntime) IsRunning(ctx context.Context, instance *ModuleInstance) bool {
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

func (r *PrivilegedRuntime) Logs(ctx context.Context, instance *ModuleInstance, lines int) ([]string, error) {
	return nil, fmt.Errorf("logs not available for privileged runtime")
}

type moduleLoggerPriv struct {
	moduleID string
	level    string
	buffer   strings.Builder
}

func (l *moduleLoggerPriv) Write(p []byte) (n int, err error) {
	l.buffer.Write(p)

	for {
		line, rest, found := strings.Cut(l.buffer.String(), "\n")
		if !found {
			break
		}
		l.buffer.Reset()
		l.buffer.WriteString(rest)

		entry := log.WithField("module", l.moduleID)
		switch l.level {
		case "error":
			entry.Error(line)
		default:
			entry.Info(line)
		}
	}

	return len(p), nil
}

func IsRoot() bool {
	return os.Getuid() == 0
}

func CanUseSudo() bool {
	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		return false
	}

	cmd := exec.Command(sudoPath, "-n", "true")
	return cmd.Run() == nil
}

func GetUserID(username string) (int, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(u.Uid)
}

func GetGroupID(groupname string) (int, error) {
	g, err := user.LookupGroup(groupname)
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(g.Gid)
}
