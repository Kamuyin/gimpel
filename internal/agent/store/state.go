package store

import (
	"time"

	"gimpel/pkg/storage"
)

const stateKey = "agent_state"
const deploymentKey = "deployment"

func (s *Store) SaveState(state *AgentState) error {
	return s.db.PutJSON(BucketState, stateKey, state)
}

func (s *Store) GetState() (*AgentState, error) {
	var state AgentState
	if err := s.db.GetJSON(BucketState, stateKey, &state); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &state, nil
}

func (s *Store) GetAgentState() (*AgentState, error) {
	return s.GetState()
}

func (s *Store) SaveAgentState(state *AgentState) error {
	return s.SaveState(state)
}

func (s *Store) UpdateSyncTime() error {
	state, err := s.GetState()
	if err != nil {
		return err
	}
	if state == nil {
		state = &AgentState{}
	}
	state.LastSyncAt = time.Now()
	return s.SaveState(state)
}

func (s *Store) SetRegistered(agentID, certPath, keyPath string) error {
	state, err := s.GetState()
	if err != nil {
		return err
	}
	if state == nil {
		state = &AgentState{}
	}
	state.AgentID = agentID
	state.Registered = true
	state.CertPath = certPath
	state.KeyPath = keyPath
	state.RegisteredAt = time.Now()
	return s.SaveState(state)
}

func (s *Store) IsRegistered() (bool, error) {
	state, err := s.GetState()
	if err != nil {
		return false, err
	}
	return state != nil && state.Registered, nil
}

func (s *Store) SaveDeploymentConfig(cfg *DeploymentConfig) error {
	cfg.ReceivedAt = time.Now()
	return s.db.PutJSON(BucketConfig, deploymentKey, cfg)
}

func (s *Store) GetDeploymentConfig() (*DeploymentConfig, error) {
	var cfg DeploymentConfig
	if err := s.db.GetJSON(BucketConfig, deploymentKey, &cfg); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (s *Store) GetDeploymentVersion() (int64, error) {
	cfg, err := s.GetDeploymentConfig()
	if err != nil {
		return 0, err
	}
	if cfg == nil {
		return 0, nil
	}
	return cfg.Version, nil
}
