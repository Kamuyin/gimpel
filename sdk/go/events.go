package gimpelsdk

import (
	"time"

	"github.com/google/uuid"
)

type EventType int

const (
	EventTypeConnectionOpen EventType = iota + 1
	EventTypeConnectionClose
	EventTypeDataReceived
	EventTypeDataSent
	EventTypeAuthAttempt
	EventTypeCommand
	EventTypeFileAccess
	EventTypeMalwareDetected
	EventTypeCustom
)

type Event struct {
	ID         string
	ModuleID   string
	SessionID  string
	Type       EventType
	Timestamp  time.Time
	SourceIP   string
	SourcePort uint32
	DestIP     string
	DestPort   uint32
	Protocol   string
	Labels     map[string]string
	Payload    []byte
}

type EventEmitter interface {
	Emit(event *Event) error
	EmitConnectionOpen(sessionID, sourceIP, destIP, protocol string, sourcePort, destPort uint32) error
	EmitConnectionClose(sessionID string) error
	EmitAuthAttempt(sessionID string, labels map[string]string) error
	EmitCommand(sessionID string, command []byte) error
}

type LocalEventEmitter struct {
	moduleID string
	events   chan *Event
}

func NewLocalEventEmitter(moduleID string, bufferSize int) *LocalEventEmitter {
	return &LocalEventEmitter{
		moduleID: moduleID,
		events:   make(chan *Event, bufferSize),
	}
}

func (e *LocalEventEmitter) Emit(event *Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.ModuleID == "" {
		event.ModuleID = e.moduleID
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	select {
	case e.events <- event:
		return nil
	default:
		return nil
	}
}

func (e *LocalEventEmitter) EmitConnectionOpen(sessionID, sourceIP, destIP, protocol string, sourcePort, destPort uint32) error {
	return e.Emit(&Event{
		SessionID:  sessionID,
		Type:       EventTypeConnectionOpen,
		SourceIP:   sourceIP,
		SourcePort: sourcePort,
		DestIP:     destIP,
		DestPort:   destPort,
		Protocol:   protocol,
	})
}

func (e *LocalEventEmitter) EmitConnectionClose(sessionID string) error {
	return e.Emit(&Event{
		SessionID: sessionID,
		Type:      EventTypeConnectionClose,
	})
}

func (e *LocalEventEmitter) EmitAuthAttempt(sessionID string, labels map[string]string) error {
	return e.Emit(&Event{
		SessionID: sessionID,
		Type:      EventTypeAuthAttempt,
		Labels:    labels,
	})
}

func (e *LocalEventEmitter) EmitCommand(sessionID string, command []byte) error {
	return e.Emit(&Event{
		SessionID: sessionID,
		Type:      EventTypeCommand,
		Payload:   command,
	})
}

func (e *LocalEventEmitter) Events() <-chan *Event {
	return e.events
}

func (e *LocalEventEmitter) Close() {
	close(e.events)
}
