
package gimpelv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	IngestionService_StreamEvents_FullMethodName = "/gimpel.v1.IngestionService/StreamEvents"
)

type IngestionServiceClient interface {
	StreamEvents(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[StreamEventsRequest, StreamEventsResponse], error)
}

type ingestionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewIngestionServiceClient(cc grpc.ClientConnInterface) IngestionServiceClient {
	return &ingestionServiceClient{cc}
}

func (c *ingestionServiceClient) StreamEvents(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[StreamEventsRequest, StreamEventsResponse], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &IngestionService_ServiceDesc.Streams[0], IngestionService_StreamEvents_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamEventsRequest, StreamEventsResponse]{ClientStream: stream}
	return x, nil
}

type IngestionService_StreamEventsClient = grpc.ClientStreamingClient[StreamEventsRequest, StreamEventsResponse]

type IngestionServiceServer interface {
	StreamEvents(grpc.ClientStreamingServer[StreamEventsRequest, StreamEventsResponse]) error
	mustEmbedUnimplementedIngestionServiceServer()
}

type UnimplementedIngestionServiceServer struct{}

func (UnimplementedIngestionServiceServer) StreamEvents(grpc.ClientStreamingServer[StreamEventsRequest, StreamEventsResponse]) error {
	return status.Error(codes.Unimplemented, "method StreamEvents not implemented")
}
func (UnimplementedIngestionServiceServer) mustEmbedUnimplementedIngestionServiceServer() {}
func (UnimplementedIngestionServiceServer) testEmbeddedByValue()                          {}

type UnsafeIngestionServiceServer interface {
	mustEmbedUnimplementedIngestionServiceServer()
}

func RegisterIngestionServiceServer(s grpc.ServiceRegistrar, srv IngestionServiceServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&IngestionService_ServiceDesc, srv)
}

func _IngestionService_StreamEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(IngestionServiceServer).StreamEvents(&grpc.GenericServerStream[StreamEventsRequest, StreamEventsResponse]{ServerStream: stream})
}

type IngestionService_StreamEventsServer = grpc.ClientStreamingServer[StreamEventsRequest, StreamEventsResponse]

var IngestionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gimpel.v1.IngestionService",
	HandlerType: (*IngestionServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamEvents",
			Handler:       _IngestionService_StreamEvents_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "v1/telemetry.proto",
}
