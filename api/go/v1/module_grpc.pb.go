
package gimpelv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	ModuleService_HandleConnection_FullMethodName = "/gimpel.v1.ModuleService/HandleConnection"
	ModuleService_HealthCheck_FullMethodName      = "/gimpel.v1.ModuleService/HealthCheck"
)

type ModuleServiceClient interface {
	HandleConnection(ctx context.Context, in *HandleConnectionRequest, opts ...grpc.CallOption) (*HandleConnectionResponse, error)
	HealthCheck(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error)
}

type moduleServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewModuleServiceClient(cc grpc.ClientConnInterface) ModuleServiceClient {
	return &moduleServiceClient{cc}
}

func (c *moduleServiceClient) HandleConnection(ctx context.Context, in *HandleConnectionRequest, opts ...grpc.CallOption) (*HandleConnectionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HandleConnectionResponse)
	err := c.cc.Invoke(ctx, ModuleService_HandleConnection_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *moduleServiceClient) HealthCheck(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HealthCheckResponse)
	err := c.cc.Invoke(ctx, ModuleService_HealthCheck_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type ModuleServiceServer interface {
	HandleConnection(context.Context, *HandleConnectionRequest) (*HandleConnectionResponse, error)
	HealthCheck(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error)
	mustEmbedUnimplementedModuleServiceServer()
}

type UnimplementedModuleServiceServer struct{}

func (UnimplementedModuleServiceServer) HandleConnection(context.Context, *HandleConnectionRequest) (*HandleConnectionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method HandleConnection not implemented")
}
func (UnimplementedModuleServiceServer) HealthCheck(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method HealthCheck not implemented")
}
func (UnimplementedModuleServiceServer) mustEmbedUnimplementedModuleServiceServer() {}
func (UnimplementedModuleServiceServer) testEmbeddedByValue()                       {}

type UnsafeModuleServiceServer interface {
	mustEmbedUnimplementedModuleServiceServer()
}

func RegisterModuleServiceServer(s grpc.ServiceRegistrar, srv ModuleServiceServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ModuleService_ServiceDesc, srv)
}

func _ModuleService_HandleConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandleConnectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModuleServiceServer).HandleConnection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ModuleService_HandleConnection_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModuleServiceServer).HandleConnection(ctx, req.(*HandleConnectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModuleService_HealthCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModuleServiceServer).HealthCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ModuleService_HealthCheck_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModuleServiceServer).HealthCheck(ctx, req.(*HealthCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var ModuleService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gimpel.v1.ModuleService",
	HandlerType: (*ModuleServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HandleConnection",
			Handler:    _ModuleService_HandleConnection_Handler,
		},
		{
			MethodName: "HealthCheck",
			Handler:    _ModuleService_HealthCheck_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "v1/module.proto",
}
