package mocks

import (
	"github.com/gidyon/services/pkg/api/messaging/sms"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/mock"
)

// SMSAPIClientMock is a mock for sms API client
type SMSAPIClientMock interface {
	sms.SMSAPIClient
}

// SMSAPI is a fake sms API
var SMSAPI = &mocks.SMSAPIClientMock{}

func init() {
	SMSAPI.On("SendSMS", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
}
