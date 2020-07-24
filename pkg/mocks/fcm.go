package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/gidyon/services/pkg/mocks/mocks"

	"github.com/appleboy/go-fcm"
)

// FcmClient is firebase FCM
type FcmClient interface {
	SendWithContext(ctx context.Context, msg *fcm.Message) (*fcm.Response, error)
	SendWithRetry(msg *fcm.Message, retryAttempts int) (*fcm.Response, error)
	Send(msg *fcm.Message) (*fcm.Response, error)
}

// FCMAPI is mock for FCM client
var FCMAPI = &mocks.FcmClient{}

func init() {
	FCMAPI.On("SendWithContext", mock.Anything, mock.Anything).
		Return(&fcm.Response{}, nil)
	FCMAPI.On("SendWithRetry", mock.Anything, mock.Anything).
		Return(&fcm.Response{}, nil)
	FCMAPI.On("Send", mock.Anything, mock.Anything).
		Return(&fcm.Response{}, nil)
}
