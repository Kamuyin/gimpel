// Package moduleclient provides client-side functionality for fetching
package moduleclient

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/pkg/signing"
)

type Client struct {
	mu sync.RWMutex

	conn     *grpc.ClientConn
	client   gimpelv1.ModuleCatalogServiceClient
	verifier *signing.ModuleVerifier
	cacheDir string

	catalog    *gimpelv1.ModuleCatalog
	catalogVer int64
	configVer  int64
}

type Config struct {
	MasterAddress  string
	TrustedKeyPath string
	CacheDir       string
	DialOptions    []grpc.DialOption
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := grpc.Dial(cfg.MasterAddress, cfg.DialOptions...)
	if err != nil {
		return nil, fmt.Errorf("connecting to master: %w", err)
	}

	client := gimpelv1.NewModuleCatalogServiceClient(conn)

	var verifier *signing.ModuleVerifier
	if cfg.TrustedKeyPath != "" {
		keyPair, err := signing.LoadPublicKey(cfg.TrustedKeyPath)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("loading trusted key: %w", err)
		}
		verifier = signing.NewModuleVerifier(keyPair)
		log.WithField("key_id", keyPair.KeyID).Info("loaded trusted signing key")
	} else {
		log.Warn("no trusted key configured, module signatures will not be verified")
		verifier = signing.NewModuleVerifier()
	}

	cacheDir := cfg.CacheDir
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), "gimpel-modules")
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		conn.Close()
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	return &Client{
		conn:     conn,
		client:   client,
		verifier: verifier,
		cacheDir: cacheDir,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) FetchCatalog(ctx context.Context) (*gimpelv1.ModuleCatalog, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := c.client.GetCatalog(ctx, &gimpelv1.GetCatalogRequest{
		CurrentVersion: c.catalogVer,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching catalog: %w", err)
	}

	if !resp.Updated {
		return c.catalog, nil
	}

	catalog := resp.Catalog

	if err := c.verifier.VerifyCatalog(catalog); err != nil {
		return nil, fmt.Errorf("catalog signature verification failed: %w", err)
	}

	for _, module := range catalog.Modules {
		if err := c.verifier.VerifyModule(module); err != nil {
			log.WithFields(log.Fields{
				"module":  module.Id,
				"version": module.Version,
				"error":   err,
			}).Warn("module signature verification failed, skipping")
			continue
		}
	}

	c.catalog = catalog
	c.catalogVer = catalog.Version

	log.WithFields(log.Fields{
		"version":      catalog.Version,
		"module_count": len(catalog.Modules),
	}).Info("catalog updated")

	return catalog, nil
}

func (c *Client) FetchAssignments(ctx context.Context, agentID string) (*gimpelv1.AgentModuleConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	resp, err := c.client.GetModuleAssignments(ctx, &gimpelv1.GetModuleAssignmentsRequest{
		AgentId:        agentID,
		CurrentVersion: c.configVer,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching assignments: %w", err)
	}

	if !resp.Updated {
		return nil, nil
	}

	config := resp.Config

	if err := c.verifier.VerifyAgentConfig(config); err != nil {
		return nil, fmt.Errorf("config signature verification failed: %w", err)
	}

	c.configVer = config.Version

	log.WithFields(log.Fields{
		"agent_id":    agentID,
		"version":     config.Version,
		"assignments": len(config.Assignments),
	}).Info("assignments updated")

	return config, nil
}

func (c *Client) DownloadModule(ctx context.Context, moduleID, version string) (string, error) {
	cachedPath := c.cachedImagePath(moduleID, version)
	if _, err := os.Stat(cachedPath); err == nil {
		if c.verifyCachedImage(ctx, moduleID, version, cachedPath) {
			log.WithFields(log.Fields{
				"module":  moduleID,
				"version": version,
				"path":    cachedPath,
			}).Debug("using cached module image")
			return cachedPath, nil
		}
		os.Remove(cachedPath)
	}

	log.WithFields(log.Fields{
		"module":  moduleID,
		"version": version,
	}).Info("downloading module from master")

	stream, err := c.client.DownloadModule(ctx, &gimpelv1.DownloadModuleRequest{
		ModuleId: moduleID,
		Version:  version,
	})
	if err != nil {
		return "", fmt.Errorf("starting download: %w", err)
	}

	tmpPath := cachedPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpPath)

	hash := sha256.New()
	writer := io.MultiWriter(file, hash)

	var bytesReceived int64
	startTime := time.Now()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			file.Close()
			return "", fmt.Errorf("receiving chunk: %w", err)
		}

		if _, err := writer.Write(chunk.Data); err != nil {
			file.Close()
			return "", fmt.Errorf("writing chunk: %w", err)
		}

		bytesReceived += int64(len(chunk.Data))
	}

	file.Close()

	digest := "sha256:" + hex.EncodeToString(hash.Sum(nil))

	verifyResp, err := c.client.VerifyModule(ctx, &gimpelv1.VerifyModuleRequest{
		ModuleId: moduleID,
		Version:  version,
		Digest:   digest,
	})
	if err != nil {
		return "", fmt.Errorf("verifying module: %w", err)
	}

	if !verifyResp.Valid {
		return "", fmt.Errorf("module digest mismatch: possible tampering detected")
	}

	if err := os.MkdirAll(filepath.Dir(cachedPath), 0755); err != nil {
		return "", fmt.Errorf("creating cache directory: %w", err)
	}
	if err := os.Rename(tmpPath, cachedPath); err != nil {
		return "", fmt.Errorf("moving to cache: %w", err)
	}

	duration := time.Since(startTime)
	log.WithFields(log.Fields{
		"module":   moduleID,
		"version":  version,
		"size":     bytesReceived,
		"duration": duration,
		"rate_mb":  float64(bytesReceived) / duration.Seconds() / 1024 / 1024,
	}).Info("module downloaded and verified")

	return cachedPath, nil
}

func (c *Client) cachedImagePath(moduleID, version string) string {
	return filepath.Join(c.cacheDir, moduleID, version+".tar")
}

func (c *Client) verifyCachedImage(ctx context.Context, moduleID, version, path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false
	}

	digest := "sha256:" + hex.EncodeToString(hash.Sum(nil))

	resp, err := c.client.VerifyModule(ctx, &gimpelv1.VerifyModuleRequest{
		ModuleId: moduleID,
		Version:  version,
		Digest:   digest,
	})
	if err != nil {
		return false
	}

	return resp.Valid
}

func (c *Client) GetModule(moduleID, version string) (*gimpelv1.ModuleImage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.catalog == nil {
		return nil, false
	}

	for _, module := range c.catalog.Modules {
		if module.Id == moduleID {
			if version == "" || version == "latest" || module.Version == version {
				return module, true
			}
		}
	}

	return nil, false
}

func (c *Client) ListModules() []*gimpelv1.ModuleImage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.catalog == nil {
		return nil
	}

	return c.catalog.Modules
}
