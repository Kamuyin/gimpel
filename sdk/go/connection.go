package gimpelsdk

import (
	"context"
	"fmt"
	"net"
	"sync"
)

var (
	pendingConns   = make(map[string]net.Conn)
	pendingConnsMu sync.RWMutex
)

func RegisterPendingConnection(connID string, conn net.Conn) {
	pendingConnsMu.Lock()
	defer pendingConnsMu.Unlock()
	pendingConns[connID] = conn
}

func ReceiveConnection(ctx context.Context, info *ConnectionInfo) (net.Conn, error) {
	pendingConnsMu.Lock()
	conn, ok := pendingConns[info.ConnectionID]
	if ok {
		delete(pendingConns, info.ConnectionID)
	}
	pendingConnsMu.Unlock()

	if ok {
		return conn, nil
	}

	return nil, fmt.Errorf("connection %s not found", info.ConnectionID)
}

type WrappedConn struct {
	net.Conn
	info *ConnectionInfo
}

func WrapConnection(conn net.Conn, info *ConnectionInfo) *WrappedConn {
	return &WrappedConn{
		Conn: conn,
		info: info,
	}
}

func (w *WrappedConn) Info() *ConnectionInfo {
	return w.info
}

func (w *WrappedConn) SourceAddr() string {
	return fmt.Sprintf("%s:%d", w.info.SourceIP, w.info.SourcePort)
}

func (w *WrappedConn) DestAddr() string {
	return fmt.Sprintf("%s:%d", w.info.DestIP, w.info.DestPort)
}
