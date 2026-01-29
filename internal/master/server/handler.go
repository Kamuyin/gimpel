package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/master/ca"
	"gimpel/internal/master/config"
	"gimpel/internal/master/session"
	"gimpel/internal/master/store"
)

type Handler struct {
	gimpelv1.UnimplementedAgentControlServer

	cfg        *config.MasterConfig
	store      *store.Store
	ca         *ca.CA
	sessionMgr *session.SessionManager
}

func NewHandler(
	cfg *config.MasterConfig,
	s *store.Store,
	caInstance *ca.CA,
	sessionMgr *session.SessionManager,
) *Handler {
	return &Handler{
		cfg:        cfg,
		store:      s,
		ca:         caInstance,
		sessionMgr: sessionMgr,
	}
}

func (h *Handler) Register(ctx context.Context, req *gimpelv1.RegisterRequest) (*gimpelv1.RegisterResponse, error) {
	if req.Token == "" {
		return nil, fmt.Errorf("pairing token is required")
	}

	pr, err := h.store.GetPairingByToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("checking pairing token: %w", err)
	}
	if pr == nil || pr.Used || time.Now().After(pr.ExpiresAt) {
		return nil, fmt.Errorf("pairing token is invalid or expired")
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

	satellite := &store.Satellite{
		ID:           agentID,
		Hostname:     req.Hostname,
		IPAddress:    firstIP(req.PublicIps),
		OS:           req.Os,
		Arch:         req.Arch,
		Status:       store.SatelliteStatusOnline,
		RegisteredAt: time.Now(),
		LastSeenAt:   time.Now(),
	}

	if err := h.store.RegisterSatellite(satellite); err != nil {
		return nil, fmt.Errorf("storing satellite: %w", err)
	}

	log.WithFields(log.Fields{
		"agent_id": agentID,
		"hostname": req.Hostname,
		"os":       req.Os,
		"arch":     req.Arch,
	}).Info("satellite registered")

	if err := h.store.MarkPairingUsed(pr.ID, agentID, req.Hostname); err != nil {
		log.WithError(err).Warn("failed to mark pairing as used")
	}

	caBundle := h.ca.CACertPEM()
	if h.cfg.ModuleStore.PublicKeyFile != "" {
		if pubKey, err := os.ReadFile(h.cfg.ModuleStore.PublicKeyFile); err == nil {
			if len(caBundle) > 0 && caBundle[len(caBundle)-1] != '\n' {
				caBundle = append(caBundle, '\n')
			}
			caBundle = append(caBundle, pubKey...)
		}
	}

	return &gimpelv1.RegisterResponse{
		AgentId:       agentID,
		Certificate:   signedCert.Certificate,
		PrivateKey:    signedCert.PrivateKey,
		CaCertificate: caBundle,
	}, nil
}

func (h *Handler) GetConfig(ctx context.Context, req *gimpelv1.GetConfigRequest) (*gimpelv1.GetConfigResponse, error) {
	deployment, err := h.store.GetDeployment(req.AgentId)
	if err != nil {
		return nil, fmt.Errorf("getting deployment: %w", err)
	}

	if deployment == nil {
		return &gimpelv1.GetConfigResponse{Updated: false}, nil
	}

	version, _ := parseVersion(req.CurrentVersion)
	if deployment.Version == version {
		return &gimpelv1.GetConfigResponse{Updated: false}, nil
	}

	return &gimpelv1.GetConfigResponse{
		Updated: true,
	}, nil
}

func (h *Handler) Heartbeat(ctx context.Context, req *gimpelv1.HeartbeatRequest) (*gimpelv1.HeartbeatResponse, error) {
	satellite, err := h.store.GetSatellite(req.AgentId)
	if err != nil {
		return nil, fmt.Errorf("getting satellite: %w", err)
	}

	if satellite == nil {
		return &gimpelv1.HeartbeatResponse{Ok: false}, nil
	}

	if err := h.store.UpdateSatelliteStatus(req.AgentId, store.SatelliteStatusOnline); err != nil {
		log.WithError(err).Warn("failed to update satellite status")
	}

	configStale := false

	return &gimpelv1.HeartbeatResponse{
		Ok:          true,
		ConfigStale: configStale,
	}, nil
}

func (h *Handler) RequestHISession(ctx context.Context, req *gimpelv1.HISessionRequest) (*gimpelv1.HISessionResponse, error) {
	satellite, err := h.store.GetSatellite(req.AgentId)
	if err != nil {
		return nil, fmt.Errorf("getting satellite: %w", err)
	}

	if satellite == nil {
		return nil, fmt.Errorf("satellite not registered")
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

func generateAgentID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "sat-" + hex.EncodeToString(b), nil
}

func firstIP(ips []string) string {
	if len(ips) > 0 {
		return ips[0]
	}
	return ""
}

func parseVersion(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

var _ gimpelv1.AgentControlServer = (*Handler)(nil)
