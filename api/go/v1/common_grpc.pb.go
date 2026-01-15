
package gimpelv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	GimpelControl_Ping_FullMethodName      = "/gimpel.v1.GimpelControl/Ping"
	GimpelControl_Heartbeat_FullMethodName = "/gimpel.v1.GimpelControl/Heartbeat"
)

type GimpelControlClient interface {
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
}

type gimpelControlClient struct {
	cc grpc.ClientConnInterface
}

func NewGimpelControlClient(cc grpc.ClientConnInterface) GimpelControlClient {
	return &gimpelControlClient{cc}
}

func (c *gimpelControlClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, GimpelControl_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gimpelControlClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, GimpelControl_Heartbeat_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type GimpelControlServer interface {
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	mustEmbedUnimplementedGimpelControlServer()
}

type UnimplementedGimpelControlServer struct{}

func (UnimplementedGimpelControlServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedGimpelControlServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Heartbeat not implemented")
}
func (UnimplementedGimpelControlServer) mustEmbedUnimplementedGimpelControlServer() {}
func (UnimplementedGimpelControlServer) testEmbeddedByValue()                       {}

type UnsafeGimpelControlServer interface {
	mustEmbedUnimplementedGimpelControlServer()
}

func RegisterGimpelControlServer(s grpc.ServiceRegistrar, srv GimpelControlServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&GimpelControl_ServiceDesc, srv)
}

func _GimpelControl_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GimpelControlServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GimpelControl_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GimpelControlServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GimpelControl_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GimpelControlServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GimpelControl_Heartbeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GimpelControlServer).Heartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var GimpelControl_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gimpel.v1.GimpelControl",
	HandlerType: (*GimpelControlServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _GimpelControl_Ping_Handler,
		},
		{
			MethodName: "Heartbeat",
			Handler:    _GimpelControl_Heartbeat_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "v1/common.proto",
}
