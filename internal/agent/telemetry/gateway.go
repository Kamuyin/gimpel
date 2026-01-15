package telemetry

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/agent/config"
	"gimpel/internal/agent/control"
)

type GatewayClient struct {
	cfg *config.AgentConfig

	mu     sync.RWMutex
	conn   *grpc.ClientConn
	client gimpelv1.IngestionServiceClient
	stream gimpelv1.IngestionService_StreamEventsClient
}

func NewGatewayClient(cfg *config.AgentConfig) (*GatewayClient, error) {
	gc := &GatewayClient{cfg: cfg}
	return gc, nil
}

func (gc *GatewayClient) connect(ctx context.Context) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if gc.conn != nil {
		return nil
	}

	var opts []grpc.DialOption

	tlsCfg := gc.cfg.Gateway.TLS
	if tlsCfg.CertFile != "" && tlsCfg.KeyFile != "" {
		creds, err := control.LoadClientCredentials(tlsCfg.CertFile, tlsCfg.KeyFile, tlsCfg.CAFile, tlsCfg.SkipVerify)
		if err != nil {
			return fmt.Errorf("loading TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(gc.cfg.Gateway.Address, opts...)
	if err != nil {
		return fmt.Errorf("dialing gateway: %w", err)
	}

	gc.conn = conn
	gc.client = gimpelv1.NewIngestionServiceClient(conn)

	stream, err := gc.client.StreamEvents(ctx)
	if err != nil {
		gc.conn.Close()
		gc.conn = nil
		gc.client = nil
		return fmt.Errorf("opening stream: %w", err)
	}
	gc.stream = stream

	log.WithField("address", gc.cfg.Gateway.Address).Info("connected to gateway")
	return nil
}

func (gc *GatewayClient) SendBatch(ctx context.Context, agentID string, events []*gimpelv1.Event) error {
	if err := gc.connect(ctx); err != nil {
		return err
	}

	gc.mu.RLock()
	stream := gc.stream
	gc.mu.RUnlock()

	if stream == nil {
		return fmt.Errorf("stream not available")
	}

	req := &gimpelv1.StreamEventsRequest{
		Batch: &gimpelv1.EventBatch{
			AgentId: agentID,
			Events:  events,
		},
	}

	if err := stream.Send(req); err != nil {
		gc.mu.Lock()
		gc.stream = nil
		gc.mu.Unlock()
		return fmt.Errorf("sending batch: %w", err)
	}

	return nil
}

func (gc *GatewayClient) Close() error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if gc.stream != nil {
		gc.stream.CloseSend()
		gc.stream = nil
	}

	if gc.conn != nil {
		gc.conn.Close()
		gc.conn = nil
	}

	return nil
}
