package gimpelsdk

import (
	"context"
	"net"
)

type Module interface {
	Name() string
	Init(ctx *ModuleContext) error
	HandleConnection(ctx context.Context, conn net.Conn, info *ConnectionInfo) error
	HealthCheck(ctx context.Context) (bool, string)
	Shutdown(ctx context.Context) error
}

type ConnectionInfo struct {
	ConnectionID string
	SourceIP     string
	SourcePort   uint32
	DestIP       string
	DestPort     uint32
	Protocol     string
	Timestamp    int64
}

type ModuleContext struct {
	ModuleID   string
	SocketPath string
	Emitter    EventEmitter
	Logger     Logger

	done chan struct{}
}

func NewModuleContext(moduleID, socketPath string) *ModuleContext {
	return &ModuleContext{
		ModuleID:   moduleID,
		SocketPath: socketPath,
		Logger:     &defaultLogger{moduleID: moduleID},
		done:       make(chan struct{}),
	}
}

func (c *ModuleContext) Done() <-chan struct{} {
	return c.done
}

func (c *ModuleContext) Close() {
	close(c.done)
}

type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
}

type defaultLogger struct {
	moduleID string
}

func (l *defaultLogger) Debug(msg string, fields ...any) {}
func (l *defaultLogger) Info(msg string, fields ...any)  {}
func (l *defaultLogger) Warn(msg string, fields ...any)  {}
func (l *defaultLogger) Error(msg string, fields ...any) {}
