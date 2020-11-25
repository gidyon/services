// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package account

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

// AccountAPIClient is the client API for AccountAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AccountAPIClient interface {
	// Signs in a user into their account
	SignIn(ctx context.Context, in *SignInRequest, opts ...grpc.CallOption) (*SignInResponse, error)
	// Signs in a user using third parties like Google, Facebook, Twitter etc
	SignInExternal(ctx context.Context, in *SignInExternalRequest, opts ...grpc.CallOption) (*SignInResponse, error)
	// Fetch new JWT using refresh token and updates session
	RefreshSession(ctx context.Context, in *RefreshSessionRequest, opts ...grpc.CallOption) (*SignInResponse, error)
	// Creates an account for a new user
	CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*CreateAccountResponse, error)
	// Activates an account to being active
	ActivateAccount(ctx context.Context, in *ActivateAccountRequest, opts ...grpc.CallOption) (*ActivateAccountResponse, error)
	// Updates a user account
	UpdateAccount(ctx context.Context, in *UpdateAccountRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// Request to change private account information
	RequestChangePrivateAccount(ctx context.Context, in *RequestChangePrivateAccountRequest, opts ...grpc.CallOption) (*RequestChangePrivateAccountResponse, error)
	// Updates a user private account information
	UpdatePrivateAccount(ctx context.Context, in *UpdatePrivateAccountRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// Deletes a user account
	DeleteAccount(ctx context.Context, in *DeleteAccountRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// Retrieves a user account
	GetAccount(ctx context.Context, in *GetAccountRequest, opts ...grpc.CallOption) (*Account, error)
	//  Retrieves multiple user accounts
	BatchGetAccounts(ctx context.Context, in *BatchGetAccountsRequest, opts ...grpc.CallOption) (*BatchGetAccountsResponse, error)
	//  Retrieves deeply linked accounts
	GetLinkedAccounts(ctx context.Context, in *GetLinkedAccountsRequest, opts ...grpc.CallOption) (*GetLinkedAccountsResponse, error)
	// Checks if an account exists
	ExistAccount(ctx context.Context, in *ExistAccountRequest, opts ...grpc.CallOption) (*ExistAccountResponse, error)
	// Updates account
	AdminUpdateAccount(ctx context.Context, in *AdminUpdateAccountRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// Fetches collection of accounts
	ListAccounts(ctx context.Context, in *ListAccountsRequest, opts ...grpc.CallOption) (*Accounts, error)
	// Searches accounts and linked accounts
	SearchAccounts(ctx context.Context, in *SearchAccountsRequest, opts ...grpc.CallOption) (*Accounts, error)
}

type accountAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewAccountAPIClient(cc grpc.ClientConnInterface) AccountAPIClient {
	return &accountAPIClient{cc}
}

func (c *accountAPIClient) SignIn(ctx context.Context, in *SignInRequest, opts ...grpc.CallOption) (*SignInResponse, error) {
	out := new(SignInResponse)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/SignIn", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) SignInExternal(ctx context.Context, in *SignInExternalRequest, opts ...grpc.CallOption) (*SignInResponse, error) {
	out := new(SignInResponse)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/SignInExternal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) RefreshSession(ctx context.Context, in *RefreshSessionRequest, opts ...grpc.CallOption) (*SignInResponse, error) {
	out := new(SignInResponse)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/RefreshSession", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) CreateAccount(ctx context.Context, in *CreateAccountRequest, opts ...grpc.CallOption) (*CreateAccountResponse, error) {
	out := new(CreateAccountResponse)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/CreateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) ActivateAccount(ctx context.Context, in *ActivateAccountRequest, opts ...grpc.CallOption) (*ActivateAccountResponse, error) {
	out := new(ActivateAccountResponse)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/ActivateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) UpdateAccount(ctx context.Context, in *UpdateAccountRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/UpdateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) RequestChangePrivateAccount(ctx context.Context, in *RequestChangePrivateAccountRequest, opts ...grpc.CallOption) (*RequestChangePrivateAccountResponse, error) {
	out := new(RequestChangePrivateAccountResponse)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/RequestChangePrivateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) UpdatePrivateAccount(ctx context.Context, in *UpdatePrivateAccountRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/UpdatePrivateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) DeleteAccount(ctx context.Context, in *DeleteAccountRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/DeleteAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) GetAccount(ctx context.Context, in *GetAccountRequest, opts ...grpc.CallOption) (*Account, error) {
	out := new(Account)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/GetAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) BatchGetAccounts(ctx context.Context, in *BatchGetAccountsRequest, opts ...grpc.CallOption) (*BatchGetAccountsResponse, error) {
	out := new(BatchGetAccountsResponse)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/BatchGetAccounts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) GetLinkedAccounts(ctx context.Context, in *GetLinkedAccountsRequest, opts ...grpc.CallOption) (*GetLinkedAccountsResponse, error) {
	out := new(GetLinkedAccountsResponse)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/GetLinkedAccounts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) ExistAccount(ctx context.Context, in *ExistAccountRequest, opts ...grpc.CallOption) (*ExistAccountResponse, error) {
	out := new(ExistAccountResponse)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/ExistAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) AdminUpdateAccount(ctx context.Context, in *AdminUpdateAccountRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/AdminUpdateAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) ListAccounts(ctx context.Context, in *ListAccountsRequest, opts ...grpc.CallOption) (*Accounts, error) {
	out := new(Accounts)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/ListAccounts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountAPIClient) SearchAccounts(ctx context.Context, in *SearchAccountsRequest, opts ...grpc.CallOption) (*Accounts, error) {
	out := new(Accounts)
	err := c.cc.Invoke(ctx, "/gidyon.apis.AccountAPI/SearchAccounts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AccountAPIServer is the server API for AccountAPI service.
// All implementations must embed UnimplementedAccountAPIServer
// for forward compatibility
type AccountAPIServer interface {
	// Signs in a user into their account
	SignIn(context.Context, *SignInRequest) (*SignInResponse, error)
	// Signs in a user using third parties like Google, Facebook, Twitter etc
	SignInExternal(context.Context, *SignInExternalRequest) (*SignInResponse, error)
	// Fetch new JWT using refresh token and updates session
	RefreshSession(context.Context, *RefreshSessionRequest) (*SignInResponse, error)
	// Creates an account for a new user
	CreateAccount(context.Context, *CreateAccountRequest) (*CreateAccountResponse, error)
	// Activates an account to being active
	ActivateAccount(context.Context, *ActivateAccountRequest) (*ActivateAccountResponse, error)
	// Updates a user account
	UpdateAccount(context.Context, *UpdateAccountRequest) (*empty.Empty, error)
	// Request to change private account information
	RequestChangePrivateAccount(context.Context, *RequestChangePrivateAccountRequest) (*RequestChangePrivateAccountResponse, error)
	// Updates a user private account information
	UpdatePrivateAccount(context.Context, *UpdatePrivateAccountRequest) (*empty.Empty, error)
	// Deletes a user account
	DeleteAccount(context.Context, *DeleteAccountRequest) (*empty.Empty, error)
	// Retrieves a user account
	GetAccount(context.Context, *GetAccountRequest) (*Account, error)
	//  Retrieves multiple user accounts
	BatchGetAccounts(context.Context, *BatchGetAccountsRequest) (*BatchGetAccountsResponse, error)
	//  Retrieves deeply linked accounts
	GetLinkedAccounts(context.Context, *GetLinkedAccountsRequest) (*GetLinkedAccountsResponse, error)
	// Checks if an account exists
	ExistAccount(context.Context, *ExistAccountRequest) (*ExistAccountResponse, error)
	// Updates account
	AdminUpdateAccount(context.Context, *AdminUpdateAccountRequest) (*empty.Empty, error)
	// Fetches collection of accounts
	ListAccounts(context.Context, *ListAccountsRequest) (*Accounts, error)
	// Searches accounts and linked accounts
	SearchAccounts(context.Context, *SearchAccountsRequest) (*Accounts, error)
	mustEmbedUnimplementedAccountAPIServer()
}

// UnimplementedAccountAPIServer must be embedded to have forward compatible implementations.
type UnimplementedAccountAPIServer struct {
}

func (UnimplementedAccountAPIServer) SignIn(context.Context, *SignInRequest) (*SignInResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignIn not implemented")
}
func (UnimplementedAccountAPIServer) SignInExternal(context.Context, *SignInExternalRequest) (*SignInResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignInExternal not implemented")
}
func (UnimplementedAccountAPIServer) RefreshSession(context.Context, *RefreshSessionRequest) (*SignInResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshSession not implemented")
}
func (UnimplementedAccountAPIServer) CreateAccount(context.Context, *CreateAccountRequest) (*CreateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAccount not implemented")
}
func (UnimplementedAccountAPIServer) ActivateAccount(context.Context, *ActivateAccountRequest) (*ActivateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ActivateAccount not implemented")
}
func (UnimplementedAccountAPIServer) UpdateAccount(context.Context, *UpdateAccountRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAccount not implemented")
}
func (UnimplementedAccountAPIServer) RequestChangePrivateAccount(context.Context, *RequestChangePrivateAccountRequest) (*RequestChangePrivateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestChangePrivateAccount not implemented")
}
func (UnimplementedAccountAPIServer) UpdatePrivateAccount(context.Context, *UpdatePrivateAccountRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePrivateAccount not implemented")
}
func (UnimplementedAccountAPIServer) DeleteAccount(context.Context, *DeleteAccountRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAccount not implemented")
}
func (UnimplementedAccountAPIServer) GetAccount(context.Context, *GetAccountRequest) (*Account, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccount not implemented")
}
func (UnimplementedAccountAPIServer) BatchGetAccounts(context.Context, *BatchGetAccountsRequest) (*BatchGetAccountsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchGetAccounts not implemented")
}
func (UnimplementedAccountAPIServer) GetLinkedAccounts(context.Context, *GetLinkedAccountsRequest) (*GetLinkedAccountsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLinkedAccounts not implemented")
}
func (UnimplementedAccountAPIServer) ExistAccount(context.Context, *ExistAccountRequest) (*ExistAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExistAccount not implemented")
}
func (UnimplementedAccountAPIServer) AdminUpdateAccount(context.Context, *AdminUpdateAccountRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AdminUpdateAccount not implemented")
}
func (UnimplementedAccountAPIServer) ListAccounts(context.Context, *ListAccountsRequest) (*Accounts, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAccounts not implemented")
}
func (UnimplementedAccountAPIServer) SearchAccounts(context.Context, *SearchAccountsRequest) (*Accounts, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchAccounts not implemented")
}
func (UnimplementedAccountAPIServer) mustEmbedUnimplementedAccountAPIServer() {}

// UnsafeAccountAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AccountAPIServer will
// result in compilation errors.
type UnsafeAccountAPIServer interface {
	mustEmbedUnimplementedAccountAPIServer()
}

func RegisterAccountAPIServer(s grpc.ServiceRegistrar, srv AccountAPIServer) {
	s.RegisterService(&_AccountAPI_serviceDesc, srv)
}

func _AccountAPI_SignIn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignInRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).SignIn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/SignIn",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).SignIn(ctx, req.(*SignInRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_SignInExternal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignInExternalRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).SignInExternal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/SignInExternal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).SignInExternal(ctx, req.(*SignInExternalRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_RefreshSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefreshSessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).RefreshSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/RefreshSession",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).RefreshSession(ctx, req.(*RefreshSessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_CreateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).CreateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/CreateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).CreateAccount(ctx, req.(*CreateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_ActivateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ActivateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).ActivateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/ActivateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).ActivateAccount(ctx, req.(*ActivateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_UpdateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).UpdateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/UpdateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).UpdateAccount(ctx, req.(*UpdateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_RequestChangePrivateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestChangePrivateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).RequestChangePrivateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/RequestChangePrivateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).RequestChangePrivateAccount(ctx, req.(*RequestChangePrivateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_UpdatePrivateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdatePrivateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).UpdatePrivateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/UpdatePrivateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).UpdatePrivateAccount(ctx, req.(*UpdatePrivateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_DeleteAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).DeleteAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/DeleteAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).DeleteAccount(ctx, req.(*DeleteAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_GetAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).GetAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/GetAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).GetAccount(ctx, req.(*GetAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_BatchGetAccounts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchGetAccountsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).BatchGetAccounts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/BatchGetAccounts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).BatchGetAccounts(ctx, req.(*BatchGetAccountsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_GetLinkedAccounts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLinkedAccountsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).GetLinkedAccounts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/GetLinkedAccounts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).GetLinkedAccounts(ctx, req.(*GetLinkedAccountsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_ExistAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExistAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).ExistAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/ExistAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).ExistAccount(ctx, req.(*ExistAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_AdminUpdateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AdminUpdateAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).AdminUpdateAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/AdminUpdateAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).AdminUpdateAccount(ctx, req.(*AdminUpdateAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_ListAccounts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAccountsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).ListAccounts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/ListAccounts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).ListAccounts(ctx, req.(*ListAccountsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccountAPI_SearchAccounts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchAccountsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountAPIServer).SearchAccounts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gidyon.apis.AccountAPI/SearchAccounts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountAPIServer).SearchAccounts(ctx, req.(*SearchAccountsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _AccountAPI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gidyon.apis.AccountAPI",
	HandlerType: (*AccountAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SignIn",
			Handler:    _AccountAPI_SignIn_Handler,
		},
		{
			MethodName: "SignInExternal",
			Handler:    _AccountAPI_SignInExternal_Handler,
		},
		{
			MethodName: "RefreshSession",
			Handler:    _AccountAPI_RefreshSession_Handler,
		},
		{
			MethodName: "CreateAccount",
			Handler:    _AccountAPI_CreateAccount_Handler,
		},
		{
			MethodName: "ActivateAccount",
			Handler:    _AccountAPI_ActivateAccount_Handler,
		},
		{
			MethodName: "UpdateAccount",
			Handler:    _AccountAPI_UpdateAccount_Handler,
		},
		{
			MethodName: "RequestChangePrivateAccount",
			Handler:    _AccountAPI_RequestChangePrivateAccount_Handler,
		},
		{
			MethodName: "UpdatePrivateAccount",
			Handler:    _AccountAPI_UpdatePrivateAccount_Handler,
		},
		{
			MethodName: "DeleteAccount",
			Handler:    _AccountAPI_DeleteAccount_Handler,
		},
		{
			MethodName: "GetAccount",
			Handler:    _AccountAPI_GetAccount_Handler,
		},
		{
			MethodName: "BatchGetAccounts",
			Handler:    _AccountAPI_BatchGetAccounts_Handler,
		},
		{
			MethodName: "GetLinkedAccounts",
			Handler:    _AccountAPI_GetLinkedAccounts_Handler,
		},
		{
			MethodName: "ExistAccount",
			Handler:    _AccountAPI_ExistAccount_Handler,
		},
		{
			MethodName: "AdminUpdateAccount",
			Handler:    _AccountAPI_AdminUpdateAccount_Handler,
		},
		{
			MethodName: "ListAccounts",
			Handler:    _AccountAPI_ListAccounts_Handler,
		},
		{
			MethodName: "SearchAccounts",
			Handler:    _AccountAPI_SearchAccounts_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "account.proto",
}