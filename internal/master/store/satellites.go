package store

import (
	"fmt"
	"time"

	"gimpel/pkg/storage"
)

func (s *Store) RegisterSatellite(sat *Satellite) error {
	if sat.RegisteredAt.IsZero() {
		sat.RegisteredAt = time.Now()
	}
	sat.LastSeenAt = time.Now()
	return s.db.PutJSON(BucketSatellites, sat.ID, sat)
}

func (s *Store) GetSatellite(id string) (*Satellite, error) {
	var sat Satellite
	if err := s.db.GetJSON(BucketSatellites, id, &sat); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &sat, nil
}

func (s *Store) UpdateSatelliteStatus(id string, status SatelliteStatus) error {
	sat, err := s.GetSatellite(id)
	if err != nil {
		return err
	}
	if sat == nil {
		return fmt.Errorf("satellite %s not found", id)
	}
	sat.Status = status
	sat.LastSeenAt = time.Now()
	return s.db.PutJSON(BucketSatellites, id, sat)
}

func (s *Store) ListSatellites() ([]*Satellite, error) {
	var satellites []*Satellite
	err := s.db.ForEach(BucketSatellites, func(_, value []byte) error {
		var sat Satellite
		if err := unmarshalJSON(value, &sat); err != nil {
			return err
		}
		satellites = append(satellites, &sat)
		return nil
	})
	return satellites, err
}

func (s *Store) DeleteSatellite(id string) error {
	return s.db.Delete(BucketSatellites, id)
}

func (s *Store) CountSatellites() (int, error) {
	return s.db.Count(BucketSatellites)
}

func (s *Store) GetOnlineSatellites() ([]*Satellite, error) {
	all, err := s.ListSatellites()
	if err != nil {
		return nil, err
	}
	var online []*Satellite
	for _, sat := range all {
		if sat.Status == SatelliteStatusOnline {
			online = append(online, sat)
		}
	}
	return online, nil
}

func (s *Store) GetStaleSatellites(threshold time.Duration) ([]*Satellite, error) {
	all, err := s.ListSatellites()
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().Add(-threshold)
	var stale []*Satellite
	for _, sat := range all {
		if sat.LastSeenAt.Before(cutoff) {
			stale = append(stale, sat)
		}
	}
	return stale, nil
}
