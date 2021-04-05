package mocks

import (
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/stretchr/testify/mock"
)

// AuthAPIMock is auth API
type AuthAPIMock interface {
	auth.API
}

// AuthAPI is a fake authentication API
var AuthAPI = &mocks.AuthAPIMock{}

func init() {
	AuthAPI.On("AuthenticateRequestV2", mock.Anything).
		Return(&auth.Payload{Group: auth.DefaultAdminGroup()}, nil)
	AuthAPI.On("AuthenticateRequest", mock.Anything).
		Return(nil)
	AuthAPI.On("AuthorizeActor", mock.Anything, mock.Anything).
		Return(&auth.Payload{Group: auth.DefaultAdminGroup()}, nil)
	AuthAPI.On("AuthorizeGroups",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&auth.Payload{Group: auth.DefaultAdminGroup()}, nil)
	AuthAPI.On("AuthorizeStrict",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&auth.Payload{Group: auth.DefaultAdminGroup()}, nil)
	AuthAPI.On("AuthorizeActorOrGroup",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&auth.Payload{Group: auth.DefaultAdminGroup()}, nil)

	AuthAPI.On("GenToken", mock.Anything, mock.Anything, mock.Anything).
		Return("token", nil)
}
