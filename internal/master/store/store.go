package store

import (
	"fmt"
	"time"

	"gimpel/pkg/storage"

	log "github.com/sirupsen/logrus"
)

const (
	BucketSatellites    = "satellites"
	BucketModules       = "modules"
	BucketImages        = "images"
	BucketDeployments   = "deployments"
	BucketSessions      = "sessions"
	BucketEvents        = "events"
	BucketSettings      = "settings"
	BucketPairings      = "pairings"
	BucketPairingTokens = "pairing_tokens"
)

type Store struct {
	db       *storage.DB
	imageDir string
}

type Config struct {
	DBPath   string
	ImageDir string
}

func New(cfg *Config) (*Store, error) {
	opts := storage.DefaultOptions(cfg.DBPath)
	opts.InitBuckets = []string{
		BucketSatellites,
		BucketModules,
		BucketImages,
		BucketDeployments,
		BucketSessions,
		BucketEvents,
		BucketSettings,
		BucketPairings,
		BucketPairingTokens,
	}

	db, err := storage.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	log.WithField("path", cfg.DBPath).Info("master store opened")

	return &Store{
		db:       db,
		imageDir: cfg.ImageDir,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *storage.DB {
	return s.db
}

type Satellite struct {
	ID           string            `json:"id"`
	Hostname     string            `json:"hostname"`
	IPAddress    string            `json:"ip_address"`
	Version      string            `json:"version"`
	OS           string            `json:"os"`
	Arch         string            `json:"arch"`
	Labels       map[string]string `json:"labels"`
	Status       SatelliteStatus   `json:"status"`
	RegisteredAt time.Time         `json:"registered_at"`
	LastSeenAt   time.Time         `json:"last_seen_at"`
	CertSerial   string            `json:"cert_serial"`
}

type SatelliteStatus string

const (
	SatelliteStatusOnline      SatelliteStatus = "online"
	SatelliteStatusOffline     SatelliteStatus = "offline"
	SatelliteStatusUnreachable SatelliteStatus = "unreachable"
	SatelliteStatusPending     SatelliteStatus = "pending"
)

type Module struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Digest      string            `json:"digest"`
	Protocol    string            `json:"protocol"`
	ImageRef    string            `json:"image_ref"`
	SizeBytes   int64             `json:"size_bytes"`
	Labels      map[string]string `json:"labels"`
	Signature   []byte            `json:"signature"`
	SignedBy    string            `json:"signed_by"`
	SignedAt    time.Time         `json:"signed_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func ModuleKey(id, version string) string {
	return fmt.Sprintf("%s:%s", id, version)
}

type ImageMeta struct {
	ModuleID  string    `json:"module_id"`
	Version   string    `json:"version"`
	Digest    string    `json:"digest"`
	Path      string    `json:"path"`
	SizeBytes int64     `json:"size_bytes"`
	StoredAt  time.Time `json:"stored_at"`
}

type Deployment struct {
	SatelliteID string             `json:"satellite_id"`
	Modules     []ModuleDeployment `json:"modules"`
	Version     int64              `json:"version"`
	Signature   []byte             `json:"signature"`
	SignedBy    string             `json:"signed_by"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type ModuleDeployment struct {
	ModuleID      string            `json:"module_id"`
	ModuleVersion string            `json:"module_version"`
	Enabled       bool              `json:"enabled"`
	ExecutionMode string            `json:"execution_mode"`
	Listeners     []ListenerConfig  `json:"listeners"`
	Env           map[string]string `json:"env"`
	Resources     ResourceConfig    `json:"resources,omitempty"`
}

type ListenerConfig struct {
	ID              string `json:"id"`
	Protocol        string `json:"protocol"`
	Port            uint32 `json:"port"`
	HighInteraction bool   `json:"high_interaction"`
}

type ResourceConfig struct {
	MaxMemoryMB   int64 `json:"max_memory_mb,omitempty"`
	MaxCPUPercent int   `json:"max_cpu_percent,omitempty"`
}
