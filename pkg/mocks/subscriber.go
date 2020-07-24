package mocks

import (
	"fmt"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/services/pkg/api/subscriber"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/mock"
)

// SubscriberAPIClientMock is a mock for subscriber API client
type SubscriberAPIClientMock interface {
	subscriber.SubscriberAPIClient
}

// SubscriberAPI is a fake subscriber API
var SubscriberAPI = &mocks.SubscriberAPIClientMock{}

func init() {
	SubscriberAPI.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
	SubscriberAPI.On("Unsubscribe", mock.Anything, mock.Anything, mock.Anything).
		Return(&empty.Empty{}, nil)
	SubscriberAPI.On("ListSubscribers", mock.Anything, mock.Anything, mock.Anything).
		Return(&subscriber.ListSubscribersResponse{
			Subscribers: []*subscriber.Subscriber{
				fakeSubscriber(), fakeSubscriber(), fakeSubscriber(),
			},
		}, nil)
	SubscriberAPI.On("GetSubscriber", mock.Anything, mock.Anything, mock.Anything).
		Return(&subscriber.Subscriber{}, nil)
}

func fakeSubscriber() *subscriber.Subscriber {
	return &subscriber.Subscriber{
		SubscriberId: fmt.Sprint(randomdata.Number(100000, 999999)),
		Email:        randomdata.Email(),
		Phone:        randomdata.PhoneNumber(),
		ExternalId:   randomdata.RandStringRunes(10),
		DeviceToken:  randomdata.MacAddress(),
		Channels: []*subscriber.Channel{
			{Name: randomdata.Month(), ChannelId: fmt.Sprint(randomdata.Number(30, 99))},
			{Name: randomdata.Month(), ChannelId: fmt.Sprint(randomdata.Number(30, 99))},
			{Name: randomdata.Month(), ChannelId: fmt.Sprint(randomdata.Number(30, 99))},
		},
	}
}
