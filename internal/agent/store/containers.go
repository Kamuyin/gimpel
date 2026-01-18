package store

import (
	"time"

	"gimpel/pkg/storage"
)

func (s *Store) SaveContainer(c *Container) error {
	return s.db.PutJSON(BucketContainers, c.ID, c)
}

func (s *Store) GetContainer(id string) (*Container, error) {
	var c Container
	if err := s.db.GetJSON(BucketContainers, id, &c); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *Store) ListContainers() ([]*Container, error) {
	var containers []*Container
	err := s.db.ForEach(BucketContainers, func(_, value []byte) error {
		var c Container
		if err := unmarshalJSON(value, &c); err != nil {
			return err
		}
		containers = append(containers, &c)
		return nil
	})
	return containers, err
}

func (s *Store) GetContainersByModule(moduleID string) ([]*Container, error) {
	all, err := s.ListContainers()
	if err != nil {
		return nil, err
	}
	var result []*Container
	for _, c := range all {
		if c.ModuleID == moduleID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (s *Store) GetRunningContainers() ([]*Container, error) {
	all, err := s.ListContainers()
	if err != nil {
		return nil, err
	}
	var running []*Container
	for _, c := range all {
		if c.Status == ContainerStatusRunning {
			running = append(running, c)
		}
	}
	return running, nil
}

func (s *Store) UpdateContainerStatus(id string, status ContainerStatus, err string) error {
	c, getErr := s.GetContainer(id)
	if getErr != nil {
		return getErr
	}
	if c == nil {
		return storage.ErrNotFound
	}
	c.Status = status
	c.LastError = err
	if status == ContainerStatusStopped || status == ContainerStatusFailed {
		c.StoppedAt = time.Now()
	}
	return s.SaveContainer(c)
}

func (s *Store) DeleteContainer(id string) error {
	return s.db.Delete(BucketContainers, id)
}

func (s *Store) IncrementRestartCount(id string) error {
	c, err := s.GetContainer(id)
	if err != nil {
		return err
	}
	if c == nil {
		return storage.ErrNotFound
	}
	c.RestartCount++
	c.Status = ContainerStatusRestarting
	return s.SaveContainer(c)
}

func (s *Store) CleanupStoppedContainers(olderThan time.Duration) error {
	all, err := s.ListContainers()
	if err != nil {
		return err
	}
	cutoff := time.Now().Add(-olderThan)
	for _, c := range all {
		if c.Status == ContainerStatusStopped && !c.StoppedAt.IsZero() && c.StoppedAt.Before(cutoff) {
			s.DeleteContainer(c.ID)
		}
	}
	return nil
}
