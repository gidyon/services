// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package emailing

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

// EmailingClient is the client API for Emailing service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EmailingClient interface {
	// Sends email
	SendEmail(ctx context.Context, in *SendEmailRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}

type emailingClient struct {
	cc grpc.ClientConnInterface
}

func NewEmailingClient(cc grpc.ClientConnInterface) EmailingClient {
	return &emailingClient{cc}
}

func (c *emailingClient) SendEmail(ctx context.Context, in *SendEmailRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/gidyon.apis.Emailing/SendEmail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EmailingServer is the server API for Emailing service.
// All implementations must embed UnimplementedEmailingServer
// for forward compatibility
type EmailingServer interface {
	// Sends email
	SendEmail(context.Context, *SendEmailRequest) (*empty.Empty, error)
	mustEmbedUnimplementedEmailingServer()
}

// UnimplementedEmailingServer must be embedded to have forward compatible implementations.
type UnimplementedEmailingServer struct {
}

func (UnimplementedEmailingServer) SendEmail(context.Context, *SendEmailRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendEmail not implemented")
}
func (UnimplementedEmailingServer) mustEmbedUnimplementedEmailingServer() {}

// UnsafeEmailingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EmailingServer will
// result in compilation errors.
type UnsafeEmailingServer interface {
	mustEmbedUnimplementedEmailingServer()
}

func RegisterEmailingServer(s grpc.ServiceRegistrar, srv EmailingServer) {
	s.RegisterService(&_Emailing_serviceDesc, srv)
}

func _Emailing_SendEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendEmailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmailingServer).SendEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.Emailing/SendEmail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmailingServer).SendEmail(ctx, req.(*SendEmailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Emailing_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gidyon.apis.Emailing",
	HandlerType: (*EmailingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendEmail",
			Handler:    _Emailing_SendEmail_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "emailing.proto",
}
