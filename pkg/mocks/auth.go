package mocks

import (
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/stretchr/testify/mock"
)

// AuthAPIMock is auth API
type AuthAPIMock interface {
	auth.Interface
}

// AuthAPI is a fake authentication API
var AuthAPI = &mocks.AuthAPIMock{}

func init() {
	AuthAPI.On("AuthenticateRequestV2", mock.Anything).
		Return(&auth.Payload{Group: auth.AdminGroup()}, nil)
	AuthAPI.On("AuthenticateRequest", mock.Anything).
		Return(nil)
	AuthAPI.On("AuthorizeActor", mock.Anything, mock.Anything).
		Return(&auth.Payload{Group: auth.AdminGroup()}, nil)
	AuthAPI.On("AuthorizeGroups",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&auth.Payload{Group: auth.AdminGroup()}, nil)
	AuthAPI.On("AuthorizeStrict",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&auth.Payload{Group: auth.AdminGroup()}, nil)
	AuthAPI.On("AuthorizeActorOrGroups",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&auth.Payload{Group: auth.AdminGroup()}, nil)

	AuthAPI.On("GenToken", mock.Anything, mock.Anything, mock.Anything).
		Return("token", nil)
}
