package module

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gimpelv1 "gimpel/api/go/v1"
)

type Client struct {
	socketPath string

	mu   sync.RWMutex
	conn *grpc.ClientConn
	svc  gimpelv1.ModuleServiceClient
}

func NewClient(socketPath string) (*Client, error) {
	c := &Client{
		socketPath: socketPath,
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) connect() error {
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.DialTimeout("unix", c.socketPath, 5*time.Second)
	}

	target := "passthrough:///" + c.socketPath

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
	)
	if err != nil {
		return fmt.Errorf("dialing module: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.svc = gimpelv1.NewModuleServiceClient(conn)
	c.mu.Unlock()

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

func (c *Client) HandleConnection(ctx context.Context, info *ConnectionInfo) (int32, error) {
	c.mu.RLock()
	svc := c.svc
	c.mu.RUnlock()

	if svc == nil {
		return 0, fmt.Errorf("not connected")
	}

	resp, err := svc.HandleConnection(ctx, &gimpelv1.HandleConnectionRequest{
		Connection: &gimpelv1.ConnectionInfo{
			ConnectionId: info.ConnectionID,
			SourceIp:     info.SourceIP,
			SourcePort:   info.SourcePort,
			DestIp:       info.DestIP,
			DestPort:     info.DestPort,
			Protocol:     info.Protocol,
			TimestampNs:  time.Now().UnixNano(),
		},
	})
	if err != nil {
		return 0, fmt.Errorf("handle connection RPC: %w", err)
	}

	if !resp.Accepted {
		return 0, fmt.Errorf("connection rejected by module")
	}

	return resp.DataPort, nil
}

func (c *Client) HealthCheck(ctx context.Context) (bool, string, error) {
	c.mu.RLock()
	svc := c.svc
	c.mu.RUnlock()

	if svc == nil {
		return false, "", fmt.Errorf("not connected")
	}

	resp, err := svc.HealthCheck(ctx, &gimpelv1.HealthCheckRequest{})
	if err != nil {
		return false, "", err
	}

	return resp.Healthy, resp.Status, nil
}
