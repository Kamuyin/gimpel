
package gimpelv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	SandboxService_CreateSession_FullMethodName = "/gimpel.v1.SandboxService/CreateSession"
	SandboxService_StopSession_FullMethodName   = "/gimpel.v1.SandboxService/StopSession"
)

type SandboxServiceClient interface {
	CreateSession(ctx context.Context, in *CreateSessionRequest, opts ...grpc.CallOption) (*CreateSessionResponse, error)
	StopSession(ctx context.Context, in *StopSessionRequest, opts ...grpc.CallOption) (*StopSessionResponse, error)
}

type sandboxServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSandboxServiceClient(cc grpc.ClientConnInterface) SandboxServiceClient {
	return &sandboxServiceClient{cc}
}

func (c *sandboxServiceClient) CreateSession(ctx context.Context, in *CreateSessionRequest, opts ...grpc.CallOption) (*CreateSessionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateSessionResponse)
	err := c.cc.Invoke(ctx, SandboxService_CreateSession_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sandboxServiceClient) StopSession(ctx context.Context, in *StopSessionRequest, opts ...grpc.CallOption) (*StopSessionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StopSessionResponse)
	err := c.cc.Invoke(ctx, SandboxService_StopSession_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type SandboxServiceServer interface {
	CreateSession(context.Context, *CreateSessionRequest) (*CreateSessionResponse, error)
	StopSession(context.Context, *StopSessionRequest) (*StopSessionResponse, error)
	mustEmbedUnimplementedSandboxServiceServer()
}

type UnimplementedSandboxServiceServer struct{}

func (UnimplementedSandboxServiceServer) CreateSession(context.Context, *CreateSessionRequest) (*CreateSessionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateSession not implemented")
}
func (UnimplementedSandboxServiceServer) StopSession(context.Context, *StopSessionRequest) (*StopSessionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method StopSession not implemented")
}
func (UnimplementedSandboxServiceServer) mustEmbedUnimplementedSandboxServiceServer() {}
func (UnimplementedSandboxServiceServer) testEmbeddedByValue()                        {}

type UnsafeSandboxServiceServer interface {
	mustEmbedUnimplementedSandboxServiceServer()
}

func RegisterSandboxServiceServer(s grpc.ServiceRegistrar, srv SandboxServiceServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&SandboxService_ServiceDesc, srv)
}

func _SandboxService_CreateSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SandboxServiceServer).CreateSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SandboxService_CreateSession_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SandboxServiceServer).CreateSession(ctx, req.(*CreateSessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SandboxService_StopSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopSessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SandboxServiceServer).StopSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SandboxService_StopSession_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SandboxServiceServer).StopSession(ctx, req.(*StopSessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var SandboxService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gimpel.v1.SandboxService",
	HandlerType: (*SandboxServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateSession",
			Handler:    _SandboxService_CreateSession_Handler,
		},
		{
			MethodName: "StopSession",
			Handler:    _SandboxService_StopSession_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "v1/sandbox.proto",
}
