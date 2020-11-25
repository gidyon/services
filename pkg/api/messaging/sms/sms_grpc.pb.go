// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package sms

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

// SMSAPIClient is the client API for SMSAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SMSAPIClient interface {
	// Send an sms to its destination(s)
	SendSMS(ctx context.Context, in *SMS, opts ...grpc.CallOption) (*empty.Empty, error)
}

type sMSAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewSMSAPIClient(cc grpc.ClientConnInterface) SMSAPIClient {
	return &sMSAPIClient{cc}
}

func (c *sMSAPIClient) SendSMS(ctx context.Context, in *SMS, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/gidyon.apis.SMSAPI/SendSMS", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SMSAPIServer is the server API for SMSAPI service.
// All implementations must embed UnimplementedSMSAPIServer
// for forward compatibility
type SMSAPIServer interface {
	// Send an sms to its destination(s)
	SendSMS(context.Context, *SMS) (*empty.Empty, error)
	mustEmbedUnimplementedSMSAPIServer()
}

// UnimplementedSMSAPIServer must be embedded to have forward compatible implementations.
type UnimplementedSMSAPIServer struct {
}

func (UnimplementedSMSAPIServer) SendSMS(context.Context, *SMS) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendSMS not implemented")
}
func (UnimplementedSMSAPIServer) mustEmbedUnimplementedSMSAPIServer() {}

// UnsafeSMSAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SMSAPIServer will
// result in compilation errors.
type UnsafeSMSAPIServer interface {
	mustEmbedUnimplementedSMSAPIServer()
}

func RegisterSMSAPIServer(s grpc.ServiceRegistrar, srv SMSAPIServer) {
	s.RegisterService(&_SMSAPI_serviceDesc, srv)
}

func _SMSAPI_SendSMS_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SMS)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SMSAPIServer).SendSMS(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.SMSAPI/SendSMS",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SMSAPIServer).SendSMS(ctx, req.(*SMS))
	}
	return interceptor(ctx, in, info, handler)
}

var _SMSAPI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gidyon.apis.SMSAPI",
	HandlerType: (*SMSAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendSMS",
			Handler:    _SMSAPI_SendSMS_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sms.proto",
}