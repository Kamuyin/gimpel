package gimpelsdk

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"

	gimpelv1 "gimpel/api/go/v1"
)

type Server struct {
	module     Module
	ctx        *ModuleContext
	grpcServer *grpc.Server
	grpcLn     net.Listener
	dataLn     net.Listener
	dataPort   int

	pendingMu sync.Mutex
	pending   map[string]*ConnectionInfo
}

func NewServer(module Module) *Server {
	moduleID := os.Getenv("GIMPEL_MODULE_ID")
	if moduleID == "" {
		moduleID = module.Name()
	}

	socketPath := os.Getenv("GIMPEL_SOCKET")
	if socketPath == "" {
		socketPath = fmt.Sprintf("/tmp/gimpel-%s.sock", moduleID)
	}

	return &Server{
		module:  module,
		ctx:     NewModuleContext(moduleID, socketPath),
		pending: make(map[string]*ConnectionInfo),
	}
}

func (s *Server) Run() error {
	if err := s.module.Init(s.ctx); err != nil {
		return fmt.Errorf("module init: %w", err)
	}

	os.Remove(s.ctx.SocketPath)

	grpcLn, err := net.Listen("unix", s.ctx.SocketPath)
	if err != nil {
		return fmt.Errorf("listen on socket: %w", err)
	}
	s.grpcLn = grpcLn

	dataLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		grpcLn.Close()
		return fmt.Errorf("listen on data port: %w", err)
	}
	s.dataLn = dataLn
	s.dataPort = dataLn.Addr().(*net.TCPAddr).Port

	log.Printf("module data port: %d", s.dataPort)

	s.grpcServer = grpc.NewServer()
	gimpelv1.RegisterModuleServiceServer(s.grpcServer, &serviceHandler{server: s})

	errCh := make(chan error, 2)
	go func() {
		errCh <- s.grpcServer.Serve(grpcLn)
	}()
	go s.acceptDataConnections()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		return s.Shutdown()
	}
}

func (s *Server) acceptDataConnections() {
	for {
		conn, err := s.dataLn.Accept()
		if err != nil {
			return
		}

		buf := make([]byte, 64)
		n, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			continue
		}
		connID := string(buf[:n])

		s.pendingMu.Lock()
		info, ok := s.pending[connID]
		if ok {
			delete(s.pending, connID)
		}
		s.pendingMu.Unlock()

		if ok && info != nil {
			go func() {
				defer conn.Close()
				s.module.HandleConnection(context.Background(), conn, info)
			}()
		} else {
			conn.Close()
		}
	}
}

func (s *Server) Shutdown() error {
	ctx := context.Background()

	if err := s.module.Shutdown(ctx); err != nil {
		return fmt.Errorf("module shutdown: %w", err)
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.grpcLn != nil {
		s.grpcLn.Close()
	}
	if s.dataLn != nil {
		s.dataLn.Close()
	}

	s.ctx.Close()
	return nil
}

func (s *Server) DataPort() int {
	return s.dataPort
}

type serviceHandler struct {
	gimpelv1.UnimplementedModuleServiceServer
	server *Server
}

func (h *serviceHandler) HandleConnection(ctx context.Context, req *gimpelv1.HandleConnectionRequest) (*gimpelv1.HandleConnectionResponse, error) {
	info := &ConnectionInfo{
		ConnectionID: req.Connection.ConnectionId,
		SourceIP:     req.Connection.SourceIp,
		SourcePort:   req.Connection.SourcePort,
		DestIP:       req.Connection.DestIp,
		DestPort:     req.Connection.DestPort,
		Protocol:     req.Connection.Protocol,
		Timestamp:    req.Connection.TimestampNs,
	}

	h.server.pendingMu.Lock()
	h.server.pending[info.ConnectionID] = info
	h.server.pendingMu.Unlock()

	return &gimpelv1.HandleConnectionResponse{
		Accepted: true,
		DataPort: int32(h.server.dataPort),
	}, nil
}

func (h *serviceHandler) HealthCheck(ctx context.Context, req *gimpelv1.HealthCheckRequest) (*gimpelv1.HealthCheckResponse, error) {
	healthy, status := h.server.module.HealthCheck(ctx)
	return &gimpelv1.HealthCheckResponse{
		Healthy: healthy,
		Status:  status,
	}, nil
}
