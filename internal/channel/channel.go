package channel

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/codes"

	"github.com/speps/go-hashids"
	"google.golang.org/grpc/grpclog"

	"strings"

	"github.com/gidyon/services/pkg/api/channel"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/errs"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jinzhu/gorm"
)

const channelsPageSize = 20

type channelAPIServer struct {
	sqlDB   *gorm.DB
	logger  grpclog.LoggerV2
	authAPI auth.Interface
	hasher  *hashids.HashID
}

// Options contains parameters required while creating a channel API server
type Options struct {
	SQLDB         *gorm.DB
	Logger        grpclog.LoggerV2
	JWTSigningKey []byte
}

func newHasher(salt string) (*hashids.HashID, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 30

	return hashids.NewWithData(hd)
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
	case opt.SQLDB == nil:
		err = errs.NilObject("sqlDB")
	case opt.Logger == nil:
		err = errs.NilObject("logger")
	case opt.JWTSigningKey == nil:
		err = errs.MissingField("jwt signing key")
	}
	if err != nil {
		return nil, err
	}

	// Authentication API
	authAPI, err := auth.NewAPI(opt.JWTSigningKey)
	if err != nil {
		return nil, err
	}

	// Pagination hasher
	hasher, err := newHasher(string(opt.JWTSigningKey))
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash id: %v", err)
	}

	channelAPI := &channelAPIServer{
		sqlDB:   opt.SQLDB,
		logger:  opt.Logger,
		authAPI: authAPI,
		hasher:  hasher,
	}

	// Automigration
	err = channelAPI.sqlDB.AutoMigrate(&Channel{}).Error
	if err != nil {
		return nil, fmt.Errorf("failed to automigrate channels table: %w", err)
	}

	return channelAPI, nil
}

func (channelAPI *channelAPIServer) CreateChannel(
	ctx context.Context, createReq *channel.CreateChannelRequest,
) (*channel.CreateChannelResponse, error) {
	// Request must not be nil
	if createReq == nil {
		return nil, errs.NilObject("CreateChannelRequest")
	}

	// Authenticate the request
	err := channelAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validate the channel payload
	channelPB := createReq.GetChannel()
	switch {
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
	err = channelAPI.sqlDB.Create(&channelDB).Error
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
	// Request must not be nil
	if updateReq == nil {
		return nil, errs.NilObject("UpdateChannelRequest")
	}

	// Authorize the request; must be channel owner
	_, err := channelAPI.authAPI.AuthorizeActorOrGroup(ctx, updateReq.OwnerId, auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
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
	err = channelAPI.sqlDB.Model(channelDB).Where("id=?", ID).
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
	// Request must not be nil
	if delReq == nil {
		return nil, errs.NilObject("DeleteChannelRequest")
	}

	// Authorize the actor; must be channel owner or admin
	_, err := channelAPI.authAPI.AuthorizeActorOrGroup(ctx, delReq.OwnerId, auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case delReq.Id == "":
		return nil, errs.MissingField("channel id")
	default:
		ID, err = strconv.Atoi(delReq.Id)
		if err != nil {
			return nil, errs.IncorrectVal("channel id")
		}
	}

	// Soft delete channel in database
	err = channelAPI.sqlDB.Delete(&Channel{}, "id=?", ID).Error
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
	err := channelAPI.authAPI.AuthenticateRequest(ctx)
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
		ids, err := channelAPI.hasher.DecodeInt64WithError(listReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		ID = uint(ids[0])
	}

	channelsDB := make([]*Channel, 0, pageSize)

	db := channelAPI.sqlDB.Unscoped().Limit(pageSize).Order("id DESC")
	if ID != 0 {
		db = db.Where("id<?", ID)
	}

	err = db.Find(&channelsDB).Error
	switch {
	case err == nil:
	default:
		return nil, errs.SQLQueryFailed(err, "LIST")
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
		token, err = channelAPI.hasher.EncodeInt64([]int64{int64(ID)})
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
	// Request must not be nil
	if getReq == nil {
		return nil, errs.NilObject("GetChannelRequest")
	}

	// Authenticate the request
	err := channelAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case getReq.Id == "":
		return nil, errs.MissingField("channel id")
	default:
		ID, err = strconv.Atoi(getReq.Id)
		if err != nil {
			return nil, errs.IncorrectVal("channel id")
		}
	}

	channelDB := &Channel{}

	err = channelAPI.sqlDB.Find(channelDB, "id=?", ID).Error
	switch {
	case err == nil:
	case gorm.IsRecordNotFoundError(err):
		return nil, errs.DoesNotExist("channel", getReq.Id)
	default:
		return nil, errs.SQLQueryFailed(err, "FIND")
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
	// Request must not be nil
	if incReq == nil {
		return nil, errs.NilObject("SubscribersRequest")
	}

	// Authenticate request
	err := channelAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case incReq.Id == "":
		return nil, errs.MissingField("channel id")
	default:
		ID, err = strconv.Atoi(incReq.Id)
		if err != nil {
			return nil, errs.IncorrectVal("channel id")
		}
	}

	// Increment subscribers in database
	err = channelAPI.sqlDB.Table(channelsTable).Where("id=?", ID).
		Update("subscribers", gorm.Expr("subscribers + ?", 1)).Error
	if err != nil {
		return nil, errs.SQLQueryFailed(err, "UPDATE")
	}

	return &empty.Empty{}, nil
}

func (channelAPI *channelAPIServer) DecrementSubscribers(
	ctx context.Context, incReq *channel.SubscribersRequest,
) (*empty.Empty, error) {
	// Request must not be nil
	if incReq == nil {
		return nil, errs.NilObject("SubscribersRequest")
	}

	// Authenticate request
	err := channelAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case incReq.Id == "":
		return nil, errs.MissingField("channel id")
	default:
		ID, err = strconv.Atoi(incReq.Id)
		if err != nil {
			return nil, errs.IncorrectVal("channel id")
		}
	}

	// Decrement subscribers in database
	err = channelAPI.sqlDB.Table(channelsTable).Where("id=?", ID).
		Update("subscribers", gorm.Expr("subscribers - ?", 1)).Error
	if err != nil {
		return nil, errs.SQLQueryFailed(err, "UPDATE")
	}

	return &empty.Empty{}, nil
}
