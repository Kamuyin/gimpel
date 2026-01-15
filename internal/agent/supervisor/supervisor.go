package supervisor

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type ModuleConfig struct {
	ID         string
	Image      string
	ListenPort int
}

type Runtime interface {
	Start(ctx context.Context, cfg ModuleConfig) (string, error)
	Stop(ctx context.Context, id string) error
}

type Supervisor struct {
	runtime   Runtime
	listeners map[string]*Listener
	mu        sync.RWMutex
	log       *logrus.Logger
}

func New(runtime Runtime, log *logrus.Logger) *Supervisor {
	return &Supervisor{
		runtime:   runtime,
		listeners: make(map[string]*Listener),
		log:       log,
	}
}

func (s *Supervisor) EnsureModule(ctx context.Context, cfg ModuleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log := s.log.WithField("module", cfg.ID)

	if _, exists := s.listeners[cfg.ID]; exists {
		return nil
	}

	log.Info("Starting module...")

	socketPath, err := s.runtime.Start(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to start module runtime: %w", err)
	}

	listenAddr := fmt.Sprintf("0.0.0.0:%d", cfg.ListenPort)
	l, err := NewListener(listenAddr, socketPath, s.log)
	if err != nil {
		s.runtime.Stop(ctx, cfg.ID)
		return fmt.Errorf("failed to create listener: %w", err)
	}

	if err := l.Start(); err != nil {
		s.runtime.Stop(ctx, cfg.ID)
		return fmt.Errorf("failed to start listener: %w", err)
	}

	s.listeners[cfg.ID] = l
	log.Info("Module started successfully")

	return nil
}

func (s *Supervisor) StopModule(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if l, exists := s.listeners[id]; exists {
		l.Stop()
		delete(s.listeners, id)
	}

	return s.runtime.Stop(ctx, id)
}

func (s *Supervisor) Shutdown(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, l := range s.listeners {
		l.Stop()
		if err := s.runtime.Stop(ctx, id); err != nil {
			s.log.WithField("module", id).WithError(err).Error("Failed to stop runtime during shutdown")
		}
		delete(s.listeners, id)
	}
}
