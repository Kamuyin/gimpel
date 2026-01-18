// Package modulestore provides storage and management for signed module images.
package modulestore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/pkg/signing"

	log "github.com/sirupsen/logrus"
)

type Store interface {
	AddModule(module *gimpelv1.ModuleImage, imageData []byte) error
	GetModule(id, version string) (*gimpelv1.ModuleImage, error)
	GetModuleImage(id, version string) (io.ReadCloser, int64, error)
	ListModules() ([]*gimpelv1.ModuleImage, error)
	DeleteModule(id, version string) error

	GetCatalog() (*gimpelv1.ModuleCatalog, error)
	GetCatalogVersion() int64

	SetAgentAssignments(agentID string, config *gimpelv1.AgentModuleConfig) error
	GetAgentAssignments(agentID string) (*gimpelv1.AgentModuleConfig, error)
}

type FileStore struct {
	mu sync.RWMutex

	baseDir    string
	signer     *signing.ModuleSigner
	catalog    *gimpelv1.ModuleCatalog
	catalogVer int64

	assignments map[string]*gimpelv1.AgentModuleConfig
}

func NewFileStore(baseDir string, signer *signing.ModuleSigner) (*FileStore, error) {
	dirs := []string{
		filepath.Join(baseDir, "modules"),
		filepath.Join(baseDir, "images"),
		filepath.Join(baseDir, "assignments"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	store := &FileStore{
		baseDir:     baseDir,
		signer:      signer,
		assignments: make(map[string]*gimpelv1.AgentModuleConfig),
		catalog: &gimpelv1.ModuleCatalog{
			Version:   1,
			UpdatedAt: time.Now().Unix(),
		},
	}

	if err := store.loadModules(); err != nil {
		log.WithError(err).Warn("failed to load existing modules")
	}

	if err := store.loadAssignments(); err != nil {
		log.WithError(err).Warn("failed to load existing assignments")
	}

	return store, nil
}

func (s *FileStore) AddModule(module *gimpelv1.ModuleImage, imageData []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if module.Digest == "" {
		module.Digest = signing.ComputeImageDigest(imageData)
	}

	expectedDigest := signing.ComputeImageDigest(imageData)
	if module.Digest != expectedDigest {
		return fmt.Errorf("digest mismatch: expected %s, got %s", expectedDigest, module.Digest)
	}

	module.SizeBytes = int64(len(imageData))

	if s.signer != nil {
		if err := s.signer.SignModule(module); err != nil {
			return fmt.Errorf("signing module: %w", err)
		}
	}

	metadataPath := s.moduleMetadataPath(module.Id, module.Version)
	if err := s.saveJSON(metadataPath, module); err != nil {
		return fmt.Errorf("saving module metadata: %w", err)
	}

	imagePath := s.moduleImagePath(module.Id, module.Version)
	if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
		os.Remove(metadataPath)
		return fmt.Errorf("saving module image: %w", err)
	}

	s.updateCatalogLocked()

	log.WithFields(log.Fields{
		"module":  module.Id,
		"version": module.Version,
		"size":    module.SizeBytes,
		"digest":  module.Digest[:20] + "...",
	}).Info("module added to store")

	return nil
}

func (s *FileStore) GetModule(id, version string) (*gimpelv1.ModuleImage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if version == "latest" || version == "" {
		return s.getLatestModuleLocked(id)
	}

	metadataPath := s.moduleMetadataPath(id, version)
	var module gimpelv1.ModuleImage
	if err := s.loadJSON(metadataPath, &module); err != nil {
		return nil, fmt.Errorf("loading module %s:%s: %w", id, version, err)
	}

	return &module, nil
}

func (s *FileStore) getLatestModuleLocked(id string) (*gimpelv1.ModuleImage, error) {
	pattern := filepath.Join(s.baseDir, "modules", id, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("listing versions: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("module %s not found", id)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(matches)))

	var module gimpelv1.ModuleImage
	if err := s.loadJSON(matches[0], &module); err != nil {
		return nil, fmt.Errorf("loading latest module: %w", err)
	}

	return &module, nil
}

func (s *FileStore) GetModuleImage(id, version string) (io.ReadCloser, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if version == "latest" || version == "" {
		module, err := s.getLatestModuleLocked(id)
		if err != nil {
			return nil, 0, err
		}
		version = module.Version
	}

	imagePath := s.moduleImagePath(id, version)
	info, err := os.Stat(imagePath)
	if err != nil {
		return nil, 0, fmt.Errorf("module image not found: %w", err)
	}

	file, err := os.Open(imagePath)
	if err != nil {
		return nil, 0, fmt.Errorf("opening module image: %w", err)
	}

	return file, info.Size(), nil
}

func (s *FileStore) ListModules() ([]*gimpelv1.ModuleImage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.catalog.Modules, nil
}

func (s *FileStore) DeleteModule(id, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	metadataPath := s.moduleMetadataPath(id, version)
	imagePath := s.moduleImagePath(id, version)

	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing metadata: %w", err)
	}

	if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing image: %w", err)
	}

	s.updateCatalogLocked()

	log.WithFields(log.Fields{
		"module":  id,
		"version": version,
	}).Info("module deleted from store")

	return nil
}

func (s *FileStore) GetCatalog() (*gimpelv1.ModuleCatalog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.catalog, nil
}

func (s *FileStore) GetCatalogVersion() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.catalogVer
}

func (s *FileStore) SetAgentAssignments(agentID string, config *gimpelv1.AgentModuleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config.AgentId = agentID
	config.Version = time.Now().Unix()

	if s.signer != nil {
		if err := s.signer.SignAgentConfig(config); err != nil {
			return fmt.Errorf("signing agent config: %w", err)
		}
	}

	path := filepath.Join(s.baseDir, "assignments", agentID+".json")
	if err := s.saveJSON(path, config); err != nil {
		return fmt.Errorf("saving assignment: %w", err)
	}

	s.assignments[agentID] = config

	log.WithFields(log.Fields{
		"agent":       agentID,
		"assignments": len(config.Assignments),
	}).Info("agent assignments updated")

	return nil
}

func (s *FileStore) GetAgentAssignments(agentID string) (*gimpelv1.AgentModuleConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, ok := s.assignments[agentID]
	if !ok {
		return nil, fmt.Errorf("no assignments for agent %s", agentID)
	}

	return config, nil
}

func (s *FileStore) moduleMetadataPath(id, version string) string {
	return filepath.Join(s.baseDir, "modules", id, version+".json")
}

func (s *FileStore) moduleImagePath(id, version string) string {
	return filepath.Join(s.baseDir, "images", id+"-"+version+".tar")
}

func (s *FileStore) saveJSON(path string, v interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *FileStore) loadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (s *FileStore) loadModules() error {
	modulesDir := filepath.Join(s.baseDir, "modules")
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		moduleID := entry.Name()
		pattern := filepath.Join(modulesDir, moduleID, "*.json")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, match := range matches {
			var module gimpelv1.ModuleImage
			if err := s.loadJSON(match, &module); err != nil {
				log.WithError(err).WithField("path", match).Warn("failed to load module")
				continue
			}
			s.catalog.Modules = append(s.catalog.Modules, &module)
		}
	}

	s.updateCatalogLocked()
	return nil
}

func (s *FileStore) loadAssignments() error {
	assignmentsDir := filepath.Join(s.baseDir, "assignments")
	entries, err := os.ReadDir(assignmentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(assignmentsDir, entry.Name())
		var config gimpelv1.AgentModuleConfig
		if err := s.loadJSON(path, &config); err != nil {
			log.WithError(err).WithField("path", path).Warn("failed to load assignment")
			continue
		}

		agentID := entry.Name()[:len(entry.Name())-5]
		s.assignments[agentID] = &config
	}

	return nil
}

func (s *FileStore) updateCatalogLocked() {
	s.catalogVer++
	s.catalog.Version = s.catalogVer
	s.catalog.UpdatedAt = time.Now().Unix()

	modules := make([]*gimpelv1.ModuleImage, 0)
	modulesDir := filepath.Join(s.baseDir, "modules")

	entries, err := os.ReadDir(modulesDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			pattern := filepath.Join(modulesDir, entry.Name(), "*.json")
			matches, _ := filepath.Glob(pattern)

			for _, match := range matches {
				var module gimpelv1.ModuleImage
				if err := s.loadJSON(match, &module); err == nil {
					modules = append(modules, &module)
				}
			}
		}
	}

	s.catalog.Modules = modules

	if s.signer != nil {
		if err := s.signer.SignCatalog(s.catalog); err != nil {
			log.WithError(err).Error("failed to sign catalog")
		}
	}
}

func DigestFromReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}
