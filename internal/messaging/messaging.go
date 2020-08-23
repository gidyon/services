package messaging

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gidyon/services/pkg/api/messaging/call"
	"github.com/gidyon/services/pkg/api/messaging/emailing"
	"github.com/gidyon/services/pkg/api/messaging/push"
	"github.com/gidyon/services/pkg/api/messaging/sms"
	"github.com/gidyon/services/pkg/api/subscriber"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/errs"
	"github.com/gidyon/services/pkg/utils/mdutil"
	"github.com/speps/go-hashids"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/grpclog"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/jinzhu/gorm"

	"github.com/gidyon/services/pkg/api/messaging"
)

var emptyMsg = &empty.Empty{}

type messagingServer struct {
	sqlDB            *gorm.DB
	hasher           *hashids.HashID
	logger           grpclog.LoggerV2
	authAPI          auth.Interface
	emailClient      emailing.EmailingClient
	emailSender      string
	pushClient       push.PushMessagingClient
	smsClient        sms.SMSAPIClient
	callClient       call.CallAPIClient
	subscriberClient subscriber.SubscriberAPIClient
}

// Options contains the parameters passed while calling NewMessagingServer
type Options struct {
	SQLDB            *gorm.DB
	Logger           grpclog.LoggerV2
	JWTSigningKey    []byte
	EmailSender      string
	EmailClient      emailing.EmailingClient
	CallClient       call.CallAPIClient
	PushClient       push.PushMessagingClient
	SMSClient        sms.SMSAPIClient
	SubscriberClient subscriber.SubscriberAPIClient
}

func newHasher(salt string) (*hashids.HashID, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 30

	return hashids.NewWithData(hd)
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
	case opt.SQLDB == nil:
		err = errors.New("sqlDB is required")
	case opt.Logger == nil:
		err = errors.New("logger is required")
	case opt.EmailClient == nil:
		err = errors.New("email client is required")
	case opt.EmailSender == "":
		err = errors.New("email sender is required")
	case opt.JWTSigningKey == nil:
		err = errors.New("jwt signing key is required")
	case opt.PushClient == nil:
		err = errors.New("push client is required")
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

	// Auth API
	authAPI, err := auth.NewAPI(opt.JWTSigningKey, "Messaging API", "users")
	if err != nil {
		return nil, err
	}

	// Pagination
	hasher, err := newHasher(string(opt.JWTSigningKey))
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash id: %v", err)
	}

	api := &messagingServer{
		sqlDB:            opt.SQLDB,
		hasher:           hasher,
		logger:           opt.Logger,
		emailClient:      opt.EmailClient,
		emailSender:      opt.EmailSender,
		pushClient:       opt.PushClient,
		smsClient:        opt.SMSClient,
		callClient:       opt.CallClient,
		subscriberClient: opt.SubscriberClient,
		authAPI:          authAPI,
	}

	// Automigration
	err = api.sqlDB.AutoMigrate(&Message{}).Error
	if err != nil {
		return nil, fmt.Errorf("failed to automigrate: %v", err)
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
	case len(msg.Details) == 0:
		err = errs.MissingField("payload")
	case len(msg.SendMethods) == 0:
		err = errs.MissingField("send methods")
	default:
		// send methods
		unknown := true
		for _, sendMethod := range msg.SendMethods {
			if sendMethod != messaging.SendMethod_SEND_METHOD_UNSPECIFIED {
				unknown = false
				break
			}
		}
		if unknown {
			return errs.MissingField("send methods")
		}

		// validate user id
		_, err = strconv.Atoi(msg.UserId)
		if err != nil {
			return errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse user id in message to integer")
		}
	}
	return err
}

func (api *messagingServer) BroadCastMessage(
	ctx context.Context, req *messaging.BroadCastMessageRequest,
) (*empty.Empty, error) {
	// Request must not be nil
	if req == nil {
		return nil, errs.NilObject("BroadCastMessageRequest")
	}

	// Authenticate the request
	err := api.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validate data
	switch {
	case len(req.Channels) == 0:
		return nil, errs.MissingField("topics")
	default:
		err = validateMessage(req.GetMessage())
		if err != nil {
			return nil, err
		}
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
		subscribersRes, err := api.subscriberClient.ListSubscribers(ctxGet, &subscriber.ListSubscribersRequest{
			Channels:  req.GetChannels(),
			PageSize:  pageSize,
			PageToken: pageToken,
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
					err = api.sqlDB.Create(msgDB).Error
					if err != nil {
						api.logger.Errorf("failed to save message model: %v", err)
						return
					}
				}
			}

			for _, sendMethod := range msg.GetSendMethods() {
				switch sendMethod {
				case messaging.SendMethod_SEND_METHOD_UNSPECIFIED:
				case messaging.SendMethod_EMAIL:
					_, err = api.emailClient.SendEmail(ctx2, &emailing.Email{
						Destinations:    emails,
						From:            api.emailSender,
						Subject:         msg.Title,
						Body:            msg.Data,
						BodyContentType: "text/html",
					})
					if err != nil {
						api.logger.Errorf("failed to send email message to destinations: %v", err)
					}
				case messaging.SendMethod_SMS:
					_, err = api.smsClient.SendSMS(ctx2, &sms.SMS{
						DestinationPhones: phones,
						Keyword:           msg.Title,
						Message:           msg.Data,
					})
					if err != nil {
						api.logger.Errorf("failed to send sms message to destinations: %v", err)
					}
				case messaging.SendMethod_CALL:
					_, err = api.callClient.Call(ctx2, &call.CallPayload{
						DestinationPhones: phones,
						Keyword:           msg.Title,
						Message:           msg.Data,
					})
					if err != nil {
						api.logger.Errorf("failed to call recipients: %v", err)
					}
				case messaging.SendMethod_PUSH:
					_, err = api.pushClient.SendPushMessage(ctx2, &push.PushMessage{
						DeviceTokens: deviceTokens,
						Title:        msg.Title,
						Message:      msg.Data,
						Details:      msg.Details,
					})
					if err != nil {
						api.logger.Errorf("failed to send push message to recipients: %v", err)
					}
				}
			}
		}(subscribersRes.GetSubscribers(), msg)
	}

	return nil
}

func (api *messagingServer) SendMessage(
	ctx context.Context, msg *messaging.Message,
) (*messaging.SendMessageResponse, error) {
	// Request must not be nil
	if msg == nil {
		return nil, errs.NilObject("Message")
	}

	// Authenticate request
	err := api.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	err = validateMessage(msg)
	if err != nil {
		return nil, err
	}

	ctxGet := mdutil.AddFromCtx(ctx)

	// Get subscriber
	subscriberPB, err := api.subscriberClient.GetSubscriber(ctxGet, &subscriber.GetSubscriberRequest{
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
			_, err = api.emailClient.SendEmail(ctxGet, &emailing.Email{
				Destinations:    []string{subscriberPB.GetEmail()},
				From:            api.emailSender,
				Subject:         msg.Title,
				Body:            msg.Data,
				BodyContentType: "text/html",
			}, grpc.WaitForReady(true))
			if err != nil {
				return nil, errs.WrapErrorWithMsg(err, "failed to send email")
			}
		case messaging.SendMethod_SMS:
			_, err = api.smsClient.SendSMS(ctxGet, &sms.SMS{
				DestinationPhones: []string{subscriberPB.GetPhone()},
				Keyword:           msg.Title,
				Message:           msg.Data,
			}, grpc.WaitForReady(true))
			if err != nil {
				return nil, errs.WrapErrorWithMsg(err, "failed to send sms")
			}
		case messaging.SendMethod_CALL:
			_, err = api.callClient.Call(ctxGet, &call.CallPayload{
				DestinationPhones: []string{subscriberPB.GetPhone()},
				Keyword:           msg.Title,
				Message:           msg.Data,
			}, grpc.WaitForReady(true))
			if err != nil {
				return nil, errs.WrapErrorWithMsg(err, "failed to send call")
			}
		case messaging.SendMethod_PUSH:
			_, err = api.pushClient.SendPushMessage(ctxGet, &push.PushMessage{
				DeviceTokens: []string{subscriberPB.GetDeviceToken()},
				Title:        msg.Title,
				Message:      msg.Data,
				Details:      msg.Details,
			}, grpc.WaitForReady(true))
			if err != nil {
				return nil, errs.WrapErrorWithMsg(err, "failed to send push message")
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
		err = api.sqlDB.Create(msgDB).Error
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
	// Requst must not be nil
	if listReq == nil {
		return nil, errs.NilObject("ListMessagesRequest")
	}

	// Authorize request
	_, err := api.authAPI.AuthorizeActorOrGroup(ctx, listReq.UserId, auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case listReq.UserId != "":
		ID, err = strconv.Atoi(listReq.UserId)
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
		ids, err := api.hasher.DecodeInt64WithError(listReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		id = uint(ids[0])
	}

	db := api.sqlDB.Order("id DESC").Limit(pageSize)
	if id > 0 {
		db = db.Where("id<?", id)
	}

	if len(listReq.TypeFilters) > 0 {
		types := make([]int8, 0)
		filter := true
		for _, msgType := range listReq.GetTypeFilters() {
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

	if listReq.UserId != "" {
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
		token, err = api.hasher.EncodeInt64([]int64{int64(id)})
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
	// Request must not be nil
	if readReq == nil {
		return nil, errs.NilObject("MessageRequest")
	}

	// Authorize request
	_, err := api.authAPI.AuthorizeActorOrGroup(ctx, readReq.UserId, auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	var ID int

	// Validation
	switch {
	case readReq.UserId == "":
		return nil, errs.MissingField("user id")
	default:
		ID, err = strconv.Atoi(readReq.UserId)
		if err != nil {
			return nil, errs.IncorrectVal("user id")
		}
	}

	// Update messages
	err = api.sqlDB.Model(Message{}).Where("user_id=? AND seen=?", ID, false).
		Update("seen", true).Error
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to mark messages as read")
	}

	return emptyMsg, nil
}

func (api *messagingServer) GetNewMessagesCount(
	ctx context.Context, getReq *messaging.MessageRequest,
) (*messaging.NewMessagesCount, error) {
	// Request must not be nil
	if getReq == nil {
		return nil, errs.NilObject("MessageRequest")
	}

	// Authorize request
	_, err := api.authAPI.AuthorizeActorOrGroup(ctx, getReq.UserId, auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	var ID int

	// Validation
	switch {
	case getReq.UserId == "":
		return nil, errs.MissingField("user id")
	default:
		ID, err = strconv.Atoi(getReq.UserId)
		if err != nil {
			return nil, errs.IncorrectVal("user id")
		}
	}

	var count int32
	err = api.sqlDB.Model(Message{}).Where("user_id=? AND seen=?", ID, false).
		Count(&count).Error
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to get new messages count")
	}

	return &messaging.NewMessagesCount{
		Count: count,
	}, nil
}
