package mocks

import (
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/mock"
)

// AccountAPIMock is mock for account API
type AccountAPIMock interface {
	account.AccountAPIClient
}

// AccountAPI is a fake authentication API
var AccountAPI = &mocks.AccountAPIMock{}

func init() {
	AccountAPI.On("SignIn", mock.Anything, mock.Anything, mock.Anything).
		Return(&account.SignInResponse{}, nil)
	AccountAPI.On("CreateAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(&account.CreateAccountResponse{}, nil)
	AccountAPI.On("ActivateAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(&account.ActivateAccountResponse{}, nil)
	AccountAPI.On("UpdateAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
	AccountAPI.On("RequestChangePrivateAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(&account.RequestChangePrivateAccountResponse{}, nil)
	AccountAPI.On("UpdatePrivateAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
	AccountAPI.On("DeleteAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
	AccountAPI.On("GetAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(&account.Account{}, nil)
	AccountAPI.On("BatchGetAccounts", mock.Anything, mock.Anything, mock.Anything).
		Return(&account.BatchGetAccountsResponse{}, nil)
	AccountAPI.On("ExistAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(&account.ExistAccountResponse{}, nil)
	AccountAPI.On("AdminUpdateAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
	AccountAPI.On("ListAccounts", mock.Anything, mock.Anything, mock.Anything).
		Return(&account.Accounts{}, nil)
	AccountAPI.On("SearchAccounts", mock.Anything, mock.Anything, mock.Anything).
		Return(&account.Accounts{}, nil)
}
