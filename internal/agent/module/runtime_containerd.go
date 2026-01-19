//go:build linux

package module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	log "github.com/sirupsen/logrus"
)

type ContainerdRuntime struct {
	client    *containerd.Client
	namespace string
}

func NewContainerdRuntime(address, namespace string) (*ContainerdRuntime, error) {
	client, err := containerd.New(address)
	if err != nil {
		return nil, fmt.Errorf("connecting to containerd: %w", err)
	}

	if namespace == "" {
		namespace = "gimpel"
	}

	return &ContainerdRuntime{
		client:    client,
		namespace: namespace,
	}, nil
}

func (r *ContainerdRuntime) Name() string {
	return "containerd"
}

func (r *ContainerdRuntime) Type() ExecutionMode {
	return ExecutionModeContainerd
}

func (r *ContainerdRuntime) Start(ctx context.Context, spec *ModuleSpec) (*ModuleInstance, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	var image containerd.Image
	var err error

	if _, statErr := os.Stat(spec.Image); statErr == nil {
		log.WithField("path", spec.Image).Info("importing module from local tar file")
		
		file, err := os.Open(spec.Image)
		if err != nil {
			return nil, fmt.Errorf("opening tar file: %w", err)
		}
		defer file.Close()

		imageRef := fmt.Sprintf("gimpel/%s:latest", spec.ID)
		
		imgs, err := r.client.Import(ctx, file, containerd.WithImageRefTranslator(
			func(_ string) string {
				return imageRef
			},
		))
		if err != nil {
			return nil, fmt.Errorf("importing image: %w", err)
		}
		
		if len(imgs) == 0 {
			return nil, fmt.Errorf("no images imported from tar")
		}
		
		image, err = r.client.GetImage(ctx, imgs[0].Name)
		if err != nil {
			return nil, fmt.Errorf("getting imported image: %w", err)
		}
		
		log.WithFields(log.Fields{
			"image_name": image.Name(),
			"digest":     imgs[0].Target.Digest,
		}).Info("module imported successfully")
	} else {
		image, err = r.client.GetImage(ctx, spec.Image)
		if err != nil {
			log.WithField("image", spec.Image).Debug("image not found locally, pulling")
			image, err = r.client.Pull(ctx, spec.Image, containerd.WithPullUnpack)
			if err != nil {
				return nil, fmt.Errorf("pulling image: %w", err)
			}
		}
	}

	socketDir := filepath.Dir(spec.SocketPath)
	if !filepath.IsAbs(socketDir) {
		absSocketDir, err := filepath.Abs(socketDir)
		if err != nil {
			return nil, fmt.Errorf("resolving absolute socket path: %w", err)
		}
		socketDir = absSocketDir
		spec.SocketPath = filepath.Join(socketDir, filepath.Base(spec.SocketPath))
	}
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return nil, fmt.Errorf("creating socket directory: %w", err)
	}

	envVars := []string{
		fmt.Sprintf("GIMPEL_SOCKET=%s", spec.SocketPath),
		fmt.Sprintf("GIMPEL_MODULE_ID=%s", spec.ID),
		fmt.Sprintf("GIMPEL_EXECUTION_MODE=%s", spec.ExecutionMode),
		fmt.Sprintf("GIMPEL_CONNECTION_MODE=%s", spec.ConnectionMode),
	}
	for k, v := range spec.Env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}

	container, err := r.client.NewContainer(
		ctx,
		spec.ID,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(spec.ID+"-snapshot", image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithEnv(envVars),
			oci.WithHostNamespace(specs.NetworkNamespace),
			oci.WithMounts([]specs.Mount{
				{
					Destination: socketDir,
					Type:        "bind",
					Source:      socketDir,
					Options:     []string{"rbind", "rw"},
				},
			}),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating container: %w", err)
	}

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		container.Delete(ctx, containerd.WithSnapshotCleanup)
		return nil, fmt.Errorf("creating task: %w", err)
	}

	if err := task.Start(ctx); err != nil {
		task.Delete(ctx)
		container.Delete(ctx, containerd.WithSnapshotCleanup)
		return nil, fmt.Errorf("starting task: %w", err)
	}

	if err := waitForSocket(spec.SocketPath, 30*time.Second); err != nil {
		task.Kill(ctx, 9)
		task.Delete(ctx)
		container.Delete(ctx, containerd.WithSnapshotCleanup)
		return nil, fmt.Errorf("waiting for socket: %w", err)
	}

	instance := &ModuleInstance{
		ID:          spec.ID,
		Spec:        spec,
		ContainerID: container.ID(),
		SocketPath:  spec.SocketPath,
		StartedAt:   time.Now(),
		State:       ModuleStateRunning,
		Metrics:     &ModuleMetrics{},
		StopFunc: func() {
			stopCtx := context.Background()
			stopCtx = namespaces.WithNamespace(stopCtx, r.namespace)
			task.Kill(stopCtx, 15)
			task.Delete(stopCtx)
			container.Delete(stopCtx, containerd.WithSnapshotCleanup)
		},
	}

	log.WithFields(log.Fields{
		"module":    spec.ID,
		"image":     spec.Image,
		"container": container.ID(),
	}).Info("containerd module started")

	return instance, nil
}

func (r *ContainerdRuntime) Stop(ctx context.Context, instance *ModuleInstance) error {
	if instance.StopFunc != nil {
		instance.StopFunc()
	}
	log.WithField("module", instance.ID).Info("containerd module stopped")
	return nil
}

func (r *ContainerdRuntime) Signal(ctx context.Context, instance *ModuleInstance, signal int) error {
	return fmt.Errorf("signal not implemented for containerd runtime")
}

func (r *ContainerdRuntime) IsRunning(ctx context.Context, instance *ModuleInstance) bool {
	if instance.ContainerID == "" {
		return false
	}
	ctx = namespaces.WithNamespace(ctx, r.namespace)
	container, err := r.client.LoadContainer(ctx, instance.ContainerID)
	if err != nil {
		return false
	}
	task, err := container.Task(ctx, nil)
	if err != nil {
		return false
	}
	status, err := task.Status(ctx)
	if err != nil {
		return false
	}
	return status.Status == containerd.Running
}

func (r *ContainerdRuntime) Logs(ctx context.Context, instance *ModuleInstance, lines int) ([]string, error) {
	return nil, fmt.Errorf("logs not implemented for containerd runtime")
}

func (r *ContainerdRuntime) Close() error {
	return r.client.Close()
}
