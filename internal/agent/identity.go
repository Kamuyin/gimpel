package agent

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gimpel/internal/agent/config"
)

type Identity struct {
	ID         string
	Hostname   string
	PublicIPs  []string
	CertPath   string
	KeyPath    string
	CAPath     string
	Registered bool
}

func (i *Identity) AgentID() string        { return i.ID }
func (i *Identity) GetHostname() string    { return i.Hostname }
func (i *Identity) GetPublicIPs() []string { return i.PublicIPs }

func LoadIdentity(cfg *config.AgentConfig) (*Identity, error) {
	hostname, _ := os.Hostname()
	id := &Identity{
		Hostname: hostname,
	}

	idPath := filepath.Join(cfg.DataDir, "agent_id")
	if data, err := os.ReadFile(idPath); err == nil {
		id.ID = strings.TrimSpace(string(data))
	} else if cfg.AgentID != "" {
		id.ID = cfg.AgentID
	}

	id.CertPath = filepath.Join(cfg.DataDir, "cert.pem")
	id.KeyPath = filepath.Join(cfg.DataDir, "key.pem")
	id.CAPath = filepath.Join(cfg.DataDir, "ca.pem")

	if cfg.ControlPlane.TLS.CertFile != "" {
		id.CertPath = cfg.ControlPlane.TLS.CertFile
	}
	if cfg.ControlPlane.TLS.KeyFile != "" {
		id.KeyPath = cfg.ControlPlane.TLS.KeyFile
	}
	if cfg.ControlPlane.TLS.CAFile != "" {
		id.CAPath = cfg.ControlPlane.TLS.CAFile
	}

	if _, err := os.Stat(id.CertPath); err == nil {
		id.Registered = true
	}

	if id.ID == "" && !id.Registered {
		newID, err := generateAgentID()
		if err != nil {
			return nil, fmt.Errorf("generating agent ID: %w", err)
		}
		id.ID = newID
	}

	id.PublicIPs = discoverPublicIPs()
	return id, nil
}

func (id *Identity) Persist(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("creating data dir: %w", err)
	}

	idPath := filepath.Join(dataDir, "agent_id")
	if err := os.WriteFile(idPath, []byte(id.ID), 0600); err != nil {
		return fmt.Errorf("writing agent ID: %w", err)
	}
	return nil
}

func (id *Identity) SaveCredentials(dataDir string, cert, key, ca []byte) error {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return fmt.Errorf("creating data dir: %w", err)
	}

	certPath := filepath.Join(dataDir, "cert.pem")
	keyPath := filepath.Join(dataDir, "key.pem")
	caPath := filepath.Join(dataDir, "ca.pem")

	if err := os.WriteFile(certPath, cert, 0600); err != nil {
		return fmt.Errorf("writing cert: %w", err)
	}
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return fmt.Errorf("writing key: %w", err)
	}
	if err := os.WriteFile(caPath, ca, 0644); err != nil {
		return fmt.Errorf("writing CA: %w", err)
	}

	id.CertPath = certPath
	id.KeyPath = keyPath
	id.CAPath = caPath
	id.Registered = true
	return nil
}

func generateAgentID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "agent-" + hex.EncodeToString(b), nil
}

func discoverPublicIPs() []string {
	// TODO: Implement proper IP discovery
	return nil
}
