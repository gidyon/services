package subscriber

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/micro/v2/utils/mdutil"
	"github.com/gidyon/services/pkg/api/account"
	"gorm.io/gorm"

	"github.com/gidyon/services/pkg/api/channel"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"

	"github.com/gidyon/services/pkg/api/subscriber"
	"github.com/golang/protobuf/ptypes/empty"
)

type subscriberAPIServer struct {
	subscriber.UnimplementedSubscriberAPIServer
	*Options
}

// Options are parameters passed while calling NewSubscriberAPIServer
type Options struct {
	SQLDB         *gorm.DB
	Logger        grpclog.LoggerV2
	ChannelClient channel.ChannelAPIClient
	AccountClient account.AccountAPIClient
	AuthAPI       auth.API
}

// NewSubscriberAPIServer factory creates a subscriber API server
func NewSubscriberAPIServer(
	ctx context.Context, opt *Options,
) (subscriber.SubscriberAPIServer, error) {
	// Validation
	var err error
	switch {
	case opt == nil:
		err = errs.NilObject("options")
	case opt.SQLDB == nil:
		err = errs.NilObject("sqlDB")
	case opt.Logger == nil:
		err = errs.NilObject("logger")
	case opt.ChannelClient == nil:
		err = errs.NilObject("channel client")
	case opt.AccountClient == nil:
		err = errs.NilObject("accounts client")
	case opt.AuthAPI == nil:
		err = errs.MissingField("authentication API")
	}
	if err != nil {
		return nil, err
	}

	subscriberAPI := &subscriberAPIServer{
		Options: opt,
	}

	// Automigration
	if !subscriberAPI.SQLDB.Migrator().HasTable(&Subscriber{}) {
		err = subscriberAPI.SQLDB.AutoMigrate(&Subscriber{})
		if err != nil {
			return nil, fmt.Errorf("failed to automigrate subscriber table: %w", err)
		}
	}

	return subscriberAPI, nil
}

func (subscriberAPI *subscriberAPIServer) Subscribe(
	ctx context.Context, req *subscriber.SubscriberRequest,
) (*empty.Empty, error) {
	// Validate request
	var err error
	switch {
	case req == nil:
		return nil, errs.NilObject("request")
	case len(req.Channels) == 0:
		return nil, errs.MissingField("channel names")
	case req.SubscriberId == "":
		return nil, errs.MissingField("subscriber id")
	}

	// Authorize the request
	_, err = subscriberAPI.AuthAPI.AuthorizeActorOrGroup(ctx, req.SubscriberId, subscriberAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, err
	}

	subscribersDB := make([]*Subscriber, 0, len(req.Channels))

	for _, channel := range req.Channels {
		// Check if subscribed
		err = subscriberAPI.SQLDB.First(&Subscriber{}, "user_id = ? AND channel = ?", req.SubscriberId, channel).Error
		switch {
		case err == nil:
		case errors.Is(err, gorm.ErrRecordNotFound):
			subscribersDB = append(subscribersDB, &Subscriber{
				UserID:  req.SubscriberId,
				Channel: channel,
			})
		default:
			return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to check if subcriber is subscribe to channel")
		}
	}

	err = subscriberAPI.SQLDB.CreateInBatches(subscribersDB, len(req.Channels)+1).Error
	if err != nil {
		return nil, errs.FailedToSave("subscriber channels", err)
	}

	ctx2, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 10*time.Second)
	defer cancel()

	go func() {
		// Increment channel subscribers
		_, err = subscriberAPI.ChannelClient.IncrementSubscribers(ctx2, &channel.SubscribersRequest{
			ChannelNames: req.Channels,
		}, grpc.WaitForReady(true))
		if err != nil {
			subscriberAPI.Logger.Errorln("failed to increment channel subscribers: ", err)
		}
	}()

	return &empty.Empty{}, nil
}

func (subscriberAPI *subscriberAPIServer) Unsubscribe(
	ctx context.Context, unSubReq *subscriber.SubscriberRequest,
) (*empty.Empty, error) {
	// Validation
	var err error
	switch {
	case unSubReq == nil:
		return nil, errs.NilObject("unsubscribe request")
	case len(unSubReq.Channels) == 0:
		return nil, errs.MissingField("channels")
	case unSubReq.SubscriberId == "":
		return nil, errs.MissingField("subscriber id")
	}

	// Authorize the request
	_, err = subscriberAPI.AuthAPI.AuthorizeActorOrGroup(ctx, unSubReq.SubscriberId, subscriberAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, err
	}

	for _, channel := range unSubReq.Channels {
		// Delete the channel
		err = subscriberAPI.SQLDB.Delete(&Subscriber{}, "user_id = ? AND channel = ?", unSubReq.SubscriberId, channel).Error
		if err != nil {
			subscriberAPI.Logger.Errorln("failed to delete subscriber channel: ", err)
		}
	}

	ctx2, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 10*time.Second)
	defer cancel()

	go func() {
		// Decrement channel subscribers
		_, err = subscriberAPI.ChannelClient.DecrementSubscribers(ctx2, &channel.SubscribersRequest{
			ChannelNames: unSubReq.Channels,
		})
		if err != nil {
			subscriberAPI.Logger.Errorln("failed to decrement channel subscribers: ", err)
		}
	}()

	return &empty.Empty{}, nil
}

const defaultPageSize = 50

func (subscriberAPI *subscriberAPIServer) ListSubscribers(
	ctx context.Context, req *subscriber.ListSubscribersRequest,
) (*subscriber.ListSubscribersResponse, error) {
	// Request validation
	switch {
	case req == nil:
		return nil, errs.NilObject("ListSubscribersRequest")
	}

	// Authorize admin
	payload, err := subscriberAPI.AuthAPI.AuthorizeAdmin(ctx)
	if err != nil {
		return nil, err
	}

	pageSize := req.GetPageSize()
	if pageSize <= 0 || pageSize > defaultPageSize {
		pageSize = defaultPageSize
	}

	var ID uint
	pageToken := req.GetPageToken()
	if pageToken != "" {
		bs, err := base64.StdEncoding.DecodeString(req.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		v, err := strconv.ParseUint(string(bs), 10, 64)
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "incorrect page token")
		}
		ID = uint(v)
	}

	db := subscriberAPI.SQLDB.Model(&Subscriber{}).Limit(int(pageSize) + 1).Order("id DESC")
	if ID != 0 {
		db = db.Where("id<?", ID)
	}

	// Apply filters
	if req.Filter != nil {
		if len(req.Filter.Channels) > 0 {
			db = db.Group("user_id").Where("channel IN (?)", req.Filter.Channels)
		}
	}

	var collectionCount int64

	if pageToken == "" {
		err = db.Count(&collectionCount).Error
		if err != nil {
			return nil, errs.SQLQueryFailed(err, "count")
		}
	}

	subscriberDBs := make([]*Subscriber, 0, pageSize)

	err = db.Find(&subscriberDBs).Error
	switch {
	case err == nil:
	default:
		return nil, errs.FailedToFind("subscribers", err)
	}

	subscribersPB := make([]*subscriber.Subscriber, 0, len(subscriberDBs))

	ctxGet := mdutil.AddFromCtx(ctx)

	for index, subscriberDB := range subscriberDBs {
		if index == int(pageSize) {
			break
		}

		// Lets get the user
		pb, err := subscriberAPI.AccountClient.GetAccount(ctxGet, &account.GetAccountRequest{
			AccountId:  subscriberDB.UserID,
			Priviledge: subscriberAPI.AuthAPI.IsAdmin(payload.Group),
		})
		switch {
		case err == nil:
		case status.Code(err) == codes.NotFound:
			continue
		default:
			return nil, errs.WrapErrorWithMsg(err, "failed to get subscriber")
		}

		channels := make([]string, 0, 5)

		// Get channels
		err = subscriberAPI.SQLDB.Model(&Subscriber{}).Where("user_id = ?", subscriberDB.UserID).Select("channel").Distinct("channel").Scan(&channels).Error
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to get subcriber channels")
		}

		subscriberPB, err := GetSubscriberPB(pb, channels)
		if err != nil {
			return nil, err
		}

		subscribersPB = append(subscribersPB, subscriberPB)

		ID = subscriberDB.ID
	}

	var token string
	if len(subscriberDBs) > int(pageSize) {
		// Next page token
		token = base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(ID)))
	}

	return &subscriber.ListSubscribersResponse{
		NextPageToken:   token,
		Subscribers:     subscribersPB,
		CollectionCount: collectionCount,
	}, nil
}

func (subscriberAPI *subscriberAPIServer) GetSubscriber(
	ctx context.Context, req *subscriber.GetSubscriberRequest,
) (*subscriber.Subscriber, error) {
	// Request validation
	switch {
	case req == nil:
		return nil, errs.NilObject("request")
	case req.SubscriberId == "":
		return nil, errs.MissingField("subscriber id")
	}

	// Authorize the request
	payload, err := subscriberAPI.AuthAPI.AuthorizeActorOrGroup(ctx, req.SubscriberId, subscriberAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, err
	}

	channels := make([]string, 0, 5)

	err = subscriberAPI.SQLDB.Model(&Subscriber{}).Where("user_id = ?", req.SubscriberId).Select("channel").Distinct("channel").Scan(&channels).Error
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to get subcriber channels")
	}

	ctx, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 10*time.Second)
	defer cancel()

	// Get account details
	pb, err := subscriberAPI.AccountClient.GetAccount(ctx, &account.GetAccountRequest{
		AccountId:  req.SubscriberId,
		Priviledge: subscriberAPI.AuthAPI.IsAdmin(payload.Group),
	}, grpc.WaitForReady(true))
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to get susbcriber profile")
	}

	return GetSubscriberPB(pb, channels)
}
