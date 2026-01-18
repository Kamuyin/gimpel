package store

import (
	"time"

	"gimpel/pkg/storage"
)

func (s *Store) SetDeployment(dep *Deployment) error {
	dep.UpdatedAt = time.Now()
	dep.Version++
	return s.db.PutJSON(BucketDeployments, dep.SatelliteID, dep)
}

func (s *Store) GetDeployment(satelliteID string) (*Deployment, error) {
	var dep Deployment
	if err := s.db.GetJSON(BucketDeployments, satelliteID, &dep); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &dep, nil
}

func (s *Store) ListDeployments() ([]*Deployment, error) {
	var deployments []*Deployment
	err := s.db.ForEach(BucketDeployments, func(_, value []byte) error {
		var dep Deployment
		if err := unmarshalJSON(value, &dep); err != nil {
			return err
		}
		deployments = append(deployments, &dep)
		return nil
	})
	return deployments, err
}

func (s *Store) DeleteDeployment(satelliteID string) error {
	return s.db.Delete(BucketDeployments, satelliteID)
}

func (s *Store) AddModuleToDeployment(satelliteID string, mod ModuleDeployment) error {
	dep, err := s.GetDeployment(satelliteID)
	if err != nil {
		return err
	}
	if dep == nil {
		dep = &Deployment{
			SatelliteID: satelliteID,
			Modules:     []ModuleDeployment{mod},
		}
	} else {
		found := false
		for i, m := range dep.Modules {
			if m.ModuleID == mod.ModuleID {
				dep.Modules[i] = mod
				found = true
				break
			}
		}
		if !found {
			dep.Modules = append(dep.Modules, mod)
		}
	}
	return s.SetDeployment(dep)
}

func (s *Store) RemoveModuleFromDeployment(satelliteID, moduleID string) error {
	dep, err := s.GetDeployment(satelliteID)
	if err != nil {
		return err
	}
	if dep == nil {
		return nil
	}

	var filtered []ModuleDeployment
	for _, m := range dep.Modules {
		if m.ModuleID != moduleID {
			filtered = append(filtered, m)
		}
	}
	dep.Modules = filtered
	return s.SetDeployment(dep)
}

func (s *Store) GetSatellitesByModule(moduleID string) ([]string, error) {
	deployments, err := s.ListDeployments()
	if err != nil {
		return nil, err
	}

	var satelliteIDs []string
	for _, dep := range deployments {
		for _, mod := range dep.Modules {
			if mod.ModuleID == moduleID {
				satelliteIDs = append(satelliteIDs, dep.SatelliteID)
				break
			}
		}
	}
	return satelliteIDs, nil
}
