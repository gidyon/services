package mocks

import (
	"github.com/gidyon/services/pkg/api/messaging/call"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/mock"
)

// CallAPIClientMock is a mock for call API client
type CallAPIClientMock interface {
	call.CallAPIClient
}

// CallAPI is a fake call API
var CallAPI = &mocks.CallAPIClientMock{}

func init() {
	CallAPI.On("Call", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
}
