package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
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

const defaultChannel = "public"

func (subscriberAPI *subscriberAPIServer) createSubscriber(ID int) error {
	channels, err := json.Marshal([]string{defaultChannel})
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

	// Check that account id and channelId is provided
	var (
		ID  int
		err error
	)
	switch {
	case subReq == nil:
		return nil, errs.NilObject("SubscriberRequest")
	case len(subReq.Channels) == 0:
		return nil, errs.MissingField("channel names")
	case subReq.SubscriberId == "":
		return nil, errs.MissingField("subscriber id")
	default:
		ID, err = strconv.Atoi(subReq.SubscriberId)
		if err != nil {
			return nil, errs.IncorrectVal("subscriber id")
		}
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

	channels := []string{}

	// safe json unmarshal
	if len(sub.Channels) > 0 {
		err = json.Unmarshal(sub.Channels, &channels)
		if err != nil {
			tx.Rollback()
			return nil, errs.FromJSONUnMarshal(err, "subscribers channels")
		}
	} else {
		channels = []string{defaultChannel}
	}

	channelsV2 := make([]string, 0, len(channels)+1)

	// Check if already subscribed
	for _, channel := range subReq.Channels {
		if inArray(channel, channels) {
			continue
		}
		channelsV2 = append(channelsV2, channel)
	}

	// Add channel
	channelsV2 = append(channelsV2, channels...)

	sub.Channels, err = json.Marshal(channelsV2)
	if err != nil {
		tx.Rollback()
		return nil, errs.FromJSONMarshal(err, "channels")
	}

	// Subscribe user to the channel
	err = tx.Table(subscribersTable).Where("id=?", ID).Select("channels").Updates(sub).Error
	if err != nil {
		tx.Rollback()
		return nil, errs.FailedToUpdate("subscriber", err)
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
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (subscriberAPI *subscriberAPIServer) Unsubscribe(
	ctx context.Context, unSubReq *subscriber.SubscriberRequest,
) (*empty.Empty, error) {

	// Validation
	var (
		ID        int
		err       error
		accountID = unSubReq.GetSubscriberId()
	)
	switch {
	case unSubReq == nil:
		return nil, errs.NilObject("unsubscribe request")
	case len(unSubReq.Channels) == 0:
		return nil, errs.MissingField("channels")
	case unSubReq.SubscriberId == "":
		return nil, errs.MissingField("subscriber id")
	default:
		ID, err = strconv.Atoi(accountID)
		if err != nil {
			return nil, errs.IncorrectVal("subscriber id")
		}
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

	channels := []string{}

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
	for pos, ch := range unSubReq.Channels {
		if inArray(ch, channels) {
			// Remove with append
			channels = append(channels[:pos], channels[pos+1:]...)
			found = true
			break
		}
	}

	if !found {
		return &empty.Empty{}, nil
	}

	if len(sub.Channels) > 0 {
		sub.Channels, err = json.Marshal(channels)
		if err != nil {
			tx.Rollback()
			return nil, errs.FromJSONMarshal(err, "channels")
		}
	}

	// Unsubscribe user from the channel
	err = tx.Table(subscribersTable).Where("id=?", accountID).Updates(sub).Error
	if err != nil {
		tx.Rollback()
		return nil, errs.FailedToUpdate("subscriber", err)
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
		return nil, err
	}

	return &empty.Empty{}, nil
}

const defaultPageSize = 50

func inArray(v string, arr []string) bool {
	for _, v2 := range arr {
		if v == v2 {
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

	subscribersDBAll := make([]*Subscriber, 0, pageSize)

	if len(listReq.GetFilter().GetChannels()) > 0 {
		var (
			wg            = &sync.WaitGroup{}
			mu            = sync.Mutex{} // guards subscribersDB
			subscribersDB = make([]*Subscriber, 0, pageSize)
		)

		// Get subscriber for each channel
		for _, channel := range listReq.Filter.GetChannels() {
			wg.Add(1)

			channel := channel

			go func() {
				defer wg.Done()

				db := subscriberAPI.SQLDB.Limit(int(pageSize)).Order("id DESC")
				if ID != 0 {
					db = db.Where("id<?", ID)
				}

				err = db.Where("? MEMBER OF (channels)", channel).Find(&subscribersDB).Error
				if err != nil {
					subscriberAPI.Logger.Errorf("failed to finds subscribers: %v", err)
					return
				}

				mu.Lock()
				subscribersDBAll = append(subscribersDBAll, subscribersDB...)
				mu.Unlock()
			}()
		}

		// Collect all results
		wg.Wait()
	} else {
		db := subscriberAPI.SQLDB.Limit(int(pageSize)).Order("id DESC")
		if ID != 0 {
			db = db.Where("id<?", ID)
		}

		err = db.Find(&subscribersDBAll).Error
		switch {
		case err == nil:
		default:
			if err != nil {
				return nil, errs.SQLQueryFailed(err, "LIST")
			}
		}
	}

	var (
		subscribersPB = make([]*subscriber.Subscriber, 0, len(subscribersDBAll))
		seen          = make(map[uint]struct{}, int(pageSize))
	)

	ctxGet := mdutil.AddFromCtx(ctx)

	for _, subscriberDB := range subscribersDBAll {

		if _, ok := seen[subscriberDB.ID]; ok {
			continue
		}

		seen[subscriberDB.ID] = struct{}{}

		channels := make([]string, 0)

		if len(subscriberDB.Channels) > 0 {
			err = json.Unmarshal(subscriberDB.Channels, &channels)
			if err != nil {
				return nil, errs.FromJSONMarshal(err, "channels")
			}
		}

		// Lets get the user
		accountPB, err := subscriberAPI.AccountClient.GetAccount(ctxGet, &account.GetAccountRequest{
			AccountId:  fmt.Sprint(subscriberDB.ID),
			Priviledge: subscriberAPI.AuthAPI.IsAdmin(payload.Group),
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
	if int(pageSize) <= len(subscribersDBAll) {
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
	payload, err := subscriberAPI.AuthAPI.AuthorizeActorOrGroup(ctx, getReq.SubscriberId, subscriberAPI.AuthAPI.AdminGroups()...)
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
		Priviledge: subscriberAPI.AuthAPI.IsAdmin(payload.Group),
	}, grpc.WaitForReady(true))
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to get susbcriber profile")
	}

	return GetSubscriberPB(subscriberDB, accountPB)
}
