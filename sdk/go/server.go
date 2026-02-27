package gimpelsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"syscall"

	"google.golang.org/grpc"

	gimpelv1 "gimpel/api/go/v1"
)

// ConnectionMode determines how connections are forwarded to the module.
type ConnectionMode string

const (
	// ConnectionModeTCPRelay uses TCP relay (default, works everywhere)
	ConnectionModeTCPRelay ConnectionMode = "tcp_relay"
	// ConnectionModeFDPass passes file descriptors over Unix socket (Linux only, more efficient)
	ConnectionModeFDPass ConnectionMode = "fdpass"
	// ConnectionModeProxy uses HTTP/websocket proxy (for future use)
	ConnectionModeProxy ConnectionMode = "proxy"
)

// FDConnection represents a connection received via file descriptor passing.
type FDConnection struct {
	Conn     net.Conn
	Info     *ConnectionInfo
	RawFD    int
	Received bool
}

// Server is the module server that handles gRPC control and data connections.
// It supports multiple connection modes and can be configured with ServerOption.
type Server struct {
	module     Module
	ctx        *ModuleContext
	connMode   ConnectionMode
	grpcServer *grpc.Server
	grpcLn     *net.UnixListener
	dataLn     net.Listener
	dataPort   int

	pendingMu sync.Mutex
	pending   map[string]*ConnectionInfo

	fdPassEnabled bool
	fdChan        chan *FDConnection
}

type ServerOption func(*Server)

func WithConnectionMode(mode ConnectionMode) ServerOption {
	return func(s *Server) {
		s.connMode = mode
	}
}

// WithFDPassing enables file descriptor passing mode.
func WithFDPassing() ServerOption {
	return func(s *Server) {
		s.fdPassEnabled = true
		s.connMode = ConnectionModeFDPass
	}
}

func NewServer(module Module, opts ...ServerOption) *Server {
	moduleID := os.Getenv("GIMPEL_MODULE_ID")
	if moduleID == "" {
		moduleID = module.Name()
	}

	socketPath := os.Getenv("GIMPEL_SOCKET")
	if socketPath == "" {
		socketPath = fmt.Sprintf("/tmp/gimpel-%s.sock", moduleID)
	}

	connMode := ConnectionMode(os.Getenv("GIMPEL_CONNECTION_MODE"))
	if connMode == "" {
		connMode = ConnectionModeTCPRelay
	}

	s := &Server{
		module:   module,
		ctx:      NewModuleContext(moduleID, socketPath),
		connMode: connMode,
		pending:  make(map[string]*ConnectionInfo),
		fdChan:   make(chan *FDConnection, 100),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Run() error {
	if err := s.module.Init(s.ctx); err != nil {
		return fmt.Errorf("module init: %w", err)
	}

	os.Remove(s.ctx.SocketPath)

	controlLn, err := net.ListenUnix("unix", &net.UnixAddr{Name: s.ctx.SocketPath, Net: "unix"})
	if err != nil {
		return fmt.Errorf("listen on control socket: %w", err)
	}
	s.grpcLn = controlLn

	if s.connMode == ConnectionModeTCPRelay {
		dataLn, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			controlLn.Close()
			return fmt.Errorf("listen on data port: %w", err)
		}
		s.dataLn = dataLn
		s.dataPort = dataLn.Addr().(*net.TCPAddr).Port
		log.Printf("module data port: %d", s.dataPort)

		go s.acceptDataConnections()
	}

	s.grpcServer = grpc.NewServer()
	gimpelv1.RegisterModuleServiceServer(s.grpcServer, &serviceHandler{server: s})

	if s.fdPassEnabled || s.connMode == ConnectionModeFDPass {
		go s.handleFDPassingConnections()
	}

	go func() {
		if err := s.grpcServer.Serve(controlLn); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	<-s.ctx.Done()

	return s.Shutdown()
}

func (s *Server) acceptDataConnections() {
	for {
		conn, err := s.dataLn.Accept()
		if err != nil {
			return
		}

		go s.handleDataConnection(conn)
	}
}

func (s *Server) handleDataConnection(conn net.Conn) {
	defer conn.Close()

	lenBuf := make([]byte, 2)
	if _, err := conn.Read(lenBuf); err != nil {
		return
	}
	handshakeLen := int(lenBuf[0])<<8 | int(lenBuf[1])

	handshakeData := make([]byte, handshakeLen)
	if _, err := conn.Read(handshakeData); err != nil {
		return
	}

	var handshake struct {
		ConnectionID string            `json:"connection_id"`
		SourceIP     string            `json:"source_ip"`
		SourcePort   uint32            `json:"source_port"`
		DestIP       string            `json:"dest_ip"`
		DestPort     uint32            `json:"dest_port"`
		Protocol     string            `json:"protocol"`
		Timestamp    int64             `json:"timestamp"`
		Metadata     map[string]string `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal(handshakeData, &handshake); err != nil {
		return
	}

	info := &ConnectionInfo{
		ConnectionID: handshake.ConnectionID,
		SourceIP:     handshake.SourceIP,
		SourcePort:   handshake.SourcePort,
		DestIP:       handshake.DestIP,
		DestPort:     handshake.DestPort,
		Protocol:     handshake.Protocol,
		Timestamp:    handshake.Timestamp,
	}

	s.module.HandleConnection(context.Background(), conn, info)
}

func (s *Server) handleFDPassingConnections() {
	for {
		conn, err := s.grpcLn.AcceptUnix()
		if err != nil {
			return
		}

		go s.handleFDPassConnection(conn)
	}
}

func (s *Server) handleFDPassConnection(conn *net.UnixConn) {
	defer conn.Close()

	lenBuf := make([]byte, 2)
	if _, err := conn.Read(lenBuf); err != nil {
		return
	}
	handshakeLen := int(lenBuf[0])<<8 | int(lenBuf[1])

	handshakeData := make([]byte, handshakeLen)
	if _, err := conn.Read(handshakeData); err != nil {
		return
	}

	var handshake struct {
		ConnectionID string            `json:"connection_id"`
		SourceIP     string            `json:"source_ip"`
		SourcePort   uint32            `json:"source_port"`
		DestIP       string            `json:"dest_ip"`
		DestPort     uint32            `json:"dest_port"`
		Protocol     string            `json:"protocol"`
		Timestamp    int64             `json:"timestamp"`
		Metadata     map[string]string `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal(handshakeData, &handshake); err != nil {
		return
	}

	fd, err := receiveFD(conn)
	if err != nil {
		return
	}

	file := os.NewFile(uintptr(fd), "")
	netConn, err := net.FileConn(file)
	file.Close()
	if err != nil {
		syscall.Close(fd)
		return
	}

	info := &ConnectionInfo{
		ConnectionID: handshake.ConnectionID,
		SourceIP:     handshake.SourceIP,
		SourcePort:   handshake.SourcePort,
		DestIP:       handshake.DestIP,
		DestPort:     handshake.DestPort,
		Protocol:     handshake.Protocol,
		Timestamp:    handshake.Timestamp,
	}

	go func() {
		defer netConn.Close()
		s.module.HandleConnection(context.Background(), netConn, info)
	}()
}

func receiveFD(conn *net.UnixConn) (int, error) {
	buf := make([]byte, 1)
	oob := make([]byte, syscall.CmsgSpace(4))

	_, oobn, _, _, err := conn.ReadMsgUnix(buf, oob)
	if err != nil {
		return -1, fmt.Errorf("receiving FD: %w", err)
	}

	msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return -1, fmt.Errorf("parsing control message: %w", err)
	}

	if len(msgs) != 1 {
		return -1, fmt.Errorf("expected 1 control message, got %d", len(msgs))
	}

	fds, err := syscall.ParseUnixRights(&msgs[0])
	if err != nil {
		return -1, fmt.Errorf("parsing unix rights: %w", err)
	}

	if len(fds) != 1 {
		return -1, fmt.Errorf("expected 1 FD, got %d", len(fds))
	}

	return fds[0], nil
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
