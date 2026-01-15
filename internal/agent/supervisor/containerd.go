package supervisor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

type ContainerdRuntime struct {
	client      *containerd.Client
	modulesPath string
	log         *logrus.Logger
}

func NewContainerdRuntime(socketAddr, modulesPath string, log *logrus.Logger) (*ContainerdRuntime, error) {
	client, err := containerd.New(socketAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to containerd: %w", err)
	}

	return &ContainerdRuntime{
		client:      client,
		modulesPath: modulesPath,
		log:         log,
	}, nil
}

func (r *ContainerdRuntime) Start(ctx context.Context, cfg ModuleConfig) (string, error) {
	ctx = namespaces.WithNamespace(ctx, "gimpel")

	image, err := r.client.Pull(ctx, cfg.Image, containerd.WithPullUnpack)
	if err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}

	hostSockDir := filepath.Join(r.modulesPath, cfg.ID)
	if err := os.MkdirAll(hostSockDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create module dir: %w", err)
	}

	containerSockDir := "/run/gimpel"

	mounts := []specs.Mount{
		{
			Type:        "bind",
			Source:      hostSockDir,
			Destination: containerSockDir,
			Options:     []string{"rbind", "rw"},
		},
	}

	if err := r.client.ContainerService().Delete(ctx, cfg.ID); err != nil {
		r.log.WithError(err).Debug("Failed to delete existing container")
	}

	container, err := r.client.NewContainer(
		ctx,
		cfg.ID,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(cfg.ID+"-snapshot", image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithMounts(mounts),
		),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	if err := task.Start(ctx); err != nil {
		task.Delete(ctx)
		container.Delete(ctx, containerd.WithSnapshotCleanup)
		return "", fmt.Errorf("failed to start task: %w", err)
	}

	r.log.WithField("module", cfg.ID).Info("Container started")

	socketPath := filepath.Join(hostSockDir, "control.sock")
	return socketPath, nil
}

func (r *ContainerdRuntime) Stop(ctx context.Context, id string) error {
	ctx = namespaces.WithNamespace(ctx, "gimpel")

	container, err := r.client.LoadContainer(ctx, id)
	if err != nil {
		return err
	}

	task, err := container.Task(ctx, cio.Load)
	if err == nil {
		task.Kill(ctx, syscall.SIGKILL)
		_, err = task.Delete(ctx)
		if err != nil {
			r.log.WithError(err).Warn("Failed to delete task")
		}
	}

	return container.Delete(ctx, containerd.WithSnapshotCleanup)
}

func isDirectoryNotFoundError(err error) bool {
	return false
}
