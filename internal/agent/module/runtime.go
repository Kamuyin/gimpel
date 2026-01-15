package module

import (
	"context"
	"net"
)

type RuntimeType string

const (
	RuntimeTypeUserspace  RuntimeType = "userspace"
	RuntimeTypeContainerd RuntimeType = "containerd"
)

type Runtime interface {
	Type() RuntimeType
	Start(ctx context.Context, spec *RuntimeSpec) (*RuntimeInstance, error)
	Stop(ctx context.Context, instance *RuntimeInstance) error
}

type RuntimeSpec struct {
	ID         string
	Image      string
	SocketPath string
	Env        map[string]string
	Isolated   bool
}

type RuntimeInstance struct {
	ID         string
	Pid        int
	SocketPath string
	Conn       net.Conn
	StopFunc   func()
}

func (i *RuntimeInstance) Stop() {
	if i.StopFunc != nil {
		i.StopFunc()
	}
}
