package supervisor

import (
	"fmt"
	"net"
	"syscall"
)

type Injector struct{}

func NewInjector() *Injector {
	return &Injector{}
}

func (i *Injector) Inject(socketPath string, conn *net.TCPConn) error {
	f, err := conn.File()
	if err != nil {
		return fmt.Errorf("failed to get conn file: %w", err)
	}
	defer f.Close()

	unixAddr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to resolve unix addr: %w", err)
	}

	unixConn, err := net.DialUnix("unix", nil, unixAddr)
	if err != nil {
		return fmt.Errorf("failed to dial module socket: %w", err)
	}
	defer unixConn.Close()

	oob := syscall.UnixRights(int(f.Fd()))

	_, _, err = unixConn.WriteMsgUnix([]byte("CONN"), oob, nil)
	if err != nil {
		return fmt.Errorf("failed to write unix msg: %w", err)
	}

	return nil
}
