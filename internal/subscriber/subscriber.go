package subscriber

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/micro/v2/utils/mdutil"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/speps/go-hashids"
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
	SQLDB            *gorm.DB
	Logger           grpclog.LoggerV2
	ChannelClient    channel.ChannelAPIClient
	AccountClient    account.AccountAPIClient
	PaginationHasher *hashids.HashID
	AuthAPI          auth.API
}

// NewSubscriberAPIServer factory creates a subscriber API server
func NewSubscriberAPIServer(
	ctx context.Context, opt *Options,
) (subscriber.SubscriberAPIServer, error) {
	// Validation
	var err error
	switch {
	case ctx == nil:
		err = errs.NilObject("context")
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
	case opt.PaginationHasher == nil:
		err = errs.MissingField("pagination PaginationHasher")
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
	ctx context.Context, subReq *subscriber.SubscriberRequest,
) (*empty.Empty, error) {

	// Check that account id and channelId is provided
	var err error
	switch {
	case subReq == nil:
		return nil, errs.NilObject("SubscriberRequest")
	case len(subReq.Channels) == 0:
		return nil, errs.MissingField("channel names")
	case subReq.SubscriberId == "":
		return nil, errs.MissingField("subscriber id")
	}

	// Authorize the request
	_, err = subscriberAPI.AuthAPI.AuthorizeActorOrGroup(ctx, subReq.SubscriberId, subscriberAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, err
	}

	// Start a transaction
	tx := subscriberAPI.SQLDB.Begin()
	defer func() {
		if err := recover(); err != nil {
			subscriberAPI.Logger.Errorf("recovering from panic: %v", err)
		}
	}()

	if tx.Error != nil {
		tx.Rollback()
		return nil, errs.FailedToBeginTx(err)
	}

	subscribersDB := make([]*Subscriber, 0, len(subReq.Channels))

	for _, channel := range subReq.Channels {
		// Check if subscribed
		err = tx.First(&Subscriber{}, "user_id = ? AND channel = ?", subReq.SubscriberId, channel).Error
		switch {
		case err == nil:
		case errors.Is(err, gorm.ErrRecordNotFound):
			subscribersDB = append(subscribersDB, &Subscriber{
				UserID:  subReq.SubscriberId,
				Channel: channel,
			})
		default:
			tx.Rollback()
			return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to check if subcriber is subscribe to channel")
		}
	}

	err = tx.CreateInBatches(subscribersDB, len(subReq.Channels)+1).Error
	if err != nil {
		tx.Rollback()
		return nil, errs.SQLQueryFailed(err, "creating subscriber channel")
	}

	ctx, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 10*time.Second)
	defer cancel()

	// Increment channel subscribers
	_, err = subscriberAPI.ChannelClient.IncrementSubscribers(ctx, &channel.SubscribersRequest{
		ChannelNames: subReq.Channels,
	}, grpc.WaitForReady(true))
	if err != nil {
		tx.Rollback()
		return nil, errs.WrapErrorWithMsg(err, "failed to increment channel subscriber")
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, errs.FailedToCommitTx(err)
	}

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

	// Start a transaction
	tx := subscriberAPI.SQLDB.Begin()
	defer func() {
		if err := recover(); err != nil {
			subscriberAPI.Logger.Errorf("recovering from panic: %v", err)
		}
	}()

	if tx.Error != nil {
		tx.Rollback()
		return nil, errs.FailedToBeginTx(err)
	}

	for _, channel := range unSubReq.Channels {
		// Delete the channel
		err = tx.Delete(&Subscriber{}, "user_id = ? AND channel = ?", unSubReq.SubscriberId, channel).Error
		if err != nil {
			tx.Rollback()
			return nil, errs.FailedToDelete("subcriber channel", err)
		}
	}

	ctx, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 10*time.Second)
	defer cancel()

	// Decrement channel subscribers
	_, err = subscriberAPI.ChannelClient.DecrementSubscribers(ctx, &channel.SubscribersRequest{
		ChannelNames: unSubReq.Channels,
	})
	if err != nil {
		tx.Rollback()
		return nil, errs.WrapErrorWithMsg(err, "failed to decrement channel subscribers")
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, errs.FailedToCommitTx(err)
	}

	return &empty.Empty{}, nil
}

const defaultPageSize = 50

func (subscriberAPI *subscriberAPIServer) ListSubscribers(
	ctx context.Context, listReq *subscriber.ListSubscribersRequest,
) (*subscriber.ListSubscribersResponse, error) {
	// Request must not be nil
	if listReq == nil {
		return nil, errs.NilObject("ListSubscribersRequest")
	}

	// Authorize the request
	payload, err := subscriberAPI.AuthAPI.AuthorizeAdmin(ctx)
	if err != nil {
		return nil, err
	}

	pageSize := listReq.GetPageSize()
	if pageSize <= 0 || pageSize > defaultPageSize {
		pageSize = defaultPageSize
	}

	var ID uint
	pageToken := listReq.GetPageToken()
	if pageToken != "" {
		ids, err := subscriberAPI.PaginationHasher.DecodeInt64WithError(listReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapMessage(codes.InvalidArgument, "bad page token value")
		}
		ID = uint(ids[0])
	}

	db := subscriberAPI.SQLDB.Limit(int(pageSize)).Order("id DESC").Model(&Subscriber{})
	if ID != 0 {
		db = db.Where("id<?", ID)
	}

	// Apply filters
	if listReq.Filter != nil {
		if len(listReq.Filter.Channels) > 0 {
			db = db.Group("user_id").Where("channel IN (?)", listReq.Filter.Channels)
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
		return nil, errs.SQLQueryFailed(err, "Get subscribers")
	}

	subscribersPB := make([]*subscriber.Subscriber, 0, len(subscriberDBs))

	ctxGet := mdutil.AddFromCtx(ctx)

	for _, subscriberDB := range subscriberDBs {

		// Lets get the user
		accountPB, err := subscriberAPI.AccountClient.GetAccount(ctxGet, &account.GetAccountRequest{
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

		err = subscriberAPI.SQLDB.Model(&Subscriber{}).Where("user_id = ?", subscriberDB.UserID).Select("channel").Distinct("channel").Scan(&channels).Error
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to get subcriber channels")
		}

		subscriberPB, err := GetSubscriberPB(accountPB, channels)
		if err != nil {
			return nil, err
		}

		subscribersPB = append(subscribersPB, subscriberPB)

		ID = subscriberDB.ID
	}

	var token string
	if len(subscriberDBs) > int(pageSize) {
		// Next page token
		token, err = subscriberAPI.PaginationHasher.EncodeInt64([]int64{int64(ID)})
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to generate page token")
		}
	}

	return &subscriber.ListSubscribersResponse{
		NextPageToken:   token,
		Subscribers:     subscribersPB,
		CollectionCount: collectionCount,
	}, nil
}

func (subscriberAPI *subscriberAPIServer) GetSubscriber(
	ctx context.Context, getReq *subscriber.GetSubscriberRequest,
) (*subscriber.Subscriber, error) {
	// Request must not be nil
	if getReq == nil {
		return nil, errs.NilObject("GetSubscriberRequest")
	}

	// Authorize the request
	payload, err := subscriberAPI.AuthAPI.AuthorizeActorOrGroup(ctx, getReq.SubscriberId, subscriberAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case getReq.SubscriberId == "":
		return nil, errs.MissingField("subscriber id")
	default:
	}

	channels := make([]string, 0, 5)

	err = subscriberAPI.SQLDB.Model(&Subscriber{}).Where("user_id = ?", getReq.SubscriberId).Select("channel").Distinct("channel").Scan(&channels).Error
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to get subcriber channels")
	}

	ctx, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 10*time.Second)
	defer cancel()

	// Get account details
	accountPB, err := subscriberAPI.AccountClient.GetAccount(ctx, &account.GetAccountRequest{
		AccountId:  getReq.SubscriberId,
		Priviledge: subscriberAPI.AuthAPI.IsAdmin(payload.Group),
	}, grpc.WaitForReady(true))
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to get susbcriber profile")
	}

	return GetSubscriberPB(accountPB, channels)
}
