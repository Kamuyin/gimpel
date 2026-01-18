package module

import (
	"fmt"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func waitForSocket(socketPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
			if err == nil {
				conn.Close()
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for socket %s", socketPath)
}

type moduleLogger struct {
	moduleID string
	level    string
}

func (l *moduleLogger) Write(p []byte) (n int, err error) {
	msg := string(p)
	fields := log.Fields{"module": l.moduleID}
	
	switch l.level {
	case "error":
		log.WithFields(fields).Error(msg)
	case "warn":
		log.WithFields(fields).Warn(msg)
	default:
		log.WithFields(fields).Info(msg)
	}
	
	return len(p), nil
}
