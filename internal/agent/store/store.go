// Package store provides the agent's persistent storage layer using bbolt.
package store

import (
	"fmt"
	"time"

	"gimpel/pkg/storage"

	log "github.com/sirupsen/logrus"
)

const (
	BucketModules    = "modules"
	BucketContainers = "containers"
	BucketEvents     = "events"
	BucketConfig     = "config"
	BucketState      = "state"
)

type Store struct {
	db       *storage.DB
	cacheDir string
}

type Config struct {
	DBPath   string
	CacheDir string
}

func New(cfg *Config) (*Store, error) {
	opts := storage.DefaultOptions(cfg.DBPath)
	opts.InitBuckets = []string{
		BucketModules,
		BucketContainers,
		BucketEvents,
		BucketConfig,
		BucketState,
	}

	db, err := storage.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	log.WithField("path", cfg.DBPath).Info("agent store opened")

	return &Store{
		db:       db,
		cacheDir: cfg.CacheDir,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *storage.DB {
	return s.db
}

func (s *Store) CacheDir() string {
	return s.cacheDir
}

type ModuleCache struct {
	ModuleID     string    `json:"module_id"`
	Version      string    `json:"version"`
	Digest       string    `json:"digest"`
	ImagePath    string    `json:"image_path"`
	SizeBytes    int64     `json:"size_bytes"`
	Signature    []byte    `json:"signature"`
	SignedBy     string    `json:"signed_by"`
	DownloadedAt time.Time `json:"downloaded_at"`
	Verified     bool      `json:"verified"`
}

type Container struct {
	ID            string            `json:"id"`
	ModuleID      string            `json:"module_id"`
	ModuleVersion string            `json:"module_version"`
	ImageRef      string            `json:"image_ref"`
	Status        ContainerStatus   `json:"status"`
	PID           int               `json:"pid,omitempty"`
	Listeners     []ListenerState   `json:"listeners"`
	Env           map[string]string `json:"env"`
	StartedAt     time.Time         `json:"started_at"`
	StoppedAt     time.Time         `json:"stopped_at,omitempty"`
	RestartCount  int               `json:"restart_count"`
	LastError     string            `json:"last_error,omitempty"`
}

type ContainerStatus string

const (
	ContainerStatusPending   ContainerStatus = "pending"
	ContainerStatusStarting  ContainerStatus = "starting"
	ContainerStatusRunning   ContainerStatus = "running"
	ContainerStatusStopping  ContainerStatus = "stopping"
	ContainerStatusStopped   ContainerStatus = "stopped"
	ContainerStatusFailed    ContainerStatus = "failed"
	ContainerStatusRestarting ContainerStatus = "restarting"
)

type ListenerState struct {
	ID       string `json:"id"`
	Protocol string `json:"protocol"`
	Port     uint32 `json:"port"`
	Bound    bool   `json:"bound"`
}

type AgentState struct {
	AgentID        string    `json:"agent_id"`
	Registered     bool      `json:"registered"`
	CertPath       string    `json:"cert_path"`
	KeyPath        string    `json:"key_path"`
	CatalogVersion int64     `json:"catalog_version"`
	ConfigVersion  int64     `json:"config_version"`
	RegisteredAt   time.Time `json:"registered_at"`
	LastSyncAt     time.Time `json:"last_sync_at"`
}

type DeploymentConfig struct {
	Version     int64              `json:"version"`
	Modules     []ModuleDeployment `json:"modules"`
	Signature   []byte             `json:"signature"`
	SignedBy    string             `json:"signed_by"`
	ReceivedAt  time.Time          `json:"received_at"`
}

type ModuleDeployment struct {
	ModuleID      string            `json:"module_id"`
	ModuleVersion string            `json:"module_version"`
	Enabled       bool              `json:"enabled"`
	ExecutionMode string            `json:"execution_mode"`
	Listeners     []ListenerConfig  `json:"listeners"`
	Env           map[string]string `json:"env"`
}

type ListenerConfig struct {
	ID              string `json:"id"`
	Protocol        string `json:"protocol"`
	Port            uint32 `json:"port"`
	HighInteraction bool   `json:"high_interaction"`
}
