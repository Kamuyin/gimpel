package control

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/agent/config"
)

type Client struct {
	cfg      *config.AgentConfig
	identity interface{ GetAgentID() string }

	mu   sync.RWMutex
	conn *grpc.ClientConn
	ctrl gimpelv1.AgentControlClient
}

type IdentityProvider interface {
	GetAgentID() string
}

type identityAdapter struct {
	agentID string
}

func (i *identityAdapter) GetAgentID() string { return i.agentID }

func NewClient(cfg *config.AgentConfig, identity interface{ AgentID() string }) (*Client, error) {
	return &Client{
		cfg: cfg,
		identity: &identityAdapter{
			agentID: identity.AgentID(),
		},
	}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var opts []grpc.DialOption

	tlsCfg := c.cfg.ControlPlane.TLS
	creds, err := LoadClientCredentials(tlsCfg.CertFile, tlsCfg.KeyFile, tlsCfg.CAFile)
	if err != nil {
		return fmt.Errorf("loading TLS credentials: %w", err)
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))

	conn, err := grpc.NewClient(c.cfg.ControlPlane.Address, opts...)
	if err != nil {
		return fmt.Errorf("dialing control plane: %w", err)
	}

	c.conn = conn
	c.ctrl = gimpelv1.NewAgentControlClient(conn)

	log.WithField("address", c.cfg.ControlPlane.Address).Info("connected to control plane")
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

type RegisterResponse struct {
	AgentId       string
	Certificate   []byte
	PrivateKey    []byte
	CaCertificate []byte
}

type Identity interface {
	GetHostname() string
	GetPublicIPs() []string
}

func (c *Client) Register(ctx context.Context, token string, identity Identity) (*RegisterResponse, error) {
	c.mu.RLock()
	ctrl := c.ctrl
	c.mu.RUnlock()

	if ctrl == nil {
		return nil, fmt.Errorf("not connected")
	}

	resp, err := ctrl.Register(ctx, &gimpelv1.RegisterRequest{
		Token:     token,
		Hostname:  identity.GetHostname(),
		PublicIps: identity.GetPublicIPs(),
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	})
	if err != nil {
		return nil, fmt.Errorf("register RPC: %w", err)
	}

	return &RegisterResponse{
		AgentId:       resp.AgentId,
		Certificate:   resp.Certificate,
		PrivateKey:    resp.PrivateKey,
		CaCertificate: resp.CaCertificate,
	}, nil
}

type GetConfigResponse struct {
	Updated bool
	Config  *gimpelv1.AgentConfig
}

func (c *Client) GetConfig(ctx context.Context, currentVersion string) (*GetConfigResponse, error) {
	c.mu.RLock()
	ctrl := c.ctrl
	c.mu.RUnlock()

	if ctrl == nil {
		return nil, fmt.Errorf("not connected")
	}

	resp, err := ctrl.GetConfig(ctx, &gimpelv1.GetConfigRequest{
		AgentId:        c.identity.GetAgentID(),
		CurrentVersion: currentVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("get config RPC: %w", err)
	}

	return &GetConfigResponse{
		Updated: resp.Updated,
		Config:  resp.Config,
	}, nil
}

func (c *Client) RunHeartbeatLoop(ctx context.Context, interval time.Duration, metricsCollector func() (float64, float64)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := c.sendHeartbeat(ctx, metricsCollector); err != nil {
				log.WithError(err).Warn("heartbeat failed")
			}
		}
	}
}

func (c *Client) sendHeartbeat(ctx context.Context, metricsCollector func() (float64, float64)) error {
	c.mu.RLock()
	ctrl := c.ctrl
	c.mu.RUnlock()

	if ctrl == nil {
		return fmt.Errorf("not connected")
	}

	cpuUsage, memUsage := metricsCollector()

	resp, err := ctrl.Heartbeat(ctx, &gimpelv1.HeartbeatRequest{
		AgentId:   c.identity.GetAgentID(),
		Timestamp: time.Now().UnixNano(),
		CpuUsage:  cpuUsage,
		MemUsage:  memUsage,
	})
	if err != nil {
		return fmt.Errorf("heartbeat RPC: %w", err)
	}

	if resp.ConfigStale {
		log.Debug("control plane indicates config is stale")
	}

	return nil
}

func (c *Client) RequestHISession(ctx context.Context, listenerID, sourceIP string, sourcePort uint32) (*gimpelv1.HISessionResponse, error) {
	c.mu.RLock()
	ctrl := c.ctrl
	c.mu.RUnlock()

	if ctrl == nil {
		return nil, fmt.Errorf("not connected")
	}

	return ctrl.RequestHISession(ctx, &gimpelv1.HISessionRequest{
		AgentId:    c.identity.GetAgentID(),
		ListenerId: listenerID,
		SourceIp:   sourceIP,
		SourcePort: sourcePort,
	})
}
