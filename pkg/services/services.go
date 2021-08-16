package services

import (
	"context"

	account_app "github.com/gidyon/services/internal/account"
	channel_app "github.com/gidyon/services/internal/channel"
	longrunning_app "github.com/gidyon/services/internal/longrunning"
	messaging_app "github.com/gidyon/services/internal/messaging"
	call_app "github.com/gidyon/services/internal/messaging/call"
	emailing_app "github.com/gidyon/services/internal/messaging/emailing"
	pusher_app "github.com/gidyon/services/internal/messaging/pusher"
	sms_app "github.com/gidyon/services/internal/messaging/sms"
	settings_app "github.com/gidyon/services/internal/settings"
	subscriber_app "github.com/gidyon/services/internal/subscriber"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/channel"
	"github.com/gidyon/services/pkg/api/longrunning"
	"github.com/gidyon/services/pkg/api/messaging"
	"github.com/gidyon/services/pkg/api/messaging/call"
	"github.com/gidyon/services/pkg/api/messaging/emailing"
	"github.com/gidyon/services/pkg/api/messaging/pusher"
	"github.com/gidyon/services/pkg/api/messaging/sms"
	"github.com/gidyon/services/pkg/api/settings"
	"github.com/gidyon/services/pkg/api/subscriber"
)

// NewAccountAPIServer creates account API server
func NewAccountAPIServer(ctx context.Context, opt *account_app.Options) (account.AccountAPIServer, error) {
	return account_app.NewAccountAPI(ctx, opt)
}

// NewChannelAPIServer creates a channel API server
func NewChannelAPIServer(ctx context.Context, opt *channel_app.Options) (channel.ChannelAPIServer, error) {
	return channel_app.NewChannelAPIServer(ctx, opt)
}

// NewLongRunningAPIServer creates a channel API server
func NewLongRunningAPIServer(ctx context.Context, opt *longrunning_app.Options) (longrunning.OperationAPIServer, error) {
	return longrunning_app.NewOperationAPIService(ctx, opt)
}

// NewMessagingAPIServer creates a messaging API server
func NewMessagingAPIServer(ctx context.Context, opt *messaging_app.Options) (messaging.MessagingServer, error) {
	return messaging_app.NewMessagingServer(ctx, opt)
}

// NewCallAPIServer creates a call API server
func NewCallAPIServer(ctx context.Context, opt *call_app.Options) (call.CallAPIServer, error) {
	return call_app.NewCallAPIServer(ctx, opt)
}

// NewEmailingAPIServer creates an emailing API server
func NewEmailingAPIServer(ctx context.Context, opt *emailing_app.Options) (emailing.EmailingServer, error) {
	return emailing_app.NewEmailingAPIServer(ctx, opt)
}

// NewEmailingAPIServer creates FCM pusher API server
func NewPusherAPIServer(ctx context.Context, opt *pusher_app.Options) (pusher.PushMessagingServer, error) {
	return pusher_app.NewPushMessagingServer(ctx, opt)
}

// NewSMSAPIServer creates outgoing sms API server
func NewSMSAPIServer(ctx context.Context, opt *sms_app.Options) (sms.SMSAPIServer, error) {
	return sms_app.NewSMSAPIServer(ctx, opt)
}

// NewSettingsAPIServer creates a settings API server
func NewSettingsAPIServer(ctx context.Context, opt *settings_app.Options) (settings.SettingsAPIServer, error) {
	return settings_app.NewSettingsAPI(ctx, opt)
}

// NewSubscriberAPIServer creates a subscriber API server
func NewSubscriberAPIServer(ctx context.Context, opt *subscriber_app.Options) (subscriber.SubscriberAPIServer, error) {
	return subscriber_app.NewSubscriberAPIServer(ctx, opt)
}
