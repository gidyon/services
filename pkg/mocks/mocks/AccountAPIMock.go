// Code generated by mockery v2.12.3. DO NOT EDIT.

package mocks

import (
	context "context"

	account "github.com/gidyon/services/pkg/api/account"

	emptypb "google.golang.org/protobuf/types/known/emptypb"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"
)

// AccountAPIMock is an autogenerated mock type for the AccountAPIMock type
type AccountAPIMock struct {
	mock.Mock
}

// ActivateAccount provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) ActivateAccount(ctx context.Context, in *account.ActivateAccountRequest, opts ...grpc.CallOption) (*account.ActivateAccountResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.ActivateAccountResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.ActivateAccountRequest, ...grpc.CallOption) *account.ActivateAccountResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.ActivateAccountResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.ActivateAccountRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ActivateAccountOTP provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) ActivateAccountOTP(ctx context.Context, in *account.ActivateAccountOTPRequest, opts ...grpc.CallOption) (*account.ActivateAccountResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.ActivateAccountResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.ActivateAccountOTPRequest, ...grpc.CallOption) *account.ActivateAccountResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.ActivateAccountResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.ActivateAccountOTPRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AdminUpdateAccount provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) AdminUpdateAccount(ctx context.Context, in *account.AdminUpdateAccountRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *emptypb.Empty
	if rf, ok := ret.Get(0).(func(context.Context, *account.AdminUpdateAccountRequest, ...grpc.CallOption) *emptypb.Empty); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.AdminUpdateAccountRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BatchGetAccounts provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) BatchGetAccounts(ctx context.Context, in *account.BatchGetAccountsRequest, opts ...grpc.CallOption) (*account.BatchGetAccountsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.BatchGetAccountsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.BatchGetAccountsRequest, ...grpc.CallOption) *account.BatchGetAccountsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.BatchGetAccountsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.BatchGetAccountsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateAccount provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) CreateAccount(ctx context.Context, in *account.CreateAccountRequest, opts ...grpc.CallOption) (*account.CreateAccountResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.CreateAccountResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.CreateAccountRequest, ...grpc.CallOption) *account.CreateAccountResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.CreateAccountResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.CreateAccountRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DailyRegisteredUsers provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) DailyRegisteredUsers(ctx context.Context, in *account.DailyRegisteredUsersRequest, opts ...grpc.CallOption) (*account.CountStats, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.CountStats
	if rf, ok := ret.Get(0).(func(context.Context, *account.DailyRegisteredUsersRequest, ...grpc.CallOption) *account.CountStats); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.CountStats)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.DailyRegisteredUsersRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteAccount provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) DeleteAccount(ctx context.Context, in *account.DeleteAccountRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *emptypb.Empty
	if rf, ok := ret.Get(0).(func(context.Context, *account.DeleteAccountRequest, ...grpc.CallOption) *emptypb.Empty); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.DeleteAccountRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExistAccount provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) ExistAccount(ctx context.Context, in *account.ExistAccountRequest, opts ...grpc.CallOption) (*account.ExistAccountResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.ExistAccountResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.ExistAccountRequest, ...grpc.CallOption) *account.ExistAccountResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.ExistAccountResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.ExistAccountRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAccount provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) GetAccount(ctx context.Context, in *account.GetAccountRequest, opts ...grpc.CallOption) (*account.Account, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.Account
	if rf, ok := ret.Get(0).(func(context.Context, *account.GetAccountRequest, ...grpc.CallOption) *account.Account); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.Account)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.GetAccountRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLinkedAccounts provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) GetLinkedAccounts(ctx context.Context, in *account.GetLinkedAccountsRequest, opts ...grpc.CallOption) (*account.GetLinkedAccountsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.GetLinkedAccountsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.GetLinkedAccountsRequest, ...grpc.CallOption) *account.GetLinkedAccountsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.GetLinkedAccountsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.GetLinkedAccountsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAccounts provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) ListAccounts(ctx context.Context, in *account.ListAccountsRequest, opts ...grpc.CallOption) (*account.Accounts, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.Accounts
	if rf, ok := ret.Get(0).(func(context.Context, *account.ListAccountsRequest, ...grpc.CallOption) *account.Accounts); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.Accounts)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.ListAccountsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RefreshSession provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) RefreshSession(ctx context.Context, in *account.RefreshSessionRequest, opts ...grpc.CallOption) (*account.SignInResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.SignInResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.RefreshSessionRequest, ...grpc.CallOption) *account.SignInResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.SignInResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.RefreshSessionRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RequestActivateAccountOTP provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) RequestActivateAccountOTP(ctx context.Context, in *account.RequestActivateAccountOTPRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *emptypb.Empty
	if rf, ok := ret.Get(0).(func(context.Context, *account.RequestActivateAccountOTPRequest, ...grpc.CallOption) *emptypb.Empty); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.RequestActivateAccountOTPRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RequestChangePrivateAccount provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) RequestChangePrivateAccount(ctx context.Context, in *account.RequestChangePrivateAccountRequest, opts ...grpc.CallOption) (*account.RequestChangePrivateAccountResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.RequestChangePrivateAccountResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.RequestChangePrivateAccountRequest, ...grpc.CallOption) *account.RequestChangePrivateAccountResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.RequestChangePrivateAccountResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.RequestChangePrivateAccountRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RequestSignInOTP provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) RequestSignInOTP(ctx context.Context, in *account.RequestSignInOTPRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *emptypb.Empty
	if rf, ok := ret.Get(0).(func(context.Context, *account.RequestSignInOTPRequest, ...grpc.CallOption) *emptypb.Empty); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.RequestSignInOTPRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SearchAccounts provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) SearchAccounts(ctx context.Context, in *account.SearchAccountsRequest, opts ...grpc.CallOption) (*account.Accounts, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.Accounts
	if rf, ok := ret.Get(0).(func(context.Context, *account.SearchAccountsRequest, ...grpc.CallOption) *account.Accounts); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.Accounts)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.SearchAccountsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignIn provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) SignIn(ctx context.Context, in *account.SignInRequest, opts ...grpc.CallOption) (*account.SignInResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.SignInResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.SignInRequest, ...grpc.CallOption) *account.SignInResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.SignInResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.SignInRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignInExternal provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) SignInExternal(ctx context.Context, in *account.SignInExternalRequest, opts ...grpc.CallOption) (*account.SignInResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.SignInResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.SignInExternalRequest, ...grpc.CallOption) *account.SignInResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.SignInResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.SignInExternalRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignInOTP provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) SignInOTP(ctx context.Context, in *account.SignInOTPRequest, opts ...grpc.CallOption) (*account.SignInResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *account.SignInResponse
	if rf, ok := ret.Get(0).(func(context.Context, *account.SignInOTPRequest, ...grpc.CallOption) *account.SignInResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*account.SignInResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.SignInOTPRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateAccount provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) UpdateAccount(ctx context.Context, in *account.UpdateAccountRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *emptypb.Empty
	if rf, ok := ret.Get(0).(func(context.Context, *account.UpdateAccountRequest, ...grpc.CallOption) *emptypb.Empty); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.UpdateAccountRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdatePrivateAccount provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) UpdatePrivateAccount(ctx context.Context, in *account.UpdatePrivateAccountRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *emptypb.Empty
	if rf, ok := ret.Get(0).(func(context.Context, *account.UpdatePrivateAccountRequest, ...grpc.CallOption) *emptypb.Empty); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.UpdatePrivateAccountRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdatePrivateAccountExternal provides a mock function with given fields: ctx, in, opts
func (_m *AccountAPIMock) UpdatePrivateAccountExternal(ctx context.Context, in *account.UpdatePrivateAccountExternalRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *emptypb.Empty
	if rf, ok := ret.Get(0).(func(context.Context, *account.UpdatePrivateAccountExternalRequest, ...grpc.CallOption) *emptypb.Empty); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *account.UpdatePrivateAccountExternalRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewAccountAPIMockT interface {
	mock.TestingT
	Cleanup(func())
}

// NewAccountAPIMock creates a new instance of AccountAPIMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAccountAPIMock(t NewAccountAPIMockT) *AccountAPIMock {
	mock := &AccountAPIMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
