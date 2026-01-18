//go:build linux || darwin || freebsd || openbsd || netbsd

package module

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type ConnectionForwarder struct {
	defaultMode ConnectionMode

	mu         sync.RWMutex
	forwarders map[string]*ModuleForwarder
}

type ModuleForwarder struct {
	moduleID   string
	mode       ConnectionMode
	socketPath string
	dataPort   int

	controlConn *net.UnixConn

	activeConns sync.Map

	metrics *ForwarderMetrics
}

type ForwardedConnection struct {
	ID        string
	Request   *ConnectionRequest
	StartedAt time.Time
	BytesIn   int64
	BytesOut  int64
	Done      chan struct{}
}

type ForwarderMetrics struct {
	ConnectionsTotal  int64
	ConnectionsActive int64
	BytesSent         int64
	BytesReceived     int64
	ErrorsTotal       int64
	AvgLatencyMs      float64
}

type ConnectionHandshake struct {
	ConnectionID string            `json:"connection_id"`
	SourceIP     string            `json:"source_ip"`
	SourcePort   uint32            `json:"source_port"`
	DestIP       string            `json:"dest_ip"`
	DestPort     uint32            `json:"dest_port"`
	Protocol     string            `json:"protocol"`
	Timestamp    int64             `json:"timestamp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

func NewConnectionForwarder(defaultMode ConnectionMode) *ConnectionForwarder {
	return &ConnectionForwarder{
		defaultMode: defaultMode,
		forwarders:  make(map[string]*ModuleForwarder),
	}
}

func (cf *ConnectionForwarder) RegisterModule(moduleID, socketPath string, dataPort int, mode ConnectionMode) error {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	if mode == "" {
		mode = cf.defaultMode
	}

	forwarder := &ModuleForwarder{
		moduleID:   moduleID,
		mode:       mode,
		socketPath: socketPath,
		dataPort:   dataPort,
		metrics:    &ForwarderMetrics{},
	}

	if mode == ConnectionModeFDPass {
		conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketPath, Net: "unix"})
		if err != nil {
			return fmt.Errorf("connecting to module socket: %w", err)
		}
		forwarder.controlConn = conn
	}

	cf.forwarders[moduleID] = forwarder

	log.WithFields(log.Fields{
		"module":   moduleID,
		"mode":     mode,
		"socket":   socketPath,
		"dataPort": dataPort,
	}).Info("module registered for connection forwarding")

	return nil
}

func (cf *ConnectionForwarder) UnregisterModule(moduleID string) {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	if forwarder, ok := cf.forwarders[moduleID]; ok {
		if forwarder.controlConn != nil {
			forwarder.controlConn.Close()
		}
		delete(cf.forwarders, moduleID)
	}
}

func (cf *ConnectionForwarder) Forward(ctx context.Context, req *ConnectionRequest) error {
	cf.mu.RLock()
	forwarder, ok := cf.forwarders[req.ModuleID]
	cf.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no forwarder registered for module %s", req.ModuleID)
	}

	switch forwarder.mode {
	case ConnectionModeFDPass:
		return forwarder.forwardFDPass(ctx, req)
	case ConnectionModeTCPRelay:
		return forwarder.forwardTCPRelay(ctx, req)
	case ConnectionModeProxy:
		return forwarder.forwardProxy(ctx, req)
	default:
		return fmt.Errorf("unsupported connection mode: %s", forwarder.mode)
	}
}

func (mf *ModuleForwarder) forwardFDPass(ctx context.Context, req *ConnectionRequest) error {
	if mf.controlConn == nil {
		return fmt.Errorf("no control connection for FD passing")
	}

	tcpConn, ok := req.Conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("FD passing requires TCP connection")
	}

	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return fmt.Errorf("getting syscall conn: %w", err)
	}

	var fd int
	var controlErr error

	err = rawConn.Control(func(fdArg uintptr) {
		fd, controlErr = syscall.Dup(int(fdArg))
	})
	if err != nil {
		return fmt.Errorf("controlling connection: %w", err)
	}
	if controlErr != nil {
		return fmt.Errorf("duplicating fd: %w", controlErr)
	}

	handshake := ConnectionHandshake{
		ConnectionID: req.ConnectionID,
		SourceIP:     req.SourceIP,
		SourcePort:   req.SourcePort,
		DestIP:       req.DestIP,
		DestPort:     req.DestPort,
		Protocol:     req.Protocol,
		Timestamp:    req.Timestamp.UnixNano(),
		Metadata:     req.Metadata,
	}

	handshakeData, err := json.Marshal(handshake)
	if err != nil {
		syscall.Close(fd)
		return fmt.Errorf("marshaling handshake: %w", err)
	}

	lenBuf := []byte{byte(len(handshakeData) >> 8), byte(len(handshakeData))}
	if _, err := mf.controlConn.Write(lenBuf); err != nil {
		syscall.Close(fd)
		return fmt.Errorf("writing handshake length: %w", err)
	}
	if _, err := mf.controlConn.Write(handshakeData); err != nil {
		syscall.Close(fd)
		return fmt.Errorf("writing handshake: %w", err)
	}

	rights := syscall.UnixRights(fd)
	if _, _, err := mf.controlConn.WriteMsgUnix([]byte{0}, rights, nil); err != nil {
		syscall.Close(fd)
		return fmt.Errorf("sending FD: %w", err)
	}

	syscall.Close(fd)

	req.Conn.Close()

	log.WithFields(log.Fields{
		"module":     mf.moduleID,
		"connection": req.ConnectionID,
	}).Debug("connection FD passed to module")

	return nil
}

func (mf *ModuleForwarder) forwardTCPRelay(ctx context.Context, req *ConnectionRequest) error {
	if mf.dataPort == 0 {
		return fmt.Errorf("no data port configured for module %s", mf.moduleID)
	}

	moduleAddr := fmt.Sprintf("127.0.0.1:%d", mf.dataPort)
	moduleConn, err := net.DialTimeout("tcp", moduleAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("connecting to module data port: %w", err)
	}

	handshake := ConnectionHandshake{
		ConnectionID: req.ConnectionID,
		SourceIP:     req.SourceIP,
		SourcePort:   req.SourcePort,
		DestIP:       req.DestIP,
		DestPort:     req.DestPort,
		Protocol:     req.Protocol,
		Timestamp:    req.Timestamp.UnixNano(),
		Metadata:     req.Metadata,
	}

	handshakeData, err := json.Marshal(handshake)
	if err != nil {
		moduleConn.Close()
		return fmt.Errorf("marshaling handshake: %w", err)
	}

	lenBuf := []byte{byte(len(handshakeData) >> 8), byte(len(handshakeData))}
	if _, err := moduleConn.Write(lenBuf); err != nil {
		moduleConn.Close()
		return fmt.Errorf("writing handshake length: %w", err)
	}
	if _, err := moduleConn.Write(handshakeData); err != nil {
		moduleConn.Close()
		return fmt.Errorf("writing handshake: %w", err)
	}

	fc := &ForwardedConnection{
		ID:        req.ConnectionID,
		Request:   req,
		StartedAt: time.Now(),
		Done:      make(chan struct{}),
	}
	mf.activeConns.Store(req.ConnectionID, fc)

	go func() {
		defer close(fc.Done)
		defer req.Conn.Close()
		defer moduleConn.Close()
		defer mf.activeConns.Delete(req.ConnectionID)

		mf.relay(ctx, req.Conn, moduleConn, fc)
	}()

	log.WithFields(log.Fields{
		"module":     mf.moduleID,
		"connection": req.ConnectionID,
		"dataPort":   mf.dataPort,
	}).Debug("TCP relay established")

	return nil
}

func (mf *ModuleForwarder) forwardProxy(ctx context.Context, req *ConnectionRequest) error {

	return mf.forwardTCPRelay(ctx, req)
}

func (mf *ModuleForwarder) relay(ctx context.Context, client, server net.Conn, fc *ForwardedConnection) {
	done := make(chan struct{}, 2)

	go func() {
		n, _ := io.Copy(server, client)
		fc.BytesIn += n
		done <- struct{}{}
	}()

	go func() {
		n, _ := io.Copy(client, server)
		fc.BytesOut += n
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	mf.metrics.BytesReceived += fc.BytesIn
	mf.metrics.BytesSent += fc.BytesOut
}

func (cf *ConnectionForwarder) GetMetrics(moduleID string) *ForwarderMetrics {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	if forwarder, ok := cf.forwarders[moduleID]; ok {
		return forwarder.metrics
	}
	return nil
}

func (cf *ConnectionForwarder) GetActiveConnections(moduleID string) []*ForwardedConnection {
	cf.mu.RLock()
	forwarder, ok := cf.forwarders[moduleID]
	cf.mu.RUnlock()

	if !ok {
		return nil
	}

	var conns []*ForwardedConnection
	forwarder.activeConns.Range(func(key, value interface{}) bool {
		if fc, ok := value.(*ForwardedConnection); ok {
			conns = append(conns, fc)
		}
		return true
	})

	return conns
}
