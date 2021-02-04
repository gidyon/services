package channel

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"google.golang.org/grpc/codes"

	"github.com/speps/go-hashids"
	"google.golang.org/grpc/grpclog"
	"gorm.io/gorm"

	"strings"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/channel"
	"github.com/golang/protobuf/ptypes/empty"
)

const channelsPageSize = 20

type channelAPIServer struct {
	channel.UnimplementedChannelAPIServer
	logger grpclog.LoggerV2
	*Options
}

// Options contains parameters required while creating a channel API server
type Options struct {
	SQLDBWrites      *gorm.DB
	SQLDBReads       *gorm.DB
	Logger           grpclog.LoggerV2
	PaginationHasher *hashids.HashID
	AuthAPI          auth.API
}

// NewChannelAPIServer is factory for creating channel  APIs
func NewChannelAPIServer(
	ctx context.Context, opt *Options,
) (channel.ChannelAPIServer, error) {
	// Validation
	var err error
	switch {
	case ctx == nil:
		err = errs.NilObject("context")
	case opt == nil:
		err = errs.NilObject("options")
	case opt.SQLDBWrites == nil:
		err = errs.NilObject("sql db writes")
	case opt.SQLDBReads == nil:
		err = errs.NilObject("sql db reads")
	case opt.Logger == nil:
		err = errs.NilObject("logger")
	case opt.PaginationHasher == nil:
		err = errs.MissingField("pagination hasher")
	case opt.AuthAPI == nil:
		err = errs.MissingField("authentication API")
	}
	if err != nil {
		return nil, err
	}

	channelAPI := &channelAPIServer{
		logger:  opt.Logger,
		Options: opt,
	}

	// Automigration
	if !channelAPI.SQLDBWrites.Migrator().HasTable(&Channel{}) {
		err = channelAPI.SQLDBWrites.AutoMigrate(&Channel{})
		if err != nil {
			return nil, fmt.Errorf("failed to automigrate channels table: %w", err)
		}
	}

	return channelAPI, nil
}

func (channelAPI *channelAPIServer) CreateChannel(
	ctx context.Context, createReq *channel.CreateChannelRequest,
) (*channel.CreateChannelResponse, error) {
	// Authenticate the request
	err := channelAPI.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validate the channel payload
	channelPB := createReq.GetChannel()
	switch {
	case createReq == nil:
		return nil, errs.NilObject("CreateChannelRequest")
	case channelPB == nil:
		return nil, errs.NilObject("Channel")
	case strings.TrimSpace(channelPB.OwnerId) == "":
		return nil, errs.MissingField("Owner Id")
	case strings.TrimSpace(channelPB.Title) == "":
		return nil, errs.MissingField("Channel Title")
	case strings.TrimSpace(channelPB.Description) == "":
		return nil, errs.MissingField("Channel Description")
	}

	channelPB.Subscribers = 0

	channelDB, err := GetChannelDB(channelPB)
	if err != nil {
		return nil, err
	}

	// Save channel in database
	err = channelAPI.SQLDBWrites.Create(&channelDB).Error
	switch {
	case err == nil:
	default:
		return nil, errs.SQLQueryFailed(err, "CreateChannel")
	}

	return &channel.CreateChannelResponse{
		Id: fmt.Sprint(channelDB.ID),
	}, nil
}

func (channelAPI *channelAPIServer) UpdateChannel(
	ctx context.Context, updateReq *channel.UpdateChannelRequest,
) (*empty.Empty, error) {
	// Authorize the request; must be channel owner
	_, err := channelAPI.AuthAPI.AuthorizeActorOrGroup(ctx, updateReq.GetOwnerId(), channelAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case updateReq == nil:
		return nil, errs.NilObject("UpdateChannelRequest")
	case updateReq.Channel == nil:
		return nil, errs.NilObject("channel")
	case updateReq.Channel.Id == "":
		return nil, errs.MissingField("channel id")
	default:
		ID, err = strconv.Atoi(updateReq.Channel.Id)
		if err != nil {
			return nil, errs.IncorrectVal("channel id")
		}
	}

	channelDB, err := GetChannelDB(updateReq.Channel)
	if err != nil {
		return nil, err
	}

	// Update model
	err = channelAPI.SQLDBWrites.Model(channelDB).Where("id=?", ID).
		Omit("id, subscribers").Updates(channelDB).Error
	switch {
	case err == nil:
	default:
		return nil, errs.SQLQueryFailed(err, "UpdateChannel")
	}

	return &empty.Empty{}, nil
}

func (channelAPI *channelAPIServer) DeleteChannel(
	ctx context.Context, delReq *channel.DeleteChannelRequest,
) (*empty.Empty, error) {
	// Authorize the request; must be channel owner
	_, err := channelAPI.AuthAPI.AuthorizeActorOrGroup(ctx, delReq.GetOwnerId(), channelAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case delReq == nil:
		return nil, errs.NilObject("DeleteChannelRequest")
	case delReq.Id == "":
		return nil, errs.MissingField("channel id")
	default:
		ID, err = strconv.Atoi(delReq.Id)
		if err != nil {
			return nil, errs.IncorrectVal("channel id")
		}
	}

	// Soft delete channel in database
	err = channelAPI.SQLDBWrites.Delete(&Channel{}, "id=?", ID).Error
	if err != nil {
		return nil, errs.SQLQueryFailed(err, "DeleteChannel")
	}

	return &empty.Empty{}, nil
}

const defaultPageSize = 20

func (channelAPI *channelAPIServer) ListChannels(
	ctx context.Context, listReq *channel.ListChannelsRequest,
) (*channel.ListChannelsResponse, error) {
	// Request must not be nil
	if listReq == nil {
		return nil, errs.NilObject("ListChannelsRequest")
	}

	// Authenticate the request;
	err := channelAPI.AuthAPI.AuthenticateRequest(ctx)
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
		ids, err := channelAPI.PaginationHasher.DecodeInt64WithError(listReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		ID = uint(ids[0])
	}

	channelsDB := make([]*Channel, 0, pageSize)

	db := channelAPI.SQLDBReads.Unscoped().Limit(int(pageSize)).Order("id DESC")
	if ID != 0 {
		db = db.Where("id<?", ID)
	}

	err = db.Find(&channelsDB).Error
	switch {
	case err == nil:
	default:
		return nil, errs.FailedToUpdate("channels", err)
	}

	channelsPB := make([]*channel.Channel, 0, len(channelsDB))

	for _, channelDB := range channelsDB {
		channelPB, err := GetChannelPB(channelDB)
		if err != nil {
			return nil, err
		}
		channelsPB = append(channelsPB, channelPB)
		ID = channelDB.ID
	}

	var token = pageToken
	if int(pageSize) == len(channelsDB) {
		// Next page token
		token, err = channelAPI.PaginationHasher.EncodeInt64([]int64{int64(ID)})
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to generate page token")
		}
	}

	return &channel.ListChannelsResponse{
		NextPageToken: token,
		Channels:      channelsPB,
	}, nil
}

func (channelAPI *channelAPIServer) GetChannel(
	ctx context.Context, getReq *channel.GetChannelRequest,
) (*channel.Channel, error) {
	// Authenticate the request
	err := channelAPI.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case getReq == nil:
		return nil, errs.NilObject("GetChannelRequest")
	case getReq.Id == "":
		return nil, errs.MissingField("channel id")
	default:
		ID, err = strconv.Atoi(getReq.Id)
		if err != nil {
			return nil, errs.IncorrectVal("channel id")
		}
	}

	channelDB := &Channel{}

	err = channelAPI.SQLDBReads.First(channelDB, "id=?", ID).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.DoesNotExist("channel", getReq.Id)
	default:
		return nil, errs.FailedToFind("channel", err)
	}

	channelPB, err := GetChannelPB(channelDB)
	if err != nil {
		return nil, err
	}

	return channelPB, nil
}

func (channelAPI *channelAPIServer) IncrementSubscribers(
	ctx context.Context, incReq *channel.SubscribersRequest,
) (*empty.Empty, error) {
	// Authenticate request
	err := channelAPI.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case incReq == nil:
		return nil, errs.NilObject("SubscribersRequest")
	case len(incReq.ChannelNames) == 0:
		return nil, errs.MissingField("channel names")
	}

	// Increment subscribers in database
	err = channelAPI.SQLDBWrites.Table(channelsTable).Where("title IN (?)", incReq.ChannelNames).
		Update("subscribers", gorm.Expr("subscribers + ?", 1)).Error
	if err != nil {
		return nil, errs.FailedToUpdate("channel", err)
	}

	return &empty.Empty{}, nil
}

func (channelAPI *channelAPIServer) DecrementSubscribers(
	ctx context.Context, decReq *channel.SubscribersRequest,
) (*empty.Empty, error) {
	// Authenticate request
	err := channelAPI.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case decReq == nil:
		return nil, errs.NilObject("decrement request")
	case len(decReq.ChannelNames) == 0:
		return nil, errs.MissingField("channel names")
	}

	// Decrement subscribers in database
	err = channelAPI.SQLDBWrites.Table(channelsTable).Where("title IN (?)", decReq.ChannelNames).
		Update("subscribers", gorm.Expr("subscribers - ?", 1)).Error
	if err != nil {
		return nil, errs.FailedToUpdate("channel", err)
	}

	return &empty.Empty{}, nil
}
