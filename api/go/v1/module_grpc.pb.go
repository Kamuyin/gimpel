
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

const (
	ModuleCatalogService_GetCatalog_FullMethodName           = "/gimpel.v1.ModuleCatalogService/GetCatalog"
	ModuleCatalogService_GetModuleAssignments_FullMethodName = "/gimpel.v1.ModuleCatalogService/GetModuleAssignments"
	ModuleCatalogService_DownloadModule_FullMethodName       = "/gimpel.v1.ModuleCatalogService/DownloadModule"
	ModuleCatalogService_VerifyModule_FullMethodName         = "/gimpel.v1.ModuleCatalogService/VerifyModule"
)

type ModuleCatalogServiceClient interface {
	GetCatalog(ctx context.Context, in *GetCatalogRequest, opts ...grpc.CallOption) (*GetCatalogResponse, error)
	GetModuleAssignments(ctx context.Context, in *GetModuleAssignmentsRequest, opts ...grpc.CallOption) (*GetModuleAssignmentsResponse, error)
	DownloadModule(ctx context.Context, in *DownloadModuleRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[ModuleImageChunk], error)
	VerifyModule(ctx context.Context, in *VerifyModuleRequest, opts ...grpc.CallOption) (*VerifyModuleResponse, error)
}

type moduleCatalogServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewModuleCatalogServiceClient(cc grpc.ClientConnInterface) ModuleCatalogServiceClient {
	return &moduleCatalogServiceClient{cc}
}

func (c *moduleCatalogServiceClient) GetCatalog(ctx context.Context, in *GetCatalogRequest, opts ...grpc.CallOption) (*GetCatalogResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetCatalogResponse)
	err := c.cc.Invoke(ctx, ModuleCatalogService_GetCatalog_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *moduleCatalogServiceClient) GetModuleAssignments(ctx context.Context, in *GetModuleAssignmentsRequest, opts ...grpc.CallOption) (*GetModuleAssignmentsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetModuleAssignmentsResponse)
	err := c.cc.Invoke(ctx, ModuleCatalogService_GetModuleAssignments_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *moduleCatalogServiceClient) DownloadModule(ctx context.Context, in *DownloadModuleRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[ModuleImageChunk], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &ModuleCatalogService_ServiceDesc.Streams[0], ModuleCatalogService_DownloadModule_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[DownloadModuleRequest, ModuleImageChunk]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ModuleCatalogService_DownloadModuleClient = grpc.ServerStreamingClient[ModuleImageChunk]

func (c *moduleCatalogServiceClient) VerifyModule(ctx context.Context, in *VerifyModuleRequest, opts ...grpc.CallOption) (*VerifyModuleResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(VerifyModuleResponse)
	err := c.cc.Invoke(ctx, ModuleCatalogService_VerifyModule_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type ModuleCatalogServiceServer interface {
	GetCatalog(context.Context, *GetCatalogRequest) (*GetCatalogResponse, error)
	GetModuleAssignments(context.Context, *GetModuleAssignmentsRequest) (*GetModuleAssignmentsResponse, error)
	DownloadModule(*DownloadModuleRequest, grpc.ServerStreamingServer[ModuleImageChunk]) error
	VerifyModule(context.Context, *VerifyModuleRequest) (*VerifyModuleResponse, error)
	mustEmbedUnimplementedModuleCatalogServiceServer()
}

type UnimplementedModuleCatalogServiceServer struct{}

func (UnimplementedModuleCatalogServiceServer) GetCatalog(context.Context, *GetCatalogRequest) (*GetCatalogResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetCatalog not implemented")
}
func (UnimplementedModuleCatalogServiceServer) GetModuleAssignments(context.Context, *GetModuleAssignmentsRequest) (*GetModuleAssignmentsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetModuleAssignments not implemented")
}
func (UnimplementedModuleCatalogServiceServer) DownloadModule(*DownloadModuleRequest, grpc.ServerStreamingServer[ModuleImageChunk]) error {
	return status.Error(codes.Unimplemented, "method DownloadModule not implemented")
}
func (UnimplementedModuleCatalogServiceServer) VerifyModule(context.Context, *VerifyModuleRequest) (*VerifyModuleResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method VerifyModule not implemented")
}
func (UnimplementedModuleCatalogServiceServer) mustEmbedUnimplementedModuleCatalogServiceServer() {}
func (UnimplementedModuleCatalogServiceServer) testEmbeddedByValue()                              {}

type UnsafeModuleCatalogServiceServer interface {
	mustEmbedUnimplementedModuleCatalogServiceServer()
}

func RegisterModuleCatalogServiceServer(s grpc.ServiceRegistrar, srv ModuleCatalogServiceServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ModuleCatalogService_ServiceDesc, srv)
}

func _ModuleCatalogService_GetCatalog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCatalogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModuleCatalogServiceServer).GetCatalog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ModuleCatalogService_GetCatalog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModuleCatalogServiceServer).GetCatalog(ctx, req.(*GetCatalogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModuleCatalogService_GetModuleAssignments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetModuleAssignmentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModuleCatalogServiceServer).GetModuleAssignments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ModuleCatalogService_GetModuleAssignments_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModuleCatalogServiceServer).GetModuleAssignments(ctx, req.(*GetModuleAssignmentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ModuleCatalogService_DownloadModule_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(DownloadModuleRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ModuleCatalogServiceServer).DownloadModule(m, &grpc.GenericServerStream[DownloadModuleRequest, ModuleImageChunk]{ServerStream: stream})
}

type ModuleCatalogService_DownloadModuleServer = grpc.ServerStreamingServer[ModuleImageChunk]

func _ModuleCatalogService_VerifyModule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyModuleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModuleCatalogServiceServer).VerifyModule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ModuleCatalogService_VerifyModule_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModuleCatalogServiceServer).VerifyModule(ctx, req.(*VerifyModuleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var ModuleCatalogService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gimpel.v1.ModuleCatalogService",
	HandlerType: (*ModuleCatalogServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetCatalog",
			Handler:    _ModuleCatalogService_GetCatalog_Handler,
		},
		{
			MethodName: "GetModuleAssignments",
			Handler:    _ModuleCatalogService_GetModuleAssignments_Handler,
		},
		{
			MethodName: "VerifyModule",
			Handler:    _ModuleCatalogService_VerifyModule_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "DownloadModule",
			Handler:       _ModuleCatalogService_DownloadModule_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "v1/module.proto",
}
