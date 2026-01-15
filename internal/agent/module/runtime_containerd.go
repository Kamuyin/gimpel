//go:build linux

package module

import (
	"context"
	"fmt"
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

func (r *ContainerdRuntime) Type() RuntimeType {
	return RuntimeTypeContainerd
}

func (r *ContainerdRuntime) Start(ctx context.Context, spec *RuntimeSpec) (*RuntimeInstance, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	image, err := r.client.Pull(ctx, spec.Image, containerd.WithPullUnpack)
	if err != nil {
		image, err = r.client.GetImage(ctx, spec.Image)
		if err != nil {
			return nil, fmt.Errorf("getting image: %w", err)
		}
	}

	envVars := []string{
		fmt.Sprintf("GIMPEL_SOCKET=/run/gimpel/module.sock"),
		fmt.Sprintf("GIMPEL_MODULE_ID=%s", spec.ID),
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

	instance := &RuntimeInstance{
		ID:         spec.ID,
		SocketPath: spec.SocketPath,
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

func (r *ContainerdRuntime) Stop(ctx context.Context, instance *RuntimeInstance) error {
	instance.Stop()
	log.WithField("module", instance.ID).Info("containerd module stopped")
	return nil
}

func (r *ContainerdRuntime) Close() error {
	return r.client.Close()
}
