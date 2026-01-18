//go:build linux

package moduleclient

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/namespaces"
	log "github.com/sirupsen/logrus"
)

type ContainerdImporter struct {
	client    *containerd.Client
	namespace string
}

func NewContainerdImporter(address, namespace string) (*ContainerdImporter, error) {
	client, err := containerd.New(address)
	if err != nil {
		return nil, fmt.Errorf("connecting to containerd: %w", err)
	}

	if namespace == "" {
		namespace = "gimpel"
	}

	return &ContainerdImporter{
		client:    client,
		namespace: namespace,
	}, nil
}

func (i *ContainerdImporter) Close() error {
	return i.client.Close()
}

func (i *ContainerdImporter) ImportImage(ctx context.Context, tarPath, imageName string) error {
	ctx = namespaces.WithNamespace(ctx, i.namespace)

	file, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("opening tarball: %w", err)
	}
	defer file.Close()

	imgs, err := i.client.Import(ctx, file, containerd.WithImageRefTranslator(func(ref string) string {
		return imageName
	}))
	if err != nil {
		return fmt.Errorf("importing image: %w", err)
	}

	if len(imgs) == 0 {
		return fmt.Errorf("no images imported from tarball")
	}

	for _, img := range imgs {
		image := containerd.NewImage(i.client, img)
		if err := image.Unpack(ctx, containerd.DefaultSnapshotter); err != nil {
			log.WithError(err).WithField("image", img.Name).Warn("failed to unpack image")
		}
	}

	log.WithFields(log.Fields{
		"tarball":      tarPath,
		"image":        imageName,
		"images_count": len(imgs),
	}).Info("image imported into containerd")

	return nil
}

func (i *ContainerdImporter) HasImage(ctx context.Context, imageName string) bool {
	ctx = namespaces.WithNamespace(ctx, i.namespace)

	_, err := i.client.GetImage(ctx, imageName)
	return err == nil
}

func (i *ContainerdImporter) DeleteImage(ctx context.Context, imageName string) error {
	ctx = namespaces.WithNamespace(ctx, i.namespace)

	imageService := i.client.ImageService()
	return imageService.Delete(ctx, imageName, images.SynchronousDelete())
}

func (i *ContainerdImporter) ListImages(ctx context.Context) ([]string, error) {
	ctx = namespaces.WithNamespace(ctx, i.namespace)

	imgs, err := i.client.ListImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing images: %w", err)
	}

	names := make([]string, len(imgs))
	for idx, img := range imgs {
		names[idx] = img.Name()
	}

	return names, nil
}
