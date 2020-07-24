package mocks

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/messaging"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/mock"
)

// MessagingAPIClientMock is a mock for messaging API client
type MessagingAPIClientMock interface {
	messaging.MessagingClient
}

// MessagingAPI is a fake messaging API
var MessagingAPI = &mocks.MessagingAPIClientMock{}

func init() {
	MessagingAPI.On("BroadCastMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(empty.Empty{}, nil)
	MessagingAPI.On("SendMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(&messaging.SendMessageResponse{MessageId: randomdata.RandStringRunes(32)}, nil)
	MessagingAPI.On("ListMessages", mock.Anything, mock.Anything, mock.Anything).
		Return(&messaging.Messages{}, nil)
	MessagingAPI.On("ReadAll", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
	MessagingAPI.On("GetNewMessagesCount", mock.Anything, mock.Anything, mock.Anything).
		Return(&messaging.NewMessagesCount{Count: 5}, nil)
}
