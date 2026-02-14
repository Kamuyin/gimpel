package manager

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/sandbox/config"
)

type Session struct {
	ID        string
	Image     string
	Env       map[string]string
	Port      int
	TunnelKey []byte
	CreatedAt time.Time
}

type Manager struct {
	cfg *config.SandboxConfig

	mu       sync.RWMutex
	sessions map[string]*Session
	nextPort int
}

func New(cfg *config.SandboxConfig) *Manager {
	return &Manager{
		cfg:      cfg,
		sessions: make(map[string]*Session),
		nextPort: 6000,
	}
}

func (m *Manager) CreateSession(sessionID, image string, env map[string]string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	port := m.nextPort
	m.nextPort++

	tunnelKey := make([]byte, 32)
	if _, err := rand.Read(tunnelKey); err != nil {
		return nil, fmt.Errorf("generating tunnel key: %w", err)
	}

	session := &Session{
		ID:        sessionID,
		Image:     image,
		Env:       env,
		Port:      port,
		TunnelKey: tunnelKey,
		CreatedAt: time.Now(),
	}

	m.sessions[sessionID] = session

	log.WithFields(log.Fields{
		"session_id": sessionID,
		"image":      image,
		"port":       port,
	}).Info("session created")

	return session, nil
}

func (m *Manager) StopSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[sessionID]; !ok {
		return fmt.Errorf("session not found")
	}

	delete(m.sessions, sessionID)

	log.WithField("session_id", sessionID).Info("session stopped")

	return nil
}
