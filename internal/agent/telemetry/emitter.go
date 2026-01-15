package telemetry

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/agent/config"
)

type Emitter struct {
	cfg     *config.AgentConfig
	agentID string

	mu     sync.Mutex
	buffer *Buffer
	gw     *GatewayClient

	eventCh chan *gimpelv1.Event
}

func NewEmitter(ctx context.Context, cfg *config.AgentConfig, agentID string) (*Emitter, error) {
	buffer, err := NewBuffer(cfg.Gateway.BufferPath, cfg.Gateway.MaxBufferBytes)
	if err != nil {
		return nil, err
	}

	gw, err := NewGatewayClient(cfg)
	if err != nil {
		buffer.Close()
		return nil, err
	}

	e := &Emitter{
		cfg:     cfg,
		agentID: agentID,
		buffer:  buffer,
		gw:      gw,
		eventCh: make(chan *gimpelv1.Event, 1000),
	}

	return e, nil
}

func (e *Emitter) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.cfg.Gateway.FlushInterval)
	defer ticker.Stop()

	batch := make([]*gimpelv1.Event, 0, e.cfg.Gateway.BatchSize)

	for {
		select {
		case <-ctx.Done():
			e.flushBatch(ctx, batch)
			return ctx.Err()

		case event := <-e.eventCh:
			batch = append(batch, event)
			if len(batch) >= e.cfg.Gateway.BatchSize {
				e.flushBatch(ctx, batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				e.flushBatch(ctx, batch)
				batch = batch[:0]
			}
			e.drainBuffer(ctx)
		}
	}
}

func (e *Emitter) Emit(event *gimpelv1.Event) {
	if event.EventId == "" {
		event.EventId = uuid.New().String()
	}
	if event.AgentId == "" {
		event.AgentId = e.agentID
	}
	if event.TimestampNs == 0 {
		event.TimestampNs = time.Now().UnixNano()
	}

	select {
	case e.eventCh <- event:
	default:
		e.buffer.Push(event)
	}
}

func (e *Emitter) EmitConnectionOpen(moduleID, sessionID, sourceIP, destIP, protocol string, sourcePort, destPort uint32) {
	e.Emit(&gimpelv1.Event{
		ModuleId:   moduleID,
		SessionId:  sessionID,
		Type:       gimpelv1.EventType_EVENT_TYPE_CONNECTION_OPEN,
		SourceIp:   sourceIP,
		SourcePort: sourcePort,
		DestIp:     destIP,
		DestPort:   destPort,
		Protocol:   protocol,
	})
}

func (e *Emitter) EmitConnectionClose(moduleID, sessionID string) {
	e.Emit(&gimpelv1.Event{
		ModuleId:  moduleID,
		SessionId: sessionID,
		Type:      gimpelv1.EventType_EVENT_TYPE_CONNECTION_CLOSE,
	})
}

func (e *Emitter) EmitAuthAttempt(moduleID, sessionID string, labels map[string]string) {
	e.Emit(&gimpelv1.Event{
		ModuleId:  moduleID,
		SessionId: sessionID,
		Type:      gimpelv1.EventType_EVENT_TYPE_AUTH_ATTEMPT,
		Labels:    labels,
	})
}

func (e *Emitter) EmitCommand(moduleID, sessionID string, payload []byte) {
	e.Emit(&gimpelv1.Event{
		ModuleId:  moduleID,
		SessionId: sessionID,
		Type:      gimpelv1.EventType_EVENT_TYPE_COMMAND,
		Payload:   payload,
	})
}

func (e *Emitter) flushBatch(ctx context.Context, events []*gimpelv1.Event) {
	if len(events) == 0 {
		return
	}

	if err := e.gw.SendBatch(ctx, e.agentID, events); err != nil {
		log.WithError(err).WithField("count", len(events)).Warn("failed to send batch, buffering")
		for _, ev := range events {
			e.buffer.Push(ev)
		}
	}
}

func (e *Emitter) drainBuffer(ctx context.Context) {
	events, err := e.buffer.Pop(e.cfg.Gateway.BatchSize)
	if err != nil {
		log.WithError(err).Warn("failed to pop from buffer")
		return
	}

	if len(events) == 0 {
		return
	}

	if err := e.gw.SendBatch(ctx, e.agentID, events); err != nil {
		log.WithError(err).Warn("failed to drain buffer")
		for _, ev := range events {
			e.buffer.Push(ev)
		}
	}
}

func (e *Emitter) Flush(ctx context.Context) {
	close(e.eventCh)

	batch := make([]*gimpelv1.Event, 0, e.cfg.Gateway.BatchSize)
	for event := range e.eventCh {
		batch = append(batch, event)
	}
	e.flushBatch(ctx, batch)
	e.drainBuffer(ctx)

	e.buffer.Close()
	e.gw.Close()
}
