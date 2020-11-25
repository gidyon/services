package longrunning

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gidyon/services/pkg/api/longrunning"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/encryption"
	"github.com/gidyon/services/pkg/utils/errs"
	redis "github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/speps/go-hashids"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

const longrunningTTL = time.Hour * 12 * 7

type longrunningAPIService struct {
	redisDB *redis.Client
	logger  grpclog.LoggerV2
	hasher  *hashids.HashID
	authAPI auth.Interface
}

// Options contains parameters passed to NewLongrunningAPIService
type Options struct {
	RedisClient   *redis.Client
	Logger        grpclog.LoggerV2
	JWTSigningKey []byte
}

// NewLongrunningAPIService is factory for creating LongrunningAPIServer singletons
func NewLongrunningAPIService(ctx context.Context, opt *Options) (longrunning.LongrunningAPIServer, error) {
	// Validation
	var err error
	switch {
	case ctx == nil:
		err = errs.NilObject("context")
	case opt == nil:
		err = errs.NilObject("options")
	case opt.RedisClient == nil:
		err = errs.NilObject("redis client")
	case opt.Logger == nil:
		err = errs.NilObject("logger")
	case opt.JWTSigningKey == nil:
		err = errs.NilObject("jwt key")
	}
	if err != nil {
		return nil, err
	}

	authAPI, err := auth.NewAPI(opt.JWTSigningKey, "Longrunning API", "users")
	if err != nil {
		return nil, err
	}

	hasher, err := encryption.NewHasher(string(opt.JWTSigningKey))
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash id: %v", err)
	}

	longrunningAPI := &longrunningAPIService{
		redisDB: opt.RedisClient,
		logger:  opt.Logger,
		hasher:  hasher,
		authAPI: authAPI,
	}

	return longrunningAPI, nil
}

func validateLongrunning(op *longrunning.Longrunning) error {
	var err error
	switch {
	case op == nil:
		err = errs.NilObject("longrunning")
	case op.UserId == "":
		err = errs.MissingField("user id")
	case op.Details == "":
		err = errs.MissingField("longrunning details")
	case op.Status == longrunning.LongrunningStatus_OPERATION_STATUS_UNSPECIFIED:
		err = errs.MissingField("longrunning status")
	}
	return err
}

func getUserOpList(userID string) string {
	return userID + ":longrunnings"
}

func getOpKey(operatioID string) string {
	return "longrunnings:" + operatioID
}

func (longrunningAPI *longrunningAPIService) saveLongrunning(ctx context.Context, op *longrunning.Longrunning) error {
	// Marshal longrunning to bytes
	bs, err := proto.Marshal(op)
	if err != nil {
		return errs.FromProtoMarshal(err, "longrunning")
	}

	tx := longrunningAPI.redisDB.TxPipeline()

	// Save longrunning to user list
	err = tx.LPush(ctx, getUserOpList(op.UserId), op.Id).Err()
	if err != nil {
		return errs.RedisCmdFailed(err, "lpush")
	}

	// Save in cache
	err = tx.Set(ctx, getOpKey(op.Id), bs, longrunningTTL).Err()
	if err != nil {
		return errs.RedisCmdFailed(err, "set")
	}

	// Save transaction
	_, err = tx.Exec()
	if err != nil {
		return errs.RedisCmdFailed(err, "exec")
	}

	return nil
}

func (longrunningAPI *longrunningAPIService) CreateLongrunning(
	ctx context.Context, createReq *longrunning.CreateLongrunningRequest,
) (*longrunning.Longrunning, error) {
	// Authenticate request
	err := longrunningAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case createReq == nil:
		return nil, errs.NilObject("CreateLongrunningRequest")
	default:
		err = validateLongrunning(createReq.Longrunning)
		if err != nil {
			return nil, err
		}
	}

	createReq.Longrunning.Id = uuid.New().String()

	err = longrunningAPI.saveLongrunning(ctx, createReq.Longrunning)
	if err != nil {
		return nil, err
	}

	return createReq.Longrunning, nil
}

func (longrunningAPI *longrunningAPIService) UpdateLongrunning(
	ctx context.Context, updateReq *longrunning.UpdateLongrunningRequest,
) (*longrunning.Longrunning, error) {
	// Authorization
	err := longrunningAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case updateReq == nil:
		return nil, errs.NilObject("UpdateLongrunningRequest")
	case updateReq.LongrunningId == "":
		err = errs.MissingField("operarion id")
	case updateReq.Result == "":
		err = errs.MissingField("longrunning result")
	case updateReq.Status == longrunning.LongrunningStatus_OPERATION_STATUS_UNSPECIFIED:
		err = errs.MissingField("longrunning status")
	}
	if err != nil {
		return nil, err
	}

	// Get the longrunning
	opStr, err := longrunningAPI.redisDB.Get(ctx, getOpKey(updateReq.LongrunningId)).Result()
	switch {
	case err == nil:
	case errors.Is(err, redis.Nil):
		return nil, errs.WrapMessagef(codes.NotFound, "longrunning with id %v not found", updateReq.LongrunningId)
	default:
		return nil, errs.RedisCmdFailed(err, "get")
	}

	opPB := &longrunning.Longrunning{}

	// Proto unmarshal
	err = proto.Unmarshal([]byte(opStr), opPB)
	if err != nil {
		return nil, errs.FromProtoUnMarshal(err, "longrunning")
	}

	opPB.Status = updateReq.Status
	opPB.Result = updateReq.Result

	// Proto marshal
	bs, err := proto.Marshal(opPB)
	if err != nil {
		return nil, errs.FromProtoMarshal(err, "longrunning")
	}

	// Save updated longrunning
	err = longrunningAPI.redisDB.Set(ctx, getOpKey(updateReq.LongrunningId), bs, longrunningTTL).Err()
	if err != nil {
		return nil, err
	}

	return opPB, nil
}

func (longrunningAPI *longrunningAPIService) DeleteLongrunning(
	ctx context.Context, delReq *longrunning.DeleteLongrunningRequest,
) (*empty.Empty, error) {
	// Authorize actor
	_, err := longrunningAPI.authAPI.AuthorizeActor(ctx, delReq.GetUserId())
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case delReq == nil:
		return nil, errs.NilObject("DeleteLongrunningRequest")
	case delReq.UserId == "":
		err = errs.MissingField("user id")
	case delReq.LongrunningId == "":
		err = errs.MissingField("longrunning id")
	}
	if err != nil {
		return nil, err
	}

	ops, err := longrunningAPI.redisDB.LRange(ctx, getUserOpList(delReq.UserId), 0, -1).Result()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "hdel")
	}

	for i, op := range ops {
		if op == delReq.LongrunningId {
			ops = append(ops[:i], ops[i+1:]...)
		}
	}

	err = longrunningAPI.redisDB.Del(ctx, getUserOpList(delReq.UserId)).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "del")
	}

	err = longrunningAPI.redisDB.Del(ctx, getOpKey(delReq.LongrunningId)).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "del")
	}

	if len(ops) > 0 {
		err = longrunningAPI.redisDB.LPush(ctx, getUserOpList(delReq.UserId), ops).Err()
		if err != nil {
			return nil, errs.RedisCmdFailed(err, "lpush")
		}
	}

	return &emptypb.Empty{}, nil
}

const defaultPageSize = 20

func (longrunningAPI *longrunningAPIService) ListLongrunnings(
	ctx context.Context, listReq *longrunning.ListLongrunningsRequest,
) (*longrunning.ListLongrunningsResponse, error) {
	// Authentication
	_, err := longrunningAPI.authAPI.AuthorizeActorOrGroups(ctx, listReq.GetFilter().GetUserId(), auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	userID := listReq.GetFilter().GetUserId()

	// Validation
	switch {
	case listReq == nil:
		return nil, errs.NilObject("ListLongrunningsRequest")
	case userID == "":
		return nil, errs.MissingField("user id")
	}

	pageSize := listReq.GetPageSize()
	if pageSize <= 0 || pageSize > defaultPageSize {
		pageSize = defaultPageSize
	}

	var id int64
	pageToken := listReq.GetPageToken()
	if pageToken != "" {
		ids, err := longrunningAPI.hasher.DecodeInt64WithError(listReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		id = int64(ids[0])
	}

	// Get longrunnings ids
	opKeys, err := longrunningAPI.redisDB.LRange(ctx, getUserOpList(userID), id, int64(pageSize)+id).Result()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "lrange")
	}

	longrunningsPB := make([]*longrunning.Longrunning, 0, len(opKeys))

	for _, val := range opKeys {
		// Get longrunning
		val, err := longrunningAPI.redisDB.Get(ctx, getOpKey(val)).Result()
		if err != nil {
			return nil, errs.RedisCmdFailed(err, "get")
		}

		opPB := &longrunning.Longrunning{}
		// Unmarshal longrunning
		err = proto.Unmarshal([]byte(val), opPB)
		if err != nil {
			return nil, errs.FromProtoUnMarshal(err, "longrunning")
		}

		longrunningsPB = append(longrunningsPB, opPB)
	}

	var token = pageToken
	if int(pageSize) == len(longrunningsPB) {
		// Next page token
		token, err = longrunningAPI.hasher.EncodeInt64([]int64{id + 1})
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to generate page token")
		}
	}

	return &longrunning.ListLongrunningsResponse{
		Longrunnings:  longrunningsPB,
		NextPageToken: token,
	}, nil
}

func (longrunningAPI *longrunningAPIService) GetLongrunning(
	ctx context.Context, getReq *longrunning.GetLongrunningRequest,
) (*longrunning.Longrunning, error) {
	// Authentication
	err := longrunningAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case getReq == nil:
		return nil, errs.NilObject("GetLongrunningRequest")
	case getReq.LongrunningId == "":
		err = errs.MissingField("longrunning id")
	}
	if err != nil {
		return nil, err
	}

	// Get longrunning
	val, err := longrunningAPI.redisDB.Get(ctx, getOpKey(getReq.LongrunningId)).Result()
	switch {
	case err == nil:
	case errors.Is(err, redis.Nil):
		return nil, errs.WrapMessagef(codes.NotFound, "longrunning with id %v not found", getReq.LongrunningId)
	default:
		return nil, errs.RedisCmdFailed(err, "get")
	}

	opPB := &longrunning.Longrunning{}
	// Unmarshal longrunning
	err = proto.Unmarshal([]byte(val), opPB)
	if err != nil {
		return nil, errs.FromProtoUnMarshal(err, "longrunning")
	}

	return opPB, nil
}
