//go:build windows

package listener

import (
	"errors"
	"net"
)

var ErrNotSupported = errors.New("FD passing is not supported on Windows")

func SendFD(unixConn *net.UnixConn, fd int) error {
	return ErrNotSupported
}

func ReceiveFD(unixConn *net.UnixConn) (int, error) {
	return -1, ErrNotSupported
}

type FDConn struct {
	conn   net.Conn
	closed bool
}

func NewConnFromFD(fd int, network string) (*FDConn, error) {
	return nil, ErrNotSupported
}

func (c *FDConn) Conn() net.Conn {
	return c.conn
}

func (c *FDConn) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func fdToFile(fd int) interface{} {
	return nil
}
