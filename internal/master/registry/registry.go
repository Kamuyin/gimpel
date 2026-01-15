package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/master/config"
)

type Registry interface {
	Register(agent *Agent) error
	Get(agentID string) (*Agent, bool)
	Update(agentID string, fn func(*Agent)) error
	Delete(agentID string)
	List() []*Agent
	Count() int
}

type InMemoryRegistry struct {
	cfg *config.RegistryConfig

	mu     sync.RWMutex
	agents map[string]*Agent
}

func NewInMemoryRegistry(cfg *config.RegistryConfig) *InMemoryRegistry {
	return &InMemoryRegistry{
		cfg:    cfg,
		agents: make(map[string]*Agent),
	}
}

func (r *InMemoryRegistry) Register(agent *Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[agent.ID]; exists {
		return fmt.Errorf("agent %s already registered", agent.ID)
	}

	agent.RegisteredAt = time.Now()
	agent.LastSeen = time.Now()
	agent.State = AgentStateOnline

	r.agents[agent.ID] = agent

	log.WithFields(log.Fields{
		"agent_id": agent.ID,
		"hostname": agent.Hostname,
	}).Info("agent registered")

	return nil
}

func (r *InMemoryRegistry) Get(agentID string) (*Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, ok := r.agents[agentID]
	return agent, ok
}

func (r *InMemoryRegistry) Update(agentID string, fn func(*Agent)) error {
	r.mu.RLock()
	agent, ok := r.agents[agentID]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("agent %s not found", agentID)
	}

	fn(agent)
	return nil
}

func (r *InMemoryRegistry) Delete(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, agentID)
	log.WithField("agent_id", agentID).Info("agent removed from registry")
}

func (r *InMemoryRegistry) List() []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents
}

func (r *InMemoryRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

func (r *InMemoryRegistry) RunHealthChecker(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.checkAgentHealth()
		}
	}
}

func (r *InMemoryRegistry) checkAgentHealth() {
	r.mu.RLock()
	agents := make([]*Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	r.mu.RUnlock()

	staleThreshold := r.cfg.StaleTimeout
	offlineThreshold := staleThreshold * 3

	for _, agent := range agents {
		sinceLastSeen := time.Since(agent.LastSeen)

		if sinceLastSeen > offlineThreshold {
			agent.MarkOffline()
			log.WithField("agent_id", agent.ID).Warn("agent marked offline")
		} else if sinceLastSeen > staleThreshold {
			agent.MarkStale()
			log.WithField("agent_id", agent.ID).Debug("agent marked stale")
		}
	}
}
