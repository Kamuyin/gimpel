package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gimpel/pkg/storage"

	log "github.com/sirupsen/logrus"
)

func (s *Store) AddModule(mod *Module) error {
	if mod.CreatedAt.IsZero() {
		mod.CreatedAt = time.Now()
	}
	mod.UpdatedAt = time.Now()
	key := ModuleKey(mod.ID, mod.Version)
	return s.db.PutJSON(BucketModules, key, mod)
}

func (s *Store) GetModule(id, version string) (*Module, error) {
	key := ModuleKey(id, version)
	var mod Module
	if err := s.db.GetJSON(BucketModules, key, &mod); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &mod, nil
}

func (s *Store) GetLatestModule(id string) (*Module, error) {
	var latest *Module
	err := s.db.ForEach(BucketModules, func(_, value []byte) error {
		var mod Module
		if err := unmarshalJSON(value, &mod); err != nil {
			return err
		}
		if mod.ID == id {
			if latest == nil || mod.CreatedAt.After(latest.CreatedAt) {
				latest = &mod
			}
		}
		return nil
	})
	return latest, err
}

func (s *Store) ListModules() ([]*Module, error) {
	var modules []*Module
	err := s.db.ForEach(BucketModules, func(_, value []byte) error {
		var mod Module
		if err := unmarshalJSON(value, &mod); err != nil {
			return err
		}
		modules = append(modules, &mod)
		return nil
	})
	return modules, err
}

func (s *Store) ListModuleVersions(id string) ([]*Module, error) {
	var versions []*Module
	err := s.db.ForEach(BucketModules, func(_, value []byte) error {
		var mod Module
		if err := unmarshalJSON(value, &mod); err != nil {
			return err
		}
		if mod.ID == id {
			versions = append(versions, &mod)
		}
		return nil
	})
	return versions, err
}

func (s *Store) DeleteModule(id, version string) error {
	key := ModuleKey(id, version)
	if err := s.DeleteImage(id, version); err != nil {
		log.WithError(err).Warn("failed to delete module image")
	}
	return s.db.Delete(BucketModules, key)
}

func (s *Store) StoreImage(moduleID, version string, reader io.Reader) (*ImageMeta, error) {
	if err := os.MkdirAll(s.imageDir, 0755); err != nil {
		return nil, fmt.Errorf("creating image directory: %w", err)
	}

	filename := fmt.Sprintf("%s_%s.tar", moduleID, version)
	path := filepath.Join(s.imageDir, filename)
	tmpPath := path + ".tmp"

	file, err := os.Create(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}

	hash := sha256.New()
	writer := io.MultiWriter(file, hash)

	size, err := io.Copy(writer, reader)
	if err != nil {
		file.Close()
		os.Remove(tmpPath)
		return nil, fmt.Errorf("writing image: %w", err)
	}
	file.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("renaming image file: %w", err)
	}

	digest := "sha256:" + hex.EncodeToString(hash.Sum(nil))

	meta := &ImageMeta{
		ModuleID:  moduleID,
		Version:   version,
		Digest:    digest,
		Path:      path,
		SizeBytes: size,
		StoredAt:  time.Now(),
	}

	key := ModuleKey(moduleID, version)
	if err := s.db.PutJSON(BucketImages, key, meta); err != nil {
		os.Remove(path)
		return nil, fmt.Errorf("storing image metadata: %w", err)
	}

	log.WithFields(log.Fields{
		"module":  moduleID,
		"version": version,
		"digest":  digest,
		"size":    size,
	}).Info("image stored")

	return meta, nil
}

func (s *Store) GetImage(moduleID, version string) (*ImageMeta, error) {
	key := ModuleKey(moduleID, version)
	var meta ImageMeta
	if err := s.db.GetJSON(BucketImages, key, &meta); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &meta, nil
}

func (s *Store) OpenImage(moduleID, version string) (io.ReadCloser, int64, error) {
	meta, err := s.GetImage(moduleID, version)
	if err != nil {
		return nil, 0, err
	}
	if meta == nil {
		return nil, 0, storage.ErrNotFound
	}

	file, err := os.Open(meta.Path)
	if err != nil {
		return nil, 0, fmt.Errorf("opening image file: %w", err)
	}

	return file, meta.SizeBytes, nil
}

func (s *Store) DeleteImage(moduleID, version string) error {
	key := ModuleKey(moduleID, version)
	meta, err := s.GetImage(moduleID, version)
	if err != nil {
		return err
	}
	if meta != nil && meta.Path != "" {
		os.Remove(meta.Path)
	}
	return s.db.Delete(BucketImages, key)
}
