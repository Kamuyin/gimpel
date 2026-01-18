package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	log "github.com/sirupsen/logrus"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/master/ca"
	"gimpel/internal/master/config"
	"gimpel/internal/master/configstore"
	"gimpel/internal/master/registry"
	"gimpel/internal/master/session"
)

type Handler struct {
	gimpelv1.UnimplementedAgentControlServer

	cfg         *config.MasterConfig
	registry    registry.Registry
	ca          *ca.CA
	configStore configstore.ConfigStore
	sessionMgr  *session.SessionManager
}

func NewHandler(
	cfg *config.MasterConfig,
	reg registry.Registry,
	caInstance *ca.CA,
	configStore configstore.ConfigStore,
	sessionMgr *session.SessionManager,
) *Handler {
	return &Handler{
		cfg:         cfg,
		registry:    reg,
		ca:          caInstance,
		configStore: configStore,
		sessionMgr:  sessionMgr,
	}
}

func (h *Handler) Register(ctx context.Context, req *gimpelv1.RegisterRequest) (*gimpelv1.RegisterResponse, error) {
	if !h.validateToken(req.Token) {
		return nil, fmt.Errorf("invalid registration token")
	}

	agentID, err := generateAgentID()
	if err != nil {
		return nil, fmt.Errorf("generating agent ID: %w", err)
	}

	signedCert, err := h.ca.IssueCertificate(&ca.CertRequest{
		AgentID:   agentID,
		Hostname:  req.Hostname,
		PublicIPs: req.PublicIps,
	})
	if err != nil {
		return nil, fmt.Errorf("issuing certificate: %w", err)
	}

	agent := &registry.Agent{
		ID:          agentID,
		Hostname:    req.Hostname,
		PublicIPs:   req.PublicIps,
		OS:          req.Os,
		Arch:        req.Arch,
		Certificate: signedCert.Certificate,
	}

	if err := h.registry.Register(agent); err != nil {
		return nil, fmt.Errorf("registering agent: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"hostname": req.Hostname,
		"os":       req.Os,
		"arch":     req.Arch,
	}).Info("agent registered")

	return &gimpelv1.RegisterResponse{
		AgentId:       agentID,
		Certificate:   signedCert.Certificate,
		PrivateKey:    signedCert.PrivateKey,
		CaCertificate: h.ca.CACertPEM(),
	}, nil
}

func (h *Handler) GetConfig(ctx context.Context, req *gimpelv1.GetConfigRequest) (*gimpelv1.GetConfigResponse, error) {
	cfg, version, ok := h.configStore.GetConfig(req.AgentId)
	if !ok {
		return &gimpelv1.GetConfigResponse{Updated: false}, nil
	}

	if version == req.CurrentVersion {
		return &gimpelv1.GetConfigResponse{Updated: false}, nil
	}

	h.registry.Update(req.AgentId, func(a *registry.Agent) {
		a.ConfigVersion = version
	})

	return &gimpelv1.GetConfigResponse{
		Updated: true,
		Config:  cfg,
	}, nil
}

func (h *Handler) Heartbeat(ctx context.Context, req *gimpelv1.HeartbeatRequest) (*gimpelv1.HeartbeatResponse, error) {
	agent, ok := h.registry.Get(req.AgentId)
	if !ok {
		return &gimpelv1.HeartbeatResponse{Ok: false}, nil
	}

	agent.UpdateHealth(req.CpuUsage, req.MemUsage)

	_, version, ok := h.configStore.GetConfig(req.AgentId)
	configStale := ok && version != agent.ConfigVersion

	return &gimpelv1.HeartbeatResponse{
		Ok:          true,
		ConfigStale: configStale,
	}, nil
}

func (h *Handler) RequestHISession(ctx context.Context, req *gimpelv1.HISessionRequest) (*gimpelv1.HISessionResponse, error) {
	_, ok := h.registry.Get(req.AgentId)
	if !ok {
		return nil, fmt.Errorf("agent not registered")
	}

	sess, err := h.sessionMgr.CreateSession(ctx, req.AgentId, req.ListenerId, req.SourceIp, req.SourcePort)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	return &gimpelv1.HISessionResponse{
		SessionId:       sess.ID,
		SandboxEndpoint: sess.SandboxEndpoint,
		TunnelKey:       sess.TunnelKey,
	}, nil
}

func (h *Handler) validateToken(token string) bool {
	if len(h.cfg.RegistrationTokens) == 0 {
		return true
	}
	for _, t := range h.cfg.RegistrationTokens {
		if t == token {
			return true
		}
	}
	return false
}

func generateAgentID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "agent-" + hex.EncodeToString(b), nil
}

var _ gimpelv1.AgentControlServer = (*Handler)(nil)
