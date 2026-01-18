package ingest

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"

	gimpelv1 "gimpel/api/go/v1"
)

type MockStream struct {
	mock.Mock
	grpc.ServerStream
}

func (m *MockStream) Recv() (*gimpelv1.StreamEventsRequest, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gimpelv1.StreamEventsRequest), args.Error(1)
}

func (m *MockStream) SendAndClose(resp *gimpelv1.StreamEventsResponse) error {
	args := m.Called(resp)
	return args.Error(0)
}

func (m *MockStream) Context() context.Context {
	return context.Background()
}

func TestStreamEvents(t *testing.T) {
	handler := NewHandler()
	mockStream := new(MockStream)

	events := []*gimpelv1.Event{
		{
			EventId:     "evt-123",
			Type:        gimpelv1.EventType_EVENT_TYPE_CONNECTION_OPEN,
			TimestampNs: time.Now().UnixNano(),
		},
	}

	mockStream.On("Recv").Return(&gimpelv1.StreamEventsRequest{
		Batch: &gimpelv1.EventBatch{
			AgentId: "agent-1",
			Events:  events,
		},
	}, nil).Once()

	mockStream.On("Recv").Return(nil, io.EOF).Once()

	mockStream.On("SendAndClose", &gimpelv1.StreamEventsResponse{
		AcceptedCount: 0,
	}).Return(nil)

	err := handler.StreamEvents(mockStream)
	assert.NoError(t, err)

	mockStream.AssertExpectations(t)
}
