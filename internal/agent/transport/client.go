package transport

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "gimpel/api/go/v1"
)

type Client struct {
	conn *grpc.ClientConn
	rpc  pb.GimpelControlClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		rpc:  pb.NewGimpelControlClient(conn),
	}, nil
}

func (c *Client) SendHeartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	return c.rpc.Heartbeat(ctx, req)
}

func (c *Client) Close() error {
	return c.conn.Close()
}
