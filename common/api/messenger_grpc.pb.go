// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.0
// source: api/messenger.proto

package api

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	AuthService_Register_FullMethodName = "/api.AuthService/Register"
	AuthService_Login_FullMethodName    = "/api.AuthService/Login"
)

// AuthServiceClient is the client API for AuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthServiceClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
}

type authServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthServiceClient(cc grpc.ClientConnInterface) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, AuthService_Register_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, AuthService_Login_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthServiceServer is the server API for AuthService service.
// All implementations must embed UnimplementedAuthServiceServer
// for forward compatibility.
type AuthServiceServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	mustEmbedUnimplementedAuthServiceServer()
}

// UnimplementedAuthServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedAuthServiceServer struct{}

func (UnimplementedAuthServiceServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedAuthServiceServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedAuthServiceServer) mustEmbedUnimplementedAuthServiceServer() {}
func (UnimplementedAuthServiceServer) testEmbeddedByValue()                     {}

// UnsafeAuthServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthServiceServer will
// result in compilation errors.
type UnsafeAuthServiceServer interface {
	mustEmbedUnimplementedAuthServiceServer()
}

func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	// If the following call pancis, it indicates UnimplementedAuthServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AuthService_ServiceDesc, srv)
}

func _AuthService_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_Register_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_Login_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AuthService_ServiceDesc is the grpc.ServiceDesc for AuthService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _AuthService_Register_Handler,
		},
		{
			MethodName: "Login",
			Handler:    _AuthService_Login_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/messenger.proto",
}

const (
	PubSubService_Subscribe_FullMethodName = "/api.PubSubService/Subscribe"
	PubSubService_Publish_FullMethodName   = "/api.PubSubService/Publish"
)

// PubSubServiceClient is the client API for PubSubService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PubSubServiceClient interface {
	// Client subscribes to a chat and receives messages via streaming.
	Subscribe(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[Message], error)
	// Publish a message to a chat.
	Publish(ctx context.Context, in *PublishRequest, opts ...grpc.CallOption) (*PublishResponse, error)
}

type pubSubServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPubSubServiceClient(cc grpc.ClientConnInterface) PubSubServiceClient {
	return &pubSubServiceClient{cc}
}

func (c *pubSubServiceClient) Subscribe(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[Message], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &PubSubService_ServiceDesc.Streams[0], PubSubService_Subscribe_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[SubscribeRequest, Message]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type PubSubService_SubscribeClient = grpc.ServerStreamingClient[Message]

func (c *pubSubServiceClient) Publish(ctx context.Context, in *PublishRequest, opts ...grpc.CallOption) (*PublishResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PublishResponse)
	err := c.cc.Invoke(ctx, PubSubService_Publish_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PubSubServiceServer is the server API for PubSubService service.
// All implementations must embed UnimplementedPubSubServiceServer
// for forward compatibility.
type PubSubServiceServer interface {
	// Client subscribes to a chat and receives messages via streaming.
	Subscribe(*SubscribeRequest, grpc.ServerStreamingServer[Message]) error
	// Publish a message to a chat.
	Publish(context.Context, *PublishRequest) (*PublishResponse, error)
	mustEmbedUnimplementedPubSubServiceServer()
}

// UnimplementedPubSubServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedPubSubServiceServer struct{}

func (UnimplementedPubSubServiceServer) Subscribe(*SubscribeRequest, grpc.ServerStreamingServer[Message]) error {
	return status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (UnimplementedPubSubServiceServer) Publish(context.Context, *PublishRequest) (*PublishResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Publish not implemented")
}
func (UnimplementedPubSubServiceServer) mustEmbedUnimplementedPubSubServiceServer() {}
func (UnimplementedPubSubServiceServer) testEmbeddedByValue()                       {}

// UnsafePubSubServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PubSubServiceServer will
// result in compilation errors.
type UnsafePubSubServiceServer interface {
	mustEmbedUnimplementedPubSubServiceServer()
}

func RegisterPubSubServiceServer(s grpc.ServiceRegistrar, srv PubSubServiceServer) {
	// If the following call pancis, it indicates UnimplementedPubSubServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&PubSubService_ServiceDesc, srv)
}

func _PubSubService_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SubscribeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PubSubServiceServer).Subscribe(m, &grpc.GenericServerStream[SubscribeRequest, Message]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type PubSubService_SubscribeServer = grpc.ServerStreamingServer[Message]

func _PubSubService_Publish_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublishRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PubSubServiceServer).Publish(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PubSubService_Publish_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PubSubServiceServer).Publish(ctx, req.(*PublishRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PubSubService_ServiceDesc is the grpc.ServiceDesc for PubSubService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PubSubService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.PubSubService",
	HandlerType: (*PubSubServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Publish",
			Handler:    _PubSubService_Publish_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Subscribe",
			Handler:       _PubSubService_Subscribe_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "api/messenger.proto",
}

const (
	ChatHistoryService_GetMessages_FullMethodName          = "/api.ChatHistoryService/GetMessages"
	ChatHistoryService_SendMessageReadEvent_FullMethodName = "/api.ChatHistoryService/SendMessageReadEvent"
)

// ChatHistoryServiceClient is the client API for ChatHistoryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChatHistoryServiceClient interface {
	GetMessages(ctx context.Context, in *GetMessagesRequest, opts ...grpc.CallOption) (*GetMessagesResponse, error)
	SendMessageReadEvent(ctx context.Context, in *SendMessageReadEventRequest, opts ...grpc.CallOption) (*SendMessageReadEventResponse, error)
}

type chatHistoryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewChatHistoryServiceClient(cc grpc.ClientConnInterface) ChatHistoryServiceClient {
	return &chatHistoryServiceClient{cc}
}

func (c *chatHistoryServiceClient) GetMessages(ctx context.Context, in *GetMessagesRequest, opts ...grpc.CallOption) (*GetMessagesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetMessagesResponse)
	err := c.cc.Invoke(ctx, ChatHistoryService_GetMessages_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatHistoryServiceClient) SendMessageReadEvent(ctx context.Context, in *SendMessageReadEventRequest, opts ...grpc.CallOption) (*SendMessageReadEventResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendMessageReadEventResponse)
	err := c.cc.Invoke(ctx, ChatHistoryService_SendMessageReadEvent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChatHistoryServiceServer is the server API for ChatHistoryService service.
// All implementations must embed UnimplementedChatHistoryServiceServer
// for forward compatibility.
type ChatHistoryServiceServer interface {
	GetMessages(context.Context, *GetMessagesRequest) (*GetMessagesResponse, error)
	SendMessageReadEvent(context.Context, *SendMessageReadEventRequest) (*SendMessageReadEventResponse, error)
	mustEmbedUnimplementedChatHistoryServiceServer()
}

// UnimplementedChatHistoryServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedChatHistoryServiceServer struct{}

func (UnimplementedChatHistoryServiceServer) GetMessages(context.Context, *GetMessagesRequest) (*GetMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMessages not implemented")
}
func (UnimplementedChatHistoryServiceServer) SendMessageReadEvent(context.Context, *SendMessageReadEventRequest) (*SendMessageReadEventResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessageReadEvent not implemented")
}
func (UnimplementedChatHistoryServiceServer) mustEmbedUnimplementedChatHistoryServiceServer() {}
func (UnimplementedChatHistoryServiceServer) testEmbeddedByValue()                            {}

// UnsafeChatHistoryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChatHistoryServiceServer will
// result in compilation errors.
type UnsafeChatHistoryServiceServer interface {
	mustEmbedUnimplementedChatHistoryServiceServer()
}

func RegisterChatHistoryServiceServer(s grpc.ServiceRegistrar, srv ChatHistoryServiceServer) {
	// If the following call pancis, it indicates UnimplementedChatHistoryServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ChatHistoryService_ServiceDesc, srv)
}

func _ChatHistoryService_GetMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatHistoryServiceServer).GetMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ChatHistoryService_GetMessages_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatHistoryServiceServer).GetMessages(ctx, req.(*GetMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChatHistoryService_SendMessageReadEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMessageReadEventRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatHistoryServiceServer).SendMessageReadEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ChatHistoryService_SendMessageReadEvent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatHistoryServiceServer).SendMessageReadEvent(ctx, req.(*SendMessageReadEventRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ChatHistoryService_ServiceDesc is the grpc.ServiceDesc for ChatHistoryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ChatHistoryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.ChatHistoryService",
	HandlerType: (*ChatHistoryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetMessages",
			Handler:    _ChatHistoryService_GetMessages_Handler,
		},
		{
			MethodName: "SendMessageReadEvent",
			Handler:    _ChatHistoryService_SendMessageReadEvent_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/messenger.proto",
}
