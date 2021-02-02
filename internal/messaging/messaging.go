package messaging

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gidyon/micro/pkg/grpc/auth"
	"github.com/gidyon/micro/utils/errs"
	"github.com/gidyon/micro/utils/mdutil"
	"github.com/gidyon/services/pkg/api/messaging/call"
	"github.com/gidyon/services/pkg/api/messaging/emailing"
	"github.com/gidyon/services/pkg/api/messaging/pusher"
	"github.com/gidyon/services/pkg/api/messaging/sms"
	"github.com/gidyon/services/pkg/api/subscriber"
	"github.com/speps/go-hashids"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"

	"google.golang.org/grpc/grpclog"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/gidyon/services/pkg/api/messaging"
)

var emptyMsg = &empty.Empty{}

type messagingServer struct {
	messaging.UnimplementedMessagingServer
	*Options
}

// Options contains the parameters passed while calling NewMessagingServer
type Options struct {
	SQLDBWrites      *gorm.DB
	SQLDBReads       *gorm.DB
	Logger           grpclog.LoggerV2
	EmailSender      string
	EmailClient      emailing.EmailingClient
	CallClient       call.CallAPIClient
	PushClient       pusher.PushMessagingClient
	SMSClient        sms.SMSAPIClient
	SubscriberClient subscriber.SubscriberAPIClient
	PaginationHasher *hashids.HashID
	AuthAPI          auth.API
}

// NewMessagingServer is factory for creating MessagingServer APIs
func NewMessagingServer(ctx context.Context, opt *Options) (messaging.MessagingServer, error) {
	// Validation
	var err error
	switch {
	case ctx == nil:
		err = errors.New("context is required")
	case opt == nil:
		err = errs.NilObject("options")
	case opt.SQLDBWrites == nil:
		err = errors.New("sql writes is required")
	case opt.SQLDBReads == nil:
		err = errors.New("sql reads is required")
	case opt.PaginationHasher == nil:
		err = errs.MissingField("pagination PaginationHasher")
	case opt.AuthAPI == nil:
		err = errs.MissingField("authentication API")
	case opt.Logger == nil:
		err = errors.New("logger is required")
	case opt.EmailClient == nil:
		err = errors.New("email client is required")
	case opt.EmailSender == "":
		err = errors.New("email sender is required")
	case opt.PushClient == nil:
		err = errors.New("pusher client is required")
	case opt.SMSClient == nil:
		err = errors.New("sms client is required")
	case opt.CallClient == nil:
		err = errors.New("call client is required")
	case opt.SubscriberClient == nil:
		err = errors.New("subscriber client is required")
	}
	if err != nil {
		return nil, err
	}

	api := &messagingServer{
		Options: opt,
	}

	// Automigration
	if !api.SQLDBWrites.Migrator().HasTable(&Message{}) {
		err = api.SQLDBWrites.AutoMigrate(&Message{})
		if err != nil {
			return nil, fmt.Errorf("failed to automigrate: %v", err)
		}
	}

	return api, nil
}

func validateMessage(msg *messaging.Message) error {
	// Validation
	var err error
	switch {
	case msg == nil:
		err = errs.NilObject("message")
	case msg.UserId == "":
		err = errs.MissingField("user id")
	case msg.Title == "":
		err = errs.MissingField("title")
	case msg.Data == "":
		err = errs.MissingField("data")
	case len(msg.SendMethods) == 0:
		err = errs.MissingField("send methods")
	default:
	}
	return err
}

func (api *messagingServer) BroadCastMessage(
	ctx context.Context, req *messaging.BroadCastMessageRequest,
) (*empty.Empty, error) {
	var err error

	// Validatation
	switch {
	case req == nil:
		return nil, errs.NilObject("BroadCastMessageRequest")
	case len(req.Channels) == 0:
		return nil, errs.MissingField("topics")
	default:
		err = validateMessage(req.GetMessage())
		if err != nil {
			return nil, err
		}
	}

	// Authenticate the request
	err = api.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Send broadcast
	err = api.sendBroadCastMessage(ctx, req)
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to send broadcast message")
	}

	return emptyMsg, nil
}

func (api *messagingServer) sendBroadCastMessage(
	ctx context.Context, req *messaging.BroadCastMessageRequest,
) error {
	var (
		md, ok      = metadata.FromIncomingContext(ctx)
		pageSize    = int32(1000)
		pageToken   = ""
		nextResults = true
		msg         = req.GetMessage()
	)

	ctxGet := mdutil.AddFromCtx(ctx)

	for nextResults {
		// Get subscribers
		subscribersRes, err := api.SubscriberClient.ListSubscribers(
			ctxGet, &subscriber.ListSubscribersRequest{
				PageSize:  pageSize,
				PageToken: pageToken,
				Filter: &subscriber.ListSubscribersFilter{
					Channels: req.GetChannels(),
				},
			})
		if err != nil {
			return errs.WrapErrorWithMsg(err, "failed to fetch subscribers")
		}

		// Update page token
		pageToken = subscribersRes.GetNextPageToken()

		if len(subscribersRes.GetSubscribers()) < int(pageSize) {
			nextResults = false
		}

		// Send using anonymous goroutine
		go func(subscribers []*subscriber.Subscriber, msg *messaging.Message) {

			ctx2, cancel := context.WithCancel(context.Background())
			defer cancel()

			if ok {
				ctx2 = mdutil.AddMD(ctx2, md)
			}

			phones := make([]string, 0, len(subscribers))
			deviceTokens := make([]string, 0, len(subscribers))
			emails := make([]string, 0, len(subscribers))

			for _, subscriberPB := range subscribers {
				emails = append(emails, subscriberPB.GetEmail())
				deviceTokens = append(deviceTokens, subscriberPB.GetDeviceToken())
				phones = append(phones, subscriberPB.GetPhone())

				// Save message
				if msg.GetSave() {
					msg.UserId = subscriberPB.SubscriberId
					msgDB, err := GetMessageDB(msg)
					if err != nil {
						return
					}
					err = api.SQLDBWrites.Create(msgDB).Error
					if err != nil {
						api.Logger.Errorf("failed to save message model: %v", err)
						return
					}
				}
			}

			for _, sendMethod := range msg.GetSendMethods() {
				switch sendMethod {
				case messaging.SendMethod_SEND_METHOD_UNSPECIFIED:
				case messaging.SendMethod_EMAIL:
					sender := firstVal(req.GetSender().GetEmailSender(), msg.Details["sender"], api.EmailSender)
					displayName := firstVal(req.GetSender().GetEmailSender(), msg.Details["display_name"])
					if displayName != "" {
						sender = fmt.Sprintf("%s <%s>", displayName, sender)
					}
					_, err = api.EmailClient.SendEmail(ctx2, &emailing.SendEmailRequest{
						Email: &emailing.Email{
							Destinations:    emails,
							From:            sender,
							Subject:         msg.Title,
							Body:            msg.Data,
							BodyContentType: "text/html",
						},
					})
					if err != nil {
						api.Logger.Errorf("failed to send email message to destinations: %v", err)
					}
				case messaging.SendMethod_SMSV2:
					_, err = api.SMSClient.SendSMS(ctx2, &sms.SendSMSRequest{
						Sms: &sms.SMS{
							DestinationPhones: phones,
							Keyword:           msg.Title,
							Message:           msg.Data,
						},
						Auth: req.GetSmsAuth(),
					})
					if err != nil {
						api.Logger.Errorf("failed to send sms message to destinations: %v", err)
					}
				case messaging.SendMethod_CALL:
					_, err = api.CallClient.Call(ctx2, &call.CallPayload{
						DestinationPhones: phones,
						Keyword:           msg.Title,
						Message:           msg.Data,
					})
					if err != nil {
						api.Logger.Errorf("failed to call recipients: %v", err)
					}
				case messaging.SendMethod_PUSH:
					_, err = api.PushClient.SendPushMessage(ctx2, &pusher.PushMessage{
						DeviceTokens: deviceTokens,
						Title:        msg.Title,
						Message:      msg.Data,
						Details:      msg.Details,
					})
					if err != nil {
						api.Logger.Errorf("failed to send pusher message to recipients: %v", err)
					}
				}
			}
		}(subscribersRes.GetSubscribers(), msg)
	}

	return nil
}

func firstVal(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (api *messagingServer) SendMessage(
	ctx context.Context, sendReq *messaging.SendMessageRequest,
) (*messaging.SendMessageResponse, error) {
	// Authenticate request
	err := api.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case sendReq == nil:
		return nil, errs.NilObject("send request")
	case sendReq.Message == nil:
		return nil, errs.NilObject("message")
	default:
		err = validateMessage(sendReq.GetMessage())
		if err != nil {
			return nil, err
		}
	}

	msg := sendReq.Message

	ctxGet, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 10*time.Second)
	defer cancel()

	// Get subscriber
	subscriberPB, err := api.SubscriberClient.GetSubscriber(ctxGet, &subscriber.GetSubscriberRequest{
		SubscriberId: msg.UserId,
	}, grpc.WaitForReady(true))
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to get subscriber")
	}

	// Send message
	for _, sendMethod := range msg.GetSendMethods() {
		switch sendMethod {
		case messaging.SendMethod_SEND_METHOD_UNSPECIFIED:
		case messaging.SendMethod_EMAIL:
			sender := firstVal(sendReq.GetSender().GetEmailSender(), msg.Details["sender"], api.EmailSender)
			displayName := firstVal(sendReq.GetSender().GetEmailSender(), msg.Details["display_name"])
			if displayName != "" {
				sender = fmt.Sprintf("%s <%s>", displayName, sender)
			}

			_, err = api.EmailClient.SendEmail(ctxGet, &emailing.SendEmailRequest{
				Email: &emailing.Email{
					Destinations:    []string{subscriberPB.GetEmail()},
					From:            sender,
					Subject:         msg.Title,
					Body:            msg.Data,
					BodyContentType: "text/html",
				},
			}, grpc.WaitForReady(true))
			if err != nil {
				return nil, errs.WrapErrorWithMsg(err, "failed to send email")
			}
		case messaging.SendMethod_SMSV2:
			_, err = api.SMSClient.SendSMS(ctxGet, &sms.SendSMSRequest{
				Sms: &sms.SMS{
					DestinationPhones: []string{subscriberPB.GetPhone()},
					Keyword:           msg.Title,
					Message:           msg.Data,
				},
				Auth: sendReq.GetSmsAuth(),
			}, grpc.WaitForReady(true))
			if err != nil {
				return nil, errs.WrapErrorWithMsg(err, "failed to send sms")
			}
		case messaging.SendMethod_CALL:
			_, err = api.CallClient.Call(ctxGet, &call.CallPayload{
				DestinationPhones: []string{subscriberPB.GetPhone()},
				Keyword:           msg.Title,
				Message:           msg.Data,
			}, grpc.WaitForReady(true))
			if err != nil {
				return nil, errs.WrapErrorWithMsg(err, "failed to send call")
			}
		case messaging.SendMethod_PUSH:
			_, err = api.PushClient.SendPushMessage(ctxGet, &pusher.PushMessage{
				DeviceTokens: []string{subscriberPB.GetDeviceToken()},
				Title:        msg.Title,
				Message:      msg.Data,
				Details:      msg.Details,
			}, grpc.WaitForReady(true))
			if err != nil {
				return nil, errs.WrapErrorWithMsg(err, "failed to send pusher message")
			}
		}
	}

	var msgID uint

	// Save message
	if msg.GetSave() {
		msgDB, err := GetMessageDB(msg)
		if err != nil {
			return nil, err
		}
		err = api.SQLDBWrites.Create(msgDB).Error
		if err != nil {
			return nil, errs.WrapErrorWithMsg(err, "failed to save message")
		}
		msgID = msgDB.ID
	}

	return &messaging.SendMessageResponse{
		MessageId: fmt.Sprint(msgID),
	}, nil
}

const (
	defaultPageSize  = 10
	defaultPageToken = 1000000000
)

func (api *messagingServer) ListMessages(
	ctx context.Context, listReq *messaging.ListMessagesRequest,
) (*messaging.Messages, error) {
	// Authorize request
	_, err := api.AuthAPI.AuthorizeActorOrGroups(ctx, listReq.GetFilter().GetUserId(), auth.Admins()...)
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case listReq == nil:
		return nil, errs.NilObject("ListMessagesRequest")
	case listReq.GetFilter().GetUserId() != "":
		ID, err = strconv.Atoi(listReq.GetFilter().GetUserId())
		if err != nil {
			return nil, errs.IncorrectVal("user id")
		}
	}

	pageSize := listReq.GetPageSize()
	if pageSize <= 0 || pageSize > defaultPageSize {
		pageSize = defaultPageSize
	}

	var id uint
	pageToken := listReq.GetPageToken()
	if pageToken != "" {
		ids, err := api.PaginationHasher.DecodeInt64WithError(listReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		id = uint(ids[0])
	}

	db := api.SQLDBReads.Order("id DESC").Limit(int(pageSize))
	if id > 0 {
		db = db.Where("id<?", id)
	}

	if len(listReq.GetFilter().GetTypeFilters()) > 0 {
		types := make([]int8, 0)
		filter := true
		for _, msgType := range listReq.GetFilter().GetTypeFilters() {
			types = append(types, int8(msgType))
			if msgType == messaging.MessageType_ALL {
				filter = false
				break
			}
		}
		if filter {
			db = db.Where("type IN(?)", types)
		}
	}

	messagesDB := make([]*Message, 0, pageSize)

	if listReq.GetFilter().GetUserId() != "" {
		err = db.Find(&messagesDB, "user_id=?", ID).Error
	} else {
		err = db.Find(&messagesDB).Error
	}
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to fetch messages")
	}

	messagesPB := make([]*messaging.Message, 0, len(messagesDB))

	for _, messageDB := range messagesDB {
		messagePB, err := GetMessagePB(messageDB)
		if err != nil {
			return nil, err
		}

		messagesPB = append(messagesPB, messagePB)
		id = messageDB.ID
	}

	var token string
	if int(pageSize) == len(messagesDB) {
		// Next page token
		token, err = api.PaginationHasher.EncodeInt64([]int64{int64(id)})
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to generate page token")
		}
	}

	return &messaging.Messages{
		Messages:      messagesPB,
		NextPageToken: token,
	}, nil
}

func (api *messagingServer) ReadAll(
	ctx context.Context, readReq *messaging.MessageRequest,
) (*empty.Empty, error) {
	// Authorize request
	_, err := api.AuthAPI.AuthorizeActorOrGroups(ctx, readReq.GetUserId(), auth.Admins()...)
	if err != nil {
		return nil, err
	}

	var ID int

	// Validation
	switch {
	case readReq == nil:
		return nil, errs.NilObject("MessageRequest")
	case readReq.UserId == "":
		return nil, errs.MissingField("user id")
	default:
		ID, err = strconv.Atoi(readReq.UserId)
		if err != nil {
			return nil, errs.IncorrectVal("user id")
		}
	}

	// Update messages
	err = api.SQLDBWrites.Model(Message{}).Where("user_id=? AND seen=?", ID, false).
		Update("seen", true).Error
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to mark messages as read")
	}

	return emptyMsg, nil
}

func (api *messagingServer) GetNewMessagesCount(
	ctx context.Context, getReq *messaging.MessageRequest,
) (*messaging.NewMessagesCount, error) {
	var (
		ID  int
		err error
	)

	// Validation
	switch {
	case getReq == nil:
		return nil, errs.NilObject("MessageRequest")
	case getReq.UserId == "":
		return nil, errs.MissingField("user id")
	default:
		ID, err = strconv.Atoi(getReq.UserId)
		if err != nil {
			return nil, errs.IncorrectVal("user id")
		}
	}

	// Authorize request
	_, err = api.AuthAPI.AuthorizeActorOrGroups(ctx, getReq.UserId, auth.Admins()...)
	if err != nil {
		return nil, err
	}

	var count int64
	err = api.SQLDBWrites.Model(Message{}).Where("user_id=? AND seen=?", ID, false).
		Count(&count).Error
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to get new messages count")
	}

	return &messaging.NewMessagesCount{
		Count: int32(count),
	}, nil
}
