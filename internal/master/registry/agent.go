package registry

import (
	"sync"
	"time"
)

type AgentState int

const (
	AgentStateUnknown AgentState = iota
	AgentStateOnline
	AgentStateStale
	AgentStateOffline
)

type Agent struct {
	ID          string
	Hostname    string
	PublicIPs   []string
	OS          string
	Arch        string
	Certificate []byte

	State        AgentState
	LastSeen     time.Time
	RegisteredAt time.Time

	CPUUsage float64
	MemUsage float64

	ConfigVersion string

	mu sync.RWMutex
}

func (a *Agent) UpdateHealth(cpuUsage, memUsage float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.CPUUsage = cpuUsage
	a.MemUsage = memUsage
	a.LastSeen = time.Now()
	a.State = AgentStateOnline
}

func (a *Agent) IsStale(timeout time.Duration) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return time.Since(a.LastSeen) > timeout
}

func (a *Agent) MarkStale() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.State = AgentStateStale
}

func (a *Agent) MarkOffline() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.State = AgentStateOffline
}

func (a *Agent) GetState() AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.State
}
