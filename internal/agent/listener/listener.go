package listener

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"gimpel/internal/agent/config"
	"gimpel/internal/agent/control"
	"gimpel/internal/agent/module"
)

type Manager struct {
	cfg           *config.AgentConfig
	supervisor    *module.Supervisor
	controlClient *control.Client

	mu        sync.RWMutex
	listeners map[string]*ManagedListener
}

type ManagedListener struct {
	Config   config.ListenerConfig
	Listener net.Listener
	cancel   context.CancelFunc
}

func NewManager(cfg *config.AgentConfig, supervisor *module.Supervisor, controlClient *control.Client) *Manager {
	return &Manager{
		cfg:           cfg,
		supervisor:    supervisor,
		controlClient: controlClient,
		listeners:     make(map[string]*ManagedListener),
	}
}

func (m *Manager) Run(ctx context.Context) error {
	for _, modCfg := range m.cfg.Modules {
		for _, lCfg := range modCfg.Listeners {
			if err := m.StartListener(ctx, lCfg); err != nil {
				log.WithError(err).WithField("listener", lCfg.ID).Error("failed to start listener")
			}
		}
	}

	<-ctx.Done()
	return ctx.Err()
}

func (m *Manager) StartListener(ctx context.Context, cfg config.ListenerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.listeners[cfg.ID]; exists {
		return nil
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	ln, err := net.Listen(cfg.Protocol, addr)
	if err != nil {
		return fmt.Errorf("binding to %s: %w", addr, err)
	}

	listenerCtx, cancel := context.WithCancel(ctx)
	ml := &ManagedListener{
		Config:   cfg,
		Listener: ln,
		cancel:   cancel,
	}
	m.listeners[cfg.ID] = ml

	go m.acceptLoop(listenerCtx, ml)

	log.WithFields(log.Fields{
		"listener": cfg.ID,
		"port":     cfg.Port,
		"protocol": cfg.Protocol,
	}).Info("listener started")

	return nil
}

func (m *Manager) StopListener(listenerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ml, ok := m.listeners[listenerID]
	if !ok {
		return nil
	}

	ml.cancel()
	ml.Listener.Close()
	delete(m.listeners, listenerID)

	log.WithField("listener", listenerID).Info("listener stopped")
	return nil
}

func (m *Manager) Stop() {
	m.mu.RLock()
	ids := make([]string, 0, len(m.listeners))
	for id := range m.listeners {
		ids = append(ids, id)
	}
	m.mu.RUnlock()

	for _, id := range ids {
		m.StopListener(id)
	}
}

func (m *Manager) acceptLoop(ctx context.Context, ml *ManagedListener) {
	for {
		conn, err := ml.Listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.WithError(err).WithField("listener", ml.Config.ID).Warn("accept error")
				continue
			}
		}

		go m.handleConnection(ctx, ml, conn)
	}
}

func (m *Manager) handleConnection(ctx context.Context, ml *ManagedListener, conn net.Conn) {
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	localAddr := conn.LocalAddr().(*net.TCPAddr)

	connID := uuid.New().String()

	log.WithFields(log.Fields{
		"connection_id": connID,
		"source":        remoteAddr.String(),
		"dest":          localAddr.String(),
		"listener":      ml.Config.ID,
	}).Debug("accepted connection")

	if ml.Config.HighInteraction {
		m.handleHIConnection(ctx, ml, conn, connID, remoteAddr)
		return
	}

	connInfo := &module.ConnectionInfo{
		ConnectionID: connID,
		SourceIP:     remoteAddr.IP.String(),
		SourcePort:   uint32(remoteAddr.Port),
		DestIP:       localAddr.IP.String(),
		DestPort:     uint32(localAddr.Port),
		Protocol:     ml.Config.Protocol,
	}

	dataPort, err := m.supervisor.HandleConnection(ctx, ml.Config.ModuleID, connInfo)
	if err != nil {
		log.WithError(err).WithField("module", ml.Config.ModuleID).Warn("module rejected connection")
		conn.Close()
		return
	}

	moduleAddr := fmt.Sprintf("127.0.0.1:%d", dataPort)
	moduleConn, err := net.DialTimeout("tcp", moduleAddr, 5*time.Second)
	if err != nil {
		log.WithError(err).WithField("module", ml.Config.ModuleID).Warn("failed to connect to module data port")
		conn.Close()
		return
	}

	moduleConn.Write([]byte(connID))

	go func() {
		defer conn.Close()
		defer moduleConn.Close()
		proxyConnections(ctx, conn, moduleConn)
	}()
}

func proxyConnections(ctx context.Context, client, server net.Conn) {
	done := make(chan struct{}, 2)

	go func() {
		copyData(client, server)
		done <- struct{}{}
	}()

	go func() {
		copyData(server, client)
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}

func copyData(dst, src net.Conn) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if ew != nil || nw != nr {
				return
			}
		}
		if er != nil {
			return
		}
	}
}

func (m *Manager) handleHIConnection(ctx context.Context, ml *ManagedListener, conn net.Conn, connID string, remoteAddr *net.TCPAddr) {
	resp, err := m.controlClient.RequestHISession(ctx, ml.Config.ID, remoteAddr.IP.String(), uint32(remoteAddr.Port))
	if err != nil {
		log.WithError(err).Warn("failed to request HI session")
		return
	}

	log.WithFields(log.Fields{
		"session_id": resp.SessionId,
		"endpoint":   resp.SandboxEndpoint,
	}).Debug("HI session established")

	if err := proxyToEndpoint(ctx, conn, resp.SandboxEndpoint); err != nil {
		log.WithError(err).Warn("HI proxy failed")
	}
}

func proxyToEndpoint(ctx context.Context, clientConn net.Conn, endpoint string) error {
	serverConn, err := net.Dial("tcp", endpoint)
	if err != nil {
		return fmt.Errorf("connecting to sandbox: %w", err)
	}
	defer serverConn.Close()

	errCh := make(chan error, 2)

	go func() {
		_, err := copyWithContext(ctx, serverConn, clientConn)
		errCh <- err
	}()

	go func() {
		_, err := copyWithContext(ctx, clientConn, serverConn)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func copyWithContext(ctx context.Context, dst net.Conn, src net.Conn) (int64, error) {
	buf := make([]byte, 32*1024)
	var written int64

	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}

		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
		}
		if er != nil {
			return written, er
		}
	}
}
