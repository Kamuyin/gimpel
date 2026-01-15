
package gimpelv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	AgentControl_Register_FullMethodName         = "/gimpel.v1.AgentControl/Register"
	AgentControl_GetConfig_FullMethodName        = "/gimpel.v1.AgentControl/GetConfig"
	AgentControl_Heartbeat_FullMethodName        = "/gimpel.v1.AgentControl/Heartbeat"
	AgentControl_RequestHISession_FullMethodName = "/gimpel.v1.AgentControl/RequestHISession"
)

type AgentControlClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error)
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
	RequestHISession(ctx context.Context, in *HISessionRequest, opts ...grpc.CallOption) (*HISessionResponse, error)
}

type agentControlClient struct {
	cc grpc.ClientConnInterface
}

func NewAgentControlClient(cc grpc.ClientConnInterface) AgentControlClient {
	return &agentControlClient{cc}
}

func (c *agentControlClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, AgentControl_Register_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentControlClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetConfigResponse)
	err := c.cc.Invoke(ctx, AgentControl_GetConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentControlClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, AgentControl_Heartbeat_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentControlClient) RequestHISession(ctx context.Context, in *HISessionRequest, opts ...grpc.CallOption) (*HISessionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HISessionResponse)
	err := c.cc.Invoke(ctx, AgentControl_RequestHISession_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type AgentControlServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error)
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	RequestHISession(context.Context, *HISessionRequest) (*HISessionResponse, error)
	mustEmbedUnimplementedAgentControlServer()
}

type UnimplementedAgentControlServer struct{}

func (UnimplementedAgentControlServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedAgentControlServer) GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetConfig not implemented")
}
func (UnimplementedAgentControlServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Heartbeat not implemented")
}
func (UnimplementedAgentControlServer) RequestHISession(context.Context, *HISessionRequest) (*HISessionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RequestHISession not implemented")
}
func (UnimplementedAgentControlServer) mustEmbedUnimplementedAgentControlServer() {}
func (UnimplementedAgentControlServer) testEmbeddedByValue()                      {}

type UnsafeAgentControlServer interface {
	mustEmbedUnimplementedAgentControlServer()
}

func RegisterAgentControlServer(s grpc.ServiceRegistrar, srv AgentControlServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AgentControl_ServiceDesc, srv)
}

func _AgentControl_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentControlServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AgentControl_Register_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentControlServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentControl_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentControlServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AgentControl_GetConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentControlServer).GetConfig(ctx, req.(*GetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentControl_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentControlServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AgentControl_Heartbeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentControlServer).Heartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentControl_RequestHISession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HISessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentControlServer).RequestHISession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AgentControl_RequestHISession_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentControlServer).RequestHISession(ctx, req.(*HISessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var AgentControl_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gimpel.v1.AgentControl",
	HandlerType: (*AgentControlServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _AgentControl_Register_Handler,
		},
		{
			MethodName: "GetConfig",
			Handler:    _AgentControl_GetConfig_Handler,
		},
		{
			MethodName: "Heartbeat",
			Handler:    _AgentControl_Heartbeat_Handler,
		},
		{
			MethodName: "RequestHISession",
			Handler:    _AgentControl_RequestHISession_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "v1/agent.proto",
}
