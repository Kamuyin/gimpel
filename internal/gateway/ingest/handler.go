package ingest

import (
	"io"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gimpelv1 "gimpel/api/go/v1"
)

type Handler struct {
	gimpelv1.UnimplementedIngestionServiceServer
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) StreamEvents(stream gimpelv1.IngestionService_StreamEventsServer) error {
	log.Info("started event stream")
	defer log.Info("stopped event stream")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&gimpelv1.StreamEventsResponse{
				AcceptedCount: 0,
			})
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "receiving stream: %v", err)
		}

		batch := req.GetBatch()
		if batch == nil {
			continue
		}

		logger := log.WithField("agent_id", batch.AgentId)
		logger.Infof("received batch with %d events", len(batch.Events))

		for _, event := range batch.Events {
			logger.WithFields(log.Fields{
				"event_id":   event.EventId,
				"type":       event.Type,
				"module_id":  event.ModuleId,
				"session_id": event.SessionId,
			}).Debug("event received")
		}
	}
}
