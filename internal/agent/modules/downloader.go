package modules

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/agent/store"
	"gimpel/pkg/signing"
)

type ModuleDownloader struct {
	catalogClient gimpelv1.ModuleCatalogServiceClient
	store         Store
	cacheDir      string
	verifier      *signing.ModuleVerifier
}

func NewModuleDownloader(client gimpelv1.ModuleCatalogServiceClient, s Store, cacheDir string, verifier *signing.ModuleVerifier) *ModuleDownloader {
	return &ModuleDownloader{
		catalogClient: client,
		store:         s,
		cacheDir:      cacheDir,
		verifier:      verifier,
	}
}

func (md *ModuleDownloader) DownloadModule(ctx context.Context, moduleID, version string) (*store.ModuleCache, error) {
	cached, err := md.store.GetModuleCache(moduleID, version)
	if err == nil && cached != nil && cached.Verified {
		log.WithFields(log.Fields{
			"module":  moduleID,
			"version": version,
		}).Debug("module already in cache")
		return cached, nil
	}

	log.WithFields(log.Fields{
		"module":  moduleID,
		"version": version,
	}).Info("downloading module")

	if err := os.MkdirAll(md.cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	tempPath := filepath.Join(md.cacheDir, fmt.Sprintf("%s_%s.tar.tmp", moduleID, version))
	finalPath := filepath.Join(md.cacheDir, fmt.Sprintf("%s_%s.tar", moduleID, version))

	file, err := os.Create(tempPath)
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	defer file.Close()

	stream, err := md.catalogClient.DownloadModule(ctx, &gimpelv1.DownloadModuleRequest{
		ModuleId: moduleID,
		Version:  version,
	})
	if err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("starting download: %w", err)
	}

	hash := sha256.New()
	writer := io.MultiWriter(file, hash)
	var totalSize int64

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tempPath)
			return nil, fmt.Errorf("receiving chunk: %w", err)
		}

		n, err := writer.Write(chunk.Data)
		if err != nil {
			os.Remove(tempPath)
			return nil, fmt.Errorf("writing chunk: %w", err)
		}
		totalSize += int64(n)
	}

	file.Close()

	digest := "sha256:" + hex.EncodeToString(hash.Sum(nil))

	log.WithFields(log.Fields{
		"module":  moduleID,
		"version": version,
		"digest":  digest,
		"size":    totalSize,
	}).Info("module downloaded, verifying")

	verifyResp, err := md.catalogClient.VerifyModule(ctx, &gimpelv1.VerifyModuleRequest{
		ModuleId: moduleID,
		Version:  version,
		Digest:   digest,
	})
	if err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("verifying digest: %w", err)
	}

	if !verifyResp.Valid {
		os.Remove(tempPath)
		return nil, fmt.Errorf("digest verification failed")
	}

	moduleImage := &gimpelv1.ModuleImage{
		Id:        moduleID,
		Version:   version,
		Digest:    digest,
		Signature: verifyResp.Signature,
	}

	if err := md.verifier.VerifyModule(moduleImage); err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	if err := os.Rename(tempPath, finalPath); err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("moving to final location: %w", err)
	}

	cache := &store.ModuleCache{
		ModuleID:     moduleID,
		Version:      version,
		Digest:       digest,
		ImagePath:    finalPath,
		SizeBytes:    totalSize,
		Signature:    verifyResp.Signature,
		SignedBy:     moduleImage.SignedBy,
		DownloadedAt: time.Now(),
		Verified:     true,
	}

	if err := md.store.SaveModuleCache(cache); err != nil {
		return nil, fmt.Errorf("saving to cache: %w", err)
	}

	log.WithFields(log.Fields{
		"module":    moduleID,
		"version":   version,
		"signed_by": cache.SignedBy,
	}).Info("module verified and cached")

	return cache, nil
}

func (md *ModuleDownloader) GetCachedModule(moduleID, version string) (*store.ModuleCache, error) {
	return md.store.GetModuleCache(moduleID, version)
}
