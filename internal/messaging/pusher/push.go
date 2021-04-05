package push

import (
	"context"
	"fmt"

	"github.com/appleboy/go-fcm"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"

	push "github.com/gidyon/services/pkg/api/messaging/pusher"
	"google.golang.org/grpc/grpclog"
)

// firebase cloud messaging
type fcmClient interface {
	SendWithContext(ctx context.Context, msg *fcm.Message) (*fcm.Response, error)
	SendWithRetry(msg *fcm.Message, retryAttempts int) (*fcm.Response, error)
	Send(msg *fcm.Message) (*fcm.Response, error)
}

type pushAPIServer struct {
	push.UnimplementedPushMessagingServer
	fcmClient fcmClient
	*Options
}

// Options contains the parameters passed while calling NewPushMessagingServer
type Options struct {
	AuthAPI      auth.API
	FCMServerKey string
	Logger       grpclog.LoggerV2
}

// NewPushMessagingServer is factory for creating push messaging servers
func NewPushMessagingServer(ctx context.Context, opt *Options) (push.PushMessagingServer, error) {
	// Validation
	var err error
	switch {
	case ctx == nil:
		err = errs.NilObject("context")
	case opt == nil:
		err = errs.NilObject("options")
	case opt.Logger == nil:
		err = errs.NilObject("logger")
	case opt.AuthAPI == nil:
		err = errs.NilObject("auth api")
	case opt.FCMServerKey == "":
		err = errs.MissingField("fcm server key")
	}
	if err != nil {
		return nil, err
	}
	// FCM
	fcmClient, err := fcm.NewClient(opt.FCMServerKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create fcm client: %v", err)
	}

	// API
	pushAPI := &pushAPIServer{
		fcmClient: fcmClient,
		Options:   opt,
	}

	return pushAPI, nil
}

var soEmpty = &empty.Empty{}

func (api *pushAPIServer) SendPushMessage(
	ctx context.Context, pushMsg *push.PushMessage,
) (*empty.Empty, error) {
	// Request must not be nil
	if pushMsg == nil {
		return nil, errs.NilObject("PushMessage")
	}

	// Authenticate request
	err := api.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validate push
	err = validatePushMessage(pushMsg)
	if err != nil {
		return nil, err
	}

	// Convert details to map[string]interaface{}
	details := make(map[string]interface{})
	for key, detail := range pushMsg.Details {
		details[key] = detail
	}

	// Send push message
	go func() {
		_, err = api.fcmClient.SendWithContext(ctx, &fcm.Message{
			RegistrationIDs: pushMsg.DeviceTokens,
			Data:            details,
			Notification: &fcm.Notification{
				Title: pushMsg.Title,
				Body:  pushMsg.Message,
			},
		})
		if err != nil {
			api.Logger.Errorf("failed to send message: %v", err)
		}
	}()

	return soEmpty, nil
}

func validatePushMessage(pushMsg *push.PushMessage) error {
	var err error
	switch {
	case len(pushMsg.DeviceTokens) == 0:
		err = errs.MissingField("device tokens")
	case pushMsg.Title == "":
		err = errs.MissingField("title")
	case pushMsg.Message == "":
		err = errs.MissingField("message")
	case len(pushMsg.Details) == 0:
		err = errs.MissingField("details")
	}
	return err
}
