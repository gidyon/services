package mocks

import (
	"github.com/gidyon/services/pkg/api/messaging/emailing"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/mock"
)

// EmailAPIClientMock is a mock for emailing API client
type EmailAPIClientMock interface {
	emailing.EmailingClient
}

// EmailAPI is a fake emailing API
var EmailAPI = &mocks.EmailAPIClientMock{}

func init() {
	EmailAPI.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
}
