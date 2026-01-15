package agent

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	pb "gimpel/api/go/v1"
	"gimpel/internal/agent/config"
	"gimpel/internal/agent/supervisor"
	"gimpel/internal/agent/transport"
)

type Agent struct {
	cfg        *config.Config
	client     *transport.Client
	supervisor *supervisor.Supervisor
	log        *logrus.Logger
}

func New(cfg *config.Config, log *logrus.Logger) (*Agent, error) {
	runtime, err := supervisor.NewContainerdRuntime("/run/containerd/containerd.sock", cfg.ModulesPath, log)
	if err != nil {
		return nil, err
	}

	super := supervisor.New(runtime, log)

	return &Agent{
		cfg:        cfg,
		supervisor: super,
		log:        log,
	}, nil
}

func (a *Agent) Run(ctx context.Context) error {
	a.log.WithField("master", a.cfg.MasterURL).Info("Starting Gimpel Agent...")

	client, err := transport.NewClient(a.cfg.MasterURL)
	if err != nil {
		return err
	}
	a.client = client
	defer client.Close()

	defer a.supervisor.Shutdown(context.Background())

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	a.heartbeat(ctx)

	for {
		select {
		case <-ctx.Done():
			a.log.Info("Context cancelled, shutting down...")
			return ctx.Err()
		case <-ticker.C:
			a.heartbeat(ctx)
		}
	}
}

func (a *Agent) heartbeat(ctx context.Context) {
	req := &pb.HeartbeatRequest{
		AgentId:   a.cfg.AgentID,
		Timestamp: time.Now().Unix(),
		CpuUsage:  0.0,
		MemUsage:  0.0,
	}

	resp, err := a.client.SendHeartbeat(ctx, req)
	if err != nil {
		a.log.WithError(err).Error("Heartbeat failed")
		return
	}

	if resp.ConfigStale {
		a.log.Info("Config is stale, requesting update (not implemented)")
	}
}
