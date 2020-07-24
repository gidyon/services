package mocks

import (
	"github.com/gidyon/services/pkg/api/messaging/push"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/mock"
)

// PushAPIClientMock is a mock for push API client
type PushAPIClientMock interface {
	push.PushMessagingClient
}

// PushAPI is a fake push API
var PushAPI = &mocks.PushAPIClientMock{}

func init() {
	PushAPI.On("SendPushMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
}
