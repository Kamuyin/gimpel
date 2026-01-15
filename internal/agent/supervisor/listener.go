package supervisor

import (
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Listener struct {
	addr       string
	targetSock string
	injector   *Injector
	ln         *net.TCPListener
	log        *logrus.Entry
	wg         sync.WaitGroup
	closing    chan struct{}
}

func NewListener(addr, targetSock string, log *logrus.Logger) (*Listener, error) {
	return &Listener{
		addr:       addr,
		targetSock: targetSock,
		injector:   NewInjector(),
		log:        log.WithField("component", "listener").WithField("addr", addr),
		closing:    make(chan struct{}),
	}, nil
}

func (l *Listener) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", l.addr)
	if err != nil {
		return err
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	l.ln = ln

	l.log.Info("Listener started")
	l.wg.Add(1)
	go l.acceptLoop()

	return nil
}

func (l *Listener) Stop() {
	close(l.closing)
	if l.ln != nil {
		l.ln.Close()
	}
	l.wg.Wait()
	l.log.Info("Listener stopped")
}

func (l *Listener) acceptLoop() {
	defer l.wg.Done()

	for {
		conn, err := l.ln.AcceptTCP()
		if err != nil {
			select {
			case <-l.closing:
				return
			default:
				l.log.WithError(err).Error("Accept failed")
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		l.handleConn(conn)
	}
}

func (l *Listener) handleConn(conn *net.TCPConn) {
	go func() {
		defer conn.Close()

		if err := l.injector.Inject(l.targetSock, conn); err != nil {
			l.log.WithError(err).Error("Failed to inject connection to module")
		} else {
			l.log.Debug("Connection injected successfully")
		}
	}()
}
