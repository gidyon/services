// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	messaging "github.com/gidyon/services/pkg/api/messaging"

	mock "github.com/stretchr/testify/mock"
)

// MessagingAPIClientMock is an autogenerated mock type for the MessagingAPIClientMock type
type MessagingAPIClientMock struct {
	mock.Mock
}

// BroadCastMessage provides a mock function with given fields: ctx, in, opts
func (_m *MessagingAPIClientMock) BroadCastMessage(ctx context.Context, in *messaging.BroadCastMessageRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *emptypb.Empty
	if rf, ok := ret.Get(0).(func(context.Context, *messaging.BroadCastMessageRequest, ...grpc.CallOption) *emptypb.Empty); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *messaging.BroadCastMessageRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNewMessagesCount provides a mock function with given fields: ctx, in, opts
func (_m *MessagingAPIClientMock) GetNewMessagesCount(ctx context.Context, in *messaging.MessageRequest, opts ...grpc.CallOption) (*messaging.NewMessagesCount, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *messaging.NewMessagesCount
	if rf, ok := ret.Get(0).(func(context.Context, *messaging.MessageRequest, ...grpc.CallOption) *messaging.NewMessagesCount); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*messaging.NewMessagesCount)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *messaging.MessageRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListMessages provides a mock function with given fields: ctx, in, opts
func (_m *MessagingAPIClientMock) ListMessages(ctx context.Context, in *messaging.ListMessagesRequest, opts ...grpc.CallOption) (*messaging.Messages, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *messaging.Messages
	if rf, ok := ret.Get(0).(func(context.Context, *messaging.ListMessagesRequest, ...grpc.CallOption) *messaging.Messages); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*messaging.Messages)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *messaging.ListMessagesRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadAll provides a mock function with given fields: ctx, in, opts
func (_m *MessagingAPIClientMock) ReadAll(ctx context.Context, in *messaging.MessageRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *emptypb.Empty
	if rf, ok := ret.Get(0).(func(context.Context, *messaging.MessageRequest, ...grpc.CallOption) *emptypb.Empty); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *messaging.MessageRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendMessage provides a mock function with given fields: ctx, in, opts
func (_m *MessagingAPIClientMock) SendMessage(ctx context.Context, in *messaging.SendMessageRequest, opts ...grpc.CallOption) (*messaging.SendMessageResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *messaging.SendMessageResponse
	if rf, ok := ret.Get(0).(func(context.Context, *messaging.SendMessageRequest, ...grpc.CallOption) *messaging.SendMessageResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*messaging.SendMessageResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *messaging.SendMessageRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
