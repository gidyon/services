package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc"

	"github.com/gidyon/micro/pkg/grpc/auth"
	"github.com/gidyon/micro/utils/errs"
	"github.com/gidyon/micro/utils/mdutil"
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

func (subscriberAPI *subscriberAPIServer) createSubscriber(ID int) error {
	channels, err := json.Marshal([]*subscriber.ChannelSubcriber{
		{
			Name:      "public",
			ChannelId: "0",
		},
	})
	if err != nil {
		return errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to json marshal")
	}
	err = subscriberAPI.SQLDB.Create(&Subscriber{
		ID:       uint(ID),
		Channels: channels,
	}).Error
	if err != nil {
		return errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to create subscriber")
	}
	return nil
}

func (subscriberAPI *subscriberAPIServer) Subscribe(
	ctx context.Context, subReq *subscriber.SubscriberRequest,
) (*empty.Empty, error) {
	// Request must not be nil
	if subReq == nil {
		return nil, errs.NilObject("SubscriberRequest")
	}

	// Authorize the request
	_, err := subscriberAPI.AuthAPI.AuthorizeActorOrGroups(ctx, subReq.SubscriberId, auth.Admins()...)
	if err != nil {
		return nil, err
	}

	// Check that account id and channelId is provided
	var ID int
	switch {
	case subReq.ChannelId == "":
		return nil, errs.MissingField("channel id")
	case subReq.ChannelName == "":
		return nil, errs.MissingField("channel name")
	case subReq.SubscriberId == "":
		return nil, errs.MissingField("subscriber id")
	default:
		ID, err = strconv.Atoi(subReq.SubscriberId)
		if err != nil {
			return nil, errs.IncorrectVal("subscriber id")
		}
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

	sub := &Subscriber{}

	// Get user channels
	err = tx.First(sub, "id=?", ID).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		err = subscriberAPI.createSubscriber(ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	default:
		tx.Rollback()
		return nil, errs.SQLQueryFailed(err, "Get")
	}

	channels := []*subscriber.ChannelSubcriber{}

	// safe json unmarshal
	if len(sub.Channels) > 0 {
		err = json.Unmarshal(sub.Channels, &channels)
		if err != nil {
			tx.Rollback()
			return nil, errs.FromJSONUnMarshal(err, "Subscribers")
		}
	}

	// Check if already subscribed
	for _, channelPB := range channels {
		if channelPB.ChannelId == subReq.ChannelId {
			return &empty.Empty{}, nil
		}
	}

	// Add channel
	channels = append(channels, &subscriber.ChannelSubcriber{
		Name:      subReq.ChannelName,
		ChannelId: subReq.ChannelId,
	})

	sub.Channels, err = json.Marshal(channels)
	if err != nil {
		tx.Rollback()
		return nil, errs.FromJSONMarshal(err, "Subscribers")
	}

	// Subscribe user to the channel
	err = tx.Table(subscribersTable).Where("id=?", ID).
		Updates(sub).Error
	if err != nil {
		tx.Rollback()
		return nil, errs.SQLQueryFailed(err, "Subscribe")
	}

	ctx, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 10*time.Second)
	defer cancel()

	// Increment channel subscribers
	_, err = subscriberAPI.ChannelClient.IncrementSubscribers(ctx, &channel.SubscribersRequest{
		Id: subReq.ChannelId,
	}, grpc.WaitForReady(true))
	if err != nil {
		tx.Rollback()
		return nil, errs.WrapErrorWithMsg(err, "failed to increment channel subscriber")
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (subscriberAPI *subscriberAPIServer) Unsubscribe(
	ctx context.Context, unSubReq *subscriber.SubscriberRequest,
) (*empty.Empty, error) {
	// Request must not be nil
	if unSubReq == nil {
		return nil, errs.NilObject("UnSubscriberRequest")
	}

	// Authorize the request
	_, err := subscriberAPI.AuthAPI.AuthorizeActorOrGroups(ctx, unSubReq.SubscriberId, auth.Admins()...)
	if err != nil {
		return nil, err
	}

	accountID := unSubReq.GetSubscriberId()
	channelID := unSubReq.GetChannelId()

	// Validation
	var ID int
	switch {
	case channelID == "":
		return nil, errs.MissingField("channel id")
	case accountID == "":
		return nil, errs.MissingField("subscriber id")
	default:
		ID, err = strconv.Atoi(accountID)
		if err != nil {
			return nil, errs.IncorrectVal("subscriber id")
		}
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

	sub := &Subscriber{}

	// Get user channels
	err = tx.First(sub, "id=?", ID).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		err = subscriberAPI.createSubscriber(ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		tx.Commit()
		return &empty.Empty{}, nil
	default:
		tx.Rollback()
		return nil, errs.SQLQueryFailed(err, "Get")
	}

	channels := []*subscriber.ChannelSubcriber{}

	// safe json unmarshal
	if len(sub.Channels) > 0 {
		err = json.Unmarshal(sub.Channels, &channels)
		if err != nil {
			tx.Rollback()
			return nil, errs.FromJSONUnMarshal(err, "Subscribers")
		}
	}

	var found bool
	// Find the channel to unsubcribe
	for pos, ch := range channels {
		if channelID == ch.GetChannelId() {
			// Remove with append
			channels = append(channels[:pos], channels[pos+1:]...)
			found = true
		}
	}

	if !found {
		return &empty.Empty{}, nil
	}

	if len(sub.Channels) > 0 {
		sub.Channels, err = json.Marshal(channels)
		if err != nil {
			tx.Rollback()
			return nil, errs.FromJSONMarshal(err, "Subscribers")
		}
	}

	// Unsubscribe user from the channel
	err = tx.Table(subscribersTable).Where("id=?", accountID).Updates(sub).Error
	if err != nil {
		tx.Rollback()
		return nil, errs.SQLQueryFailed(err, "Unsubscribe")
	}

	ctx, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 10*time.Second)
	defer cancel()

	// Decrement channel subscribers
	_, err = subscriberAPI.ChannelClient.DecrementSubscribers(ctx, &channel.SubscribersRequest{
		Id: channelID,
	})
	if err != nil {
		tx.Rollback()
		return nil, errs.WrapErrorWithMsg(err, "failed to decrement channel subscribers")
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &empty.Empty{}, nil
}

const defaultPageSize = 50

func existChannel(channels []string, channel string) bool {
	for _, channel2 := range channels {
		if channel2 == channel {
			return true
		}
	}
	return false
}

func hasChannel(subscriberChannels []*subscriber.ChannelSubcriber, channels []string) bool {
	if len(channels) == 0 {
		return true
	}
	for _, channel := range subscriberChannels {
		if existChannel(channels, channel.ChannelId) {
			return true
		}
	}
	return false
}

func inGroup(group string, groups []string) bool {
	for _, grp := range groups {
		if grp == group {
			return true
		}
	}
	return false
}

func (subscriberAPI *subscriberAPIServer) ListSubscribers(
	ctx context.Context, listReq *subscriber.ListSubscribersRequest,
) (*subscriber.ListSubscribersResponse, error) {
	// Request must not be nil
	if listReq == nil {
		return nil, errs.NilObject("ListSubscribersRequest")
	}

	// Authorize the request
	payload, err := subscriberAPI.AuthAPI.AuthorizeGroups(ctx, auth.Admins()...)
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

	subscribersDB := make([]*Subscriber, 0, pageSize)

	db := subscriberAPI.SQLDB.Limit(int(pageSize)).Order("id DESC")
	if ID != 0 {
		db = db.Where("id<?", ID)
	}

	err = db.Find(&subscribersDB).Error
	switch {
	case err == nil:
	default:
		if err != nil {
			return nil, errs.SQLQueryFailed(err, "LIST")
		}
	}

	subscribersPB := make([]*subscriber.Subscriber, 0, len(subscribersDB))

	ctxGet := mdutil.AddFromCtx(ctx)

	for _, subscriberDB := range subscribersDB {

		if len(subscriberDB.Channels) == 0 {
			continue
		}

		channels := make([]*subscriber.ChannelSubcriber, 0)

		err = json.Unmarshal(subscriberDB.Channels, &channels)
		if err != nil {
			return nil, errs.FromJSONMarshal(err, "channels")
		}

		if !hasChannel(channels, listReq.GetFilter().GetChannels()) {
			continue
		}

		// Lets get the user
		accountPB, err := subscriberAPI.AccountClient.GetAccount(ctxGet, &account.GetAccountRequest{
			AccountId:  fmt.Sprint(subscriberDB.ID),
			Priviledge: inGroup(payload.Group, auth.Admins()),
		})
		switch {
		case err == nil:
		case status.Code(err) == codes.NotFound:
			continue
		default:
			return nil, errs.WrapErrorWithMsg(err, "failed to get subscriber")
		}

		subscriberPB, err := GetSubscriberPB(subscriberDB, accountPB)
		if err != nil {
			return nil, err
		}

		subscribersPB = append(subscribersPB, subscriberPB)

		ID = subscriberDB.ID
	}

	var token string
	if int(pageSize) == len(subscribersDB) {
		// Next page token
		token, err = subscriberAPI.PaginationHasher.EncodeInt64([]int64{int64(ID)})
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to generate page token")
		}
	}

	return &subscriber.ListSubscribersResponse{
		NextPageToken: token,
		Subscribers:   subscribersPB,
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
	payload, err := subscriberAPI.AuthAPI.AuthorizeActorOrGroups(ctx, getReq.SubscriberId, auth.Admins()...)
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case getReq.SubscriberId == "":
		return nil, errs.MissingField("subscriber id")
	default:
		ID, err = strconv.Atoi(getReq.SubscriberId)
		if err != nil {
			return nil, errs.IncorrectVal("subscriber id")
		}
	}

	// Get subscriber
	subscriberDB := &Subscriber{}
	err = subscriberAPI.SQLDB.First(subscriberDB, "id=?", ID).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		err = subscriberAPI.createSubscriber(ID)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errs.SQLQueryFailed(err, "GET")
	}

	ctx, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 10*time.Second)
	defer cancel()

	// Get account details
	accountPB, err := subscriberAPI.AccountClient.GetAccount(ctx, &account.GetAccountRequest{
		AccountId:  getReq.SubscriberId,
		Priviledge: inGroup(payload.Group, auth.Admins()),
	}, grpc.WaitForReady(true))
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to get susbcriber profile")
	}

	return GetSubscriberPB(subscriberDB, accountPB)
}
