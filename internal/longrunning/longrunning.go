package longrunning

import (
	"context"
	"errors"
	"time"

	"github.com/gidyon/micro/pkg/grpc/auth"
	"github.com/gidyon/micro/utils/errs"
	"github.com/gidyon/services/pkg/api/longrunning"
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
	longrunning.UnimplementedOperationAPIServer
	*Options
}

// Options contains parameters passed to NewOperationAPIService
type Options struct {
	RedisClient      *redis.Client
	Logger           grpclog.LoggerV2
	PaginationHasher *hashids.HashID
	AuthAPI          auth.API
}

// NewOperationAPIService is factory for creating OperationAPIServer singletons
func NewOperationAPIService(ctx context.Context, opt *Options) (longrunning.OperationAPIServer, error) {
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
	case opt.AuthAPI == nil:
		err = errs.NilObject("auth api")
	case opt.PaginationHasher == nil:
		err = errs.NilObject("pagination PaginationHasher")
	}
	if err != nil {
		return nil, err
	}

	longrunningAPI := &longrunningAPIService{
		Options: opt,
	}

	return longrunningAPI, nil
}

func validateOperation(op *longrunning.Operation) error {
	var err error
	switch {
	case op == nil:
		err = errs.NilObject("longrunning")
	case op.UserId == "":
		err = errs.MissingField("user id")
	case op.Details == "":
		err = errs.MissingField("longrunning details")
	case op.Status == longrunning.OperationStatus_OPERATION_STATUS_UNSPECIFIED:
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

func (longrunningAPI *longrunningAPIService) saveOperation(ctx context.Context, op *longrunning.Operation) error {
	// Marshal longrunning to bytes
	bs, err := proto.Marshal(op)
	if err != nil {
		return errs.FromProtoMarshal(err, "longrunning")
	}

	tx := longrunningAPI.RedisClient.TxPipeline()

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
	_, err = tx.Exec(ctx)
	if err != nil {
		return errs.RedisCmdFailed(err, "exec")
	}

	return nil
}

func (longrunningAPI *longrunningAPIService) CreateOperation(
	ctx context.Context, createReq *longrunning.CreateOperationRequest,
) (*longrunning.Operation, error) {
	// Authenticate request
	err := longrunningAPI.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case createReq == nil:
		return nil, errs.NilObject("CreateOperationRequest")
	default:
		err = validateOperation(createReq.Operation)
		if err != nil {
			return nil, err
		}
	}

	createReq.Operation.Id = uuid.New().String()

	err = longrunningAPI.saveOperation(ctx, createReq.Operation)
	if err != nil {
		return nil, err
	}

	return createReq.Operation, nil
}

func (longrunningAPI *longrunningAPIService) UpdateOperation(
	ctx context.Context, updateReq *longrunning.UpdateOperationRequest,
) (*longrunning.Operation, error) {
	// Authorization
	err := longrunningAPI.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case updateReq == nil:
		return nil, errs.NilObject("UpdateOperationRequest")
	case updateReq.OperationId == "":
		err = errs.MissingField("operarion id")
	case updateReq.Result == "":
		err = errs.MissingField("longrunning result")
	case updateReq.Status == longrunning.OperationStatus_OPERATION_STATUS_UNSPECIFIED:
		err = errs.MissingField("longrunning status")
	}
	if err != nil {
		return nil, err
	}

	// Get the longrunning
	opStr, err := longrunningAPI.RedisClient.Get(ctx, getOpKey(updateReq.OperationId)).Result()
	switch {
	case err == nil:
	case errors.Is(err, redis.Nil):
		return nil, errs.WrapMessagef(codes.NotFound, "longrunning with id %v not found", updateReq.OperationId)
	default:
		return nil, errs.RedisCmdFailed(err, "get")
	}

	opPB := &longrunning.Operation{}

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
	err = longrunningAPI.RedisClient.Set(ctx, getOpKey(updateReq.OperationId), bs, longrunningTTL).Err()
	if err != nil {
		return nil, err
	}

	return opPB, nil
}

func (longrunningAPI *longrunningAPIService) DeleteOperation(
	ctx context.Context, delReq *longrunning.DeleteOperationRequest,
) (*empty.Empty, error) {
	// Authorize actor
	_, err := longrunningAPI.AuthAPI.AuthorizeActor(ctx, delReq.GetUserId())
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case delReq == nil:
		return nil, errs.NilObject("DeleteOperationRequest")
	case delReq.UserId == "":
		err = errs.MissingField("user id")
	case delReq.OperationId == "":
		err = errs.MissingField("longrunning id")
	}
	if err != nil {
		return nil, err
	}

	ops, err := longrunningAPI.RedisClient.LRange(ctx, getUserOpList(delReq.UserId), 0, -1).Result()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "hdel")
	}

	for i, op := range ops {
		if op == delReq.OperationId {
			ops = append(ops[:i], ops[i+1:]...)
		}
	}

	err = longrunningAPI.RedisClient.Del(ctx, getUserOpList(delReq.UserId)).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "del")
	}

	err = longrunningAPI.RedisClient.Del(ctx, getOpKey(delReq.OperationId)).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "del")
	}

	if len(ops) > 0 {
		err = longrunningAPI.RedisClient.LPush(ctx, getUserOpList(delReq.UserId), ops).Err()
		if err != nil {
			return nil, errs.RedisCmdFailed(err, "lpush")
		}
	}

	return &emptypb.Empty{}, nil
}

const defaultPageSize = 20

func (longrunningAPI *longrunningAPIService) ListOperations(
	ctx context.Context, listReq *longrunning.ListOperationsRequest,
) (*longrunning.ListOperationsResponse, error) {
	// Authentication
	_, err := longrunningAPI.AuthAPI.AuthorizeActorOrGroups(ctx, listReq.GetFilter().GetUserId(), auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	userID := listReq.GetFilter().GetUserId()

	// Validation
	switch {
	case listReq == nil:
		return nil, errs.NilObject("ListOperationsRequest")
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
		ids, err := longrunningAPI.PaginationHasher.DecodeInt64WithError(listReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		id = int64(ids[0])
	}

	// Get longrunnings ids
	opKeys, err := longrunningAPI.RedisClient.LRange(ctx, getUserOpList(userID), id, int64(pageSize)+id).Result()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "lrange")
	}

	longrunningsPB := make([]*longrunning.Operation, 0, len(opKeys))

	for _, val := range opKeys {
		// Get longrunning
		val, err := longrunningAPI.RedisClient.Get(ctx, getOpKey(val)).Result()
		if err != nil {
			return nil, errs.RedisCmdFailed(err, "get")
		}

		opPB := &longrunning.Operation{}
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
		token, err = longrunningAPI.PaginationHasher.EncodeInt64([]int64{id + 1})
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to generate page token")
		}
	}

	return &longrunning.ListOperationsResponse{
		Operations:    longrunningsPB,
		NextPageToken: token,
	}, nil
}

func (longrunningAPI *longrunningAPIService) GetOperation(
	ctx context.Context, getReq *longrunning.GetOperationRequest,
) (*longrunning.Operation, error) {
	// Authentication
	err := longrunningAPI.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case getReq == nil:
		return nil, errs.NilObject("GetOperationRequest")
	case getReq.OperationId == "":
		err = errs.MissingField("longrunning id")
	}
	if err != nil {
		return nil, err
	}

	// Get longrunning
	val, err := longrunningAPI.RedisClient.Get(ctx, getOpKey(getReq.OperationId)).Result()
	switch {
	case err == nil:
	case errors.Is(err, redis.Nil):
		return nil, errs.WrapMessagef(codes.NotFound, "longrunning with id %v not found", getReq.OperationId)
	default:
		return nil, errs.RedisCmdFailed(err, "get")
	}

	opPB := &longrunning.Operation{}
	// Unmarshal longrunning
	err = proto.Unmarshal([]byte(val), opPB)
	if err != nil {
		return nil, errs.FromProtoUnMarshal(err, "longrunning")
	}

	return opPB, nil
}
