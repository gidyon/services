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

type AccountOptions = account_app.Options

// NewAccountAPIServer creates account API server
func NewAccountAPIServer(ctx context.Context, opt *AccountOptions) (account.AccountAPIServer, error) {
	return account_app.NewAccountAPI(ctx, opt)
}

type ChannelOptions = channel_app.Options

// NewChannelAPIServer creates a channel API server
func NewChannelAPIServer(ctx context.Context, opt *ChannelOptions) (channel.ChannelAPIServer, error) {
	return channel_app.NewChannelAPIServer(ctx, opt)
}

type LongRunningOptions = longrunning_app.Options

// NewLongRunningAPIServer creates a channel API server
func NewLongRunningAPIServer(ctx context.Context, opt *LongRunningOptions) (longrunning.OperationAPIServer, error) {
	return longrunning_app.NewOperationAPIService(ctx, opt)
}

type MessagingOptions = messaging_app.Options

// NewMessagingAPIServer creates a messaging API server
func NewMessagingAPIServer(ctx context.Context, opt *MessagingOptions) (messaging.MessagingServer, error) {
	return messaging_app.NewMessagingServer(ctx, opt)
}

type CallOptions = call_app.Options

// NewCallAPIServer creates a call API server
func NewCallAPIServer(ctx context.Context, opt *CallOptions) (call.CallAPIServer, error) {
	return call_app.NewCallAPIServer(ctx, opt)
}

type EmailingOptions = emailing_app.Options

// NewEmailingAPIServer creates an emailing API server
func NewEmailingAPIServer(ctx context.Context, opt *EmailingOptions) (emailing.EmailingServer, error) {
	return emailing_app.NewEmailingAPIServer(ctx, opt)
}

type PusherOptions = pusher_app.Options

// NewEmailingAPIServer creates FCM pusher API server
func NewPusherAPIServer(ctx context.Context, opt *PusherOptions) (pusher.PushMessagingServer, error) {
	return pusher_app.NewPushMessagingServer(ctx, opt)
}

type SMSOptions = sms_app.Options

// NewSMSAPIServer creates outgoing sms API server
func NewSMSAPIServer(ctx context.Context, opt *SMSOptions) (sms.SMSAPIServer, error) {
	return sms_app.NewSMSAPIServer(ctx, opt)
}

type SettingsOptions = settings_app.Options

// NewSettingsAPIServer creates a settings API server
func NewSettingsAPIServer(ctx context.Context, opt *SettingsOptions) (settings.SettingsAPIServer, error) {
	return settings_app.NewSettingsAPI(ctx, opt)
}

type SubscriberOptions = subscriber_app.Options

// NewSubscriberAPIServer creates a subscriber API server
func NewSubscriberAPIServer(ctx context.Context, opt *SubscriberOptions) (subscriber.SubscriberAPIServer, error) {
	return subscriber_app.NewSubscriberAPIServer(ctx, opt)
}
