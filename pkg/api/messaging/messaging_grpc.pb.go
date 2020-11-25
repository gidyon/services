// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package messaging

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// MessagingClient is the client API for Messaging service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MessagingClient interface {
	// Broadcasts a message
	BroadCastMessage(ctx context.Context, in *BroadCastMessageRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// Sends message to a single destination
	SendMessage(ctx context.Context, in *Message, opts ...grpc.CallOption) (*SendMessageResponse, error)
	// Retrieves a collection of messages
	ListMessages(ctx context.Context, in *ListMessagesRequest, opts ...grpc.CallOption) (*Messages, error)
	// Updates unread messages statuses to read
	ReadAll(ctx context.Context, in *MessageRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// Fetches count of new messages
	GetNewMessagesCount(ctx context.Context, in *MessageRequest, opts ...grpc.CallOption) (*NewMessagesCount, error)
}

type messagingClient struct {
	cc grpc.ClientConnInterface
}

func NewMessagingClient(cc grpc.ClientConnInterface) MessagingClient {
	return &messagingClient{cc}
}

func (c *messagingClient) BroadCastMessage(ctx context.Context, in *BroadCastMessageRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/gidyon.apis.Messaging/BroadCastMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messagingClient) SendMessage(ctx context.Context, in *Message, opts ...grpc.CallOption) (*SendMessageResponse, error) {
	out := new(SendMessageResponse)
	err := c.cc.Invoke(ctx, "/gidyon.apis.Messaging/SendMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messagingClient) ListMessages(ctx context.Context, in *ListMessagesRequest, opts ...grpc.CallOption) (*Messages, error) {
	out := new(Messages)
	err := c.cc.Invoke(ctx, "/gidyon.apis.Messaging/ListMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messagingClient) ReadAll(ctx context.Context, in *MessageRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/gidyon.apis.Messaging/ReadAll", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messagingClient) GetNewMessagesCount(ctx context.Context, in *MessageRequest, opts ...grpc.CallOption) (*NewMessagesCount, error) {
	out := new(NewMessagesCount)
	err := c.cc.Invoke(ctx, "/gidyon.apis.Messaging/GetNewMessagesCount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MessagingServer is the server API for Messaging service.
// All implementations must embed UnimplementedMessagingServer
// for forward compatibility
type MessagingServer interface {
	// Broadcasts a message
	BroadCastMessage(context.Context, *BroadCastMessageRequest) (*empty.Empty, error)
	// Sends message to a single destination
	SendMessage(context.Context, *Message) (*SendMessageResponse, error)
	// Retrieves a collection of messages
	ListMessages(context.Context, *ListMessagesRequest) (*Messages, error)
	// Updates unread messages statuses to read
	ReadAll(context.Context, *MessageRequest) (*empty.Empty, error)
	// Fetches count of new messages
	GetNewMessagesCount(context.Context, *MessageRequest) (*NewMessagesCount, error)
	mustEmbedUnimplementedMessagingServer()
}

// UnimplementedMessagingServer must be embedded to have forward compatible implementations.
type UnimplementedMessagingServer struct {
}

func (UnimplementedMessagingServer) BroadCastMessage(context.Context, *BroadCastMessageRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BroadCastMessage not implemented")
}
func (UnimplementedMessagingServer) SendMessage(context.Context, *Message) (*SendMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedMessagingServer) ListMessages(context.Context, *ListMessagesRequest) (*Messages, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMessages not implemented")
}
func (UnimplementedMessagingServer) ReadAll(context.Context, *MessageRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadAll not implemented")
}
func (UnimplementedMessagingServer) GetNewMessagesCount(context.Context, *MessageRequest) (*NewMessagesCount, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNewMessagesCount not implemented")
}
func (UnimplementedMessagingServer) mustEmbedUnimplementedMessagingServer() {}

// UnsafeMessagingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MessagingServer will
// result in compilation errors.
type UnsafeMessagingServer interface {
	mustEmbedUnimplementedMessagingServer()
}

func RegisterMessagingServer(s grpc.ServiceRegistrar, srv MessagingServer) {
	s.RegisterService(&_Messaging_serviceDesc, srv)
}

func _Messaging_BroadCastMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BroadCastMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessagingServer).BroadCastMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.Messaging/BroadCastMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessagingServer).BroadCastMessage(ctx, req.(*BroadCastMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Messaging_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessagingServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.Messaging/SendMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessagingServer).SendMessage(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

func _Messaging_ListMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessagingServer).ListMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.Messaging/ListMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessagingServer).ListMessages(ctx, req.(*ListMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Messaging_ReadAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessagingServer).ReadAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.Messaging/ReadAll",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessagingServer).ReadAll(ctx, req.(*MessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Messaging_GetNewMessagesCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessagingServer).GetNewMessagesCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.Messaging/GetNewMessagesCount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessagingServer).GetNewMessagesCount(ctx, req.(*MessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Messaging_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gidyon.apis.Messaging",
	HandlerType: (*MessagingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BroadCastMessage",
			Handler:    _Messaging_BroadCastMessage_Handler,
		},
		{
			MethodName: "SendMessage",
			Handler:    _Messaging_SendMessage_Handler,
		},
		{
			MethodName: "ListMessages",
			Handler:    _Messaging_ListMessages_Handler,
		},
		{
			MethodName: "ReadAll",
			Handler:    _Messaging_ReadAll_Handler,
		},
		{
			MethodName: "GetNewMessagesCount",
			Handler:    _Messaging_GetNewMessagesCount_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "messaging.proto",
}
