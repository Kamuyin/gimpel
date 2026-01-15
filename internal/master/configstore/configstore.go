package configstore

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"

	gimpelv1 "gimpel/api/go/v1"
)

type ConfigStore interface {
	GetConfig(agentID string) (*gimpelv1.AgentConfig, string, bool)
	SetConfig(agentID string, config *gimpelv1.AgentConfig) string
	SetDefaultConfig(config *gimpelv1.AgentConfig) string
	GetDefaultConfig() (*gimpelv1.AgentConfig, string)
}

type InMemoryConfigStore struct {
	mu             sync.RWMutex
	defaultConfig  *gimpelv1.AgentConfig
	defaultVersion string
	agentConfigs   map[string]*agentConfigEntry
}

type agentConfigEntry struct {
	config  *gimpelv1.AgentConfig
	version string
}

func NewInMemoryConfigStore() *InMemoryConfigStore {
	return &InMemoryConfigStore{
		agentConfigs: make(map[string]*agentConfigEntry),
	}
}

func (s *InMemoryConfigStore) GetConfig(agentID string) (*gimpelv1.AgentConfig, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if entry, ok := s.agentConfigs[agentID]; ok {
		return entry.config, entry.version, true
	}

	if s.defaultConfig != nil {
		return s.defaultConfig, s.defaultVersion, true
	}

	return nil, "", false
}

func (s *InMemoryConfigStore) SetConfig(agentID string, config *gimpelv1.AgentConfig) string {
	version := computeVersion(config)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.agentConfigs[agentID] = &agentConfigEntry{
		config:  config,
		version: version,
	}

	return version
}

func (s *InMemoryConfigStore) SetDefaultConfig(config *gimpelv1.AgentConfig) string {
	version := computeVersion(config)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.defaultConfig = config
	s.defaultVersion = version

	return version
}

func (s *InMemoryConfigStore) GetDefaultConfig() (*gimpelv1.AgentConfig, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.defaultConfig, s.defaultVersion
}

func computeVersion(config *gimpelv1.AgentConfig) string {
	if config == nil {
		return ""
	}
	hash := sha256.New()
	hash.Write([]byte(config.String()))
	return hex.EncodeToString(hash.Sum(nil))[:12]
}
