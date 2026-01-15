//go:build linux || darwin || freebsd || openbsd || netbsd

package listener

import (
	"fmt"
	"net"
	"syscall"
)

func SendFD(unixConn *net.UnixConn, fd int) error {
	rights := syscall.UnixRights(fd)

	_, _, err := unixConn.WriteMsgUnix([]byte{0}, rights, nil)
	if err != nil {
		return fmt.Errorf("sending FD: %w", err)
	}

	return nil
}

func ReceiveFD(unixConn *net.UnixConn) (int, error) {
	buf := make([]byte, 1)
	oob := make([]byte, syscall.CmsgSpace(4))

	_, oobn, _, _, err := unixConn.ReadMsgUnix(buf, oob)
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

type FDConn struct {
	fd     int
	conn   net.Conn
	closed bool
}

func NewConnFromFD(fd int, network string) (*FDConn, error) {
	file := fdToFile(fd)
	conn, err := net.FileConn(file)
	if err != nil {
		return nil, fmt.Errorf("creating conn from FD: %w", err)
	}

	return &FDConn{
		fd:   fd,
		conn: conn,
	}, nil
}

func (c *FDConn) Conn() net.Conn {
	return c.conn
}

func (c *FDConn) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}
