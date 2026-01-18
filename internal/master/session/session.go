package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gimpelv1 "gimpel/api/go/v1"
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

func (m *SessionManager) CreateSession(ctx context.Context, agentID, listenerID, sourceIP string, sourcePort uint32) (*HISession, error) {
	if len(m.cfg.Nodes) == 0 {
		return nil, fmt.Errorf("no sandbox nodes configured")
	}

	sessionID := fmt.Sprintf("hi-%s-%d", agentID, time.Now().UnixNano())

	node := m.selectNode()

	conn, err := grpc.NewClient(node, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dialing sandbox node %s: %w", node, err)
	}
	defer conn.Close()

	client := gimpelv1.NewSandboxServiceClient(conn)

	req := &gimpelv1.CreateSessionRequest{
		SessionId: sessionID,
		Image:     "default-honeypot",
		Env:       map[string]string{"AGENT_ID": agentID},
	}

	resp, err := client.CreateSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("creating session on sandbox %s: %w", node, err)
	}

	session := &HISession{
		ID:              sessionID,
		AgentID:         agentID,
		ListenerID:      listenerID,
		SourceIP:        sourceIP,
		SourcePort:      sourcePort,
		SandboxNode:     node,
		SandboxEndpoint: resp.Endpoint,
		TunnelKey:       resp.TunnelKey,
		State:           SessionStateActive,
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
