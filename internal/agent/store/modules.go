package store

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gimpel/pkg/storage"
)

func ModuleKey(id, version string) string {
	return fmt.Sprintf("%s:%s", id, version)
}

type HighWaterMark struct {
	ModuleID  string    `json:"module_id"`
	Version   string    `json:"version"`
	Timestamp int64     `json:"timestamp"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Store) GetModuleHighWaterMark(moduleID string) (*HighWaterMark, error) {
	var hwm HighWaterMark
	if err := s.db.GetJSON(BucketHighWaterMarks, moduleID, &hwm); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &hwm, nil
}

func (s *Store) SetModuleHighWaterMark(hwm *HighWaterMark) error {
	hwm.UpdatedAt = time.Now()
	return s.db.PutJSON(BucketHighWaterMarks, hwm.ModuleID, hwm)
}

func (s *Store) SaveModuleCache(mod *ModuleCache) error {
	key := ModuleKey(mod.ModuleID, mod.Version)
	return s.db.PutJSON(BucketModules, key, mod)
}

func (s *Store) GetModuleCache(moduleID, version string) (*ModuleCache, error) {
	key := ModuleKey(moduleID, version)
	var mod ModuleCache
	if err := s.db.GetJSON(BucketModules, key, &mod); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &mod, nil
}

func (s *Store) HasModule(moduleID, version string) (bool, error) {
	mod, err := s.GetModuleCache(moduleID, version)
	if err != nil {
		return false, err
	}
	if mod == nil {
		return false, nil
	}
	if _, err := os.Stat(mod.ImagePath); os.IsNotExist(err) {
		s.DeleteModuleCache(moduleID, version)
		return false, nil
	}
	return true, nil
}

func (s *Store) ListModuleCache() ([]*ModuleCache, error) {
	var modules []*ModuleCache
	err := s.db.ForEach(BucketModules, func(_, value []byte) error {
		var mod ModuleCache
		if err := unmarshalJSON(value, &mod); err != nil {
			return err
		}
		modules = append(modules, &mod)
		return nil
	})
	return modules, err
}

func (s *Store) DeleteModuleCache(moduleID, version string) error {
	mod, err := s.GetModuleCache(moduleID, version)
	if err != nil {
		return err
	}
	if mod != nil && mod.ImagePath != "" {
		os.Remove(mod.ImagePath)
	}
	key := ModuleKey(moduleID, version)
	return s.db.Delete(BucketModules, key)
}

func (s *Store) GetModuleImagePath(moduleID, version string) string {
	return filepath.Join(s.cacheDir, moduleID, version+".tar")
}

func (s *Store) CleanupOrphanedImages() error {
	cached, err := s.ListModuleCache()
	if err != nil {
		return err
	}

	validPaths := make(map[string]bool)
	for _, mod := range cached {
		validPaths[mod.ImagePath] = true
	}

	return filepath.Walk(s.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !validPaths[path] {
			os.Remove(path)
		}
		return nil
	})
}

func (s *Store) GetCacheSize() (int64, error) {
	modules, err := s.ListModuleCache()
	if err != nil {
		return 0, err
	}
	var total int64
	for _, mod := range modules {
		total += mod.SizeBytes
	}
	return total, nil
}

func (s *Store) MarkModuleVerified(moduleID, version string) error {
	mod, err := s.GetModuleCache(moduleID, version)
	if err != nil {
		return err
	}
	if mod == nil {
		return storage.ErrNotFound
	}
	mod.Verified = true
	return s.SaveModuleCache(mod)
}

func (s *Store) GetUnverifiedModules() ([]*ModuleCache, error) {
	all, err := s.ListModuleCache()
	if err != nil {
		return nil, err
	}
	var unverified []*ModuleCache
	for _, mod := range all {
		if !mod.Verified {
			unverified = append(unverified, mod)
		}
	}
	return unverified, nil
}
