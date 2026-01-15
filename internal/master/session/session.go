package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/master/config"
)

type SessionState int

const (
	SessionStatePending SessionState = iota
	SessionStateActive
	SessionStateEnded
)

type HISession struct {
	ID              string
	AgentID         string
	ListenerID      string
	SourceIP        string
	SourcePort      uint32
	SandboxNode     string
	SandboxEndpoint string
	TunnelKey       []byte
	State           SessionState
	CreatedAt       time.Time
	EndedAt         time.Time
}

type SessionManager struct {
	cfg *config.SandboxConfig

	mu       sync.RWMutex
	sessions map[string]*HISession
	nodeIdx  int
}

func NewSessionManager(cfg *config.SandboxConfig) *SessionManager {
	return &SessionManager{
		cfg:      cfg,
		sessions: make(map[string]*HISession),
	}
}

func (m *SessionManager) CreateSession(agentID, listenerID, sourceIP string, sourcePort uint32) (*HISession, error) {
	if len(m.cfg.Nodes) == 0 {
		return nil, fmt.Errorf("no sandbox nodes configured")
	}

	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("generating session ID: %w", err)
	}

	tunnelKey := make([]byte, 32)
	if _, err := rand.Read(tunnelKey); err != nil {
		return nil, fmt.Errorf("generating tunnel key: %w", err)
	}

	node := m.selectNode()

	session := &HISession{
		ID:              sessionID,
		AgentID:         agentID,
		ListenerID:      listenerID,
		SourceIP:        sourceIP,
		SourcePort:      sourcePort,
		SandboxNode:     node,
		SandboxEndpoint: fmt.Sprintf("%s:5000", node),
		TunnelKey:       tunnelKey,
		State:           SessionStatePending,
		CreatedAt:       time.Now(),
	}

	m.mu.Lock()
	m.sessions[sessionID] = session
	m.mu.Unlock()

	log.WithFields(log.Fields{
		"session_id":   sessionID,
		"agent_id":     agentID,
		"sandbox_node": node,
	}).Info("HI session created")

	return session, nil
}

func (m *SessionManager) GetSession(sessionID string) (*HISession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[sessionID]
	return session, ok
}

func (m *SessionManager) EndSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.sessions[sessionID]; ok {
		session.State = SessionStateEnded
		session.EndedAt = time.Now()
		log.WithField("session_id", sessionID).Info("HI session ended")
	}
}

func (m *SessionManager) ListActiveSessions() []*HISession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := make([]*HISession, 0)
	for _, session := range m.sessions {
		if session.State != SessionStateEnded {
			active = append(active, session)
		}
	}
	return active
}

func (m *SessionManager) selectNode() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	node := m.cfg.Nodes[m.nodeIdx%len(m.cfg.Nodes)]
	m.nodeIdx++
	return node
}

func generateSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "hi-" + hex.EncodeToString(b), nil
}
