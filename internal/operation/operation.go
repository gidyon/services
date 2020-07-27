package operation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gidyon/services/pkg/api/operation"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/errs"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/speps/go-hashids"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

const operationTTL = time.Hour * 12 * 7

type operationAPIService struct {
	redisDB *redis.Client
	logger  grpclog.LoggerV2
	hasher  *hashids.HashID
	authAPI auth.Interface
}

// Options contains parameters passed to NewOperationAPIService
type Options struct {
	RedisClient   *redis.Client
	Logger        grpclog.LoggerV2
	JWTSigningKey []byte
}

func newHasher(salt string) (*hashids.HashID, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 30

	return hashids.NewWithData(hd)
}

// NewOperationAPIService is factory for creating OperationAPIServer singletons
func NewOperationAPIService(ctx context.Context, opt *Options) (operation.OperationAPIServer, error) {
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

	authAPI, err := auth.NewAPI(opt.JWTSigningKey)
	if err != nil {
		return nil, err
	}

	hasher, err := newHasher(string(opt.JWTSigningKey))
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash id: %v", err)
	}

	operationAPI := &operationAPIService{
		redisDB: opt.RedisClient,
		logger:  opt.Logger,
		hasher:  hasher,
		authAPI: authAPI,
	}

	return operationAPI, nil
}

func validateOperation(op *operation.Operation) error {
	var err error
	switch {
	case op == nil:
		err = errs.NilObject("operation")
	case op.UserId == "":
		err = errs.MissingField("user id")
	case op.Details == "":
		err = errs.MissingField("operation details")
	case op.Status == operation.OperationStatus_OPERATION_STATUS_UNKNOWN:
		err = errs.MissingField("operation status")
	}
	return err
}

func getUserOpList(userID string) string {
	return userID + ":operations"
}

func getOpKey(operatioID string) string {
	return "operations:" + operatioID
}

func (operationAPI *operationAPIService) saveOperation(ctx context.Context, op *operation.Operation) error {
	// Marshal operation to bytes
	bs, err := proto.Marshal(op)
	if err != nil {
		return errs.FromProtoMarshal(err, "operation")
	}

	tx := operationAPI.redisDB.TxPipeline()

	// Save operation to user list
	err = tx.LPush(getUserOpList(op.UserId), op.Id).Err()
	if err != nil {
		return errs.RedisCmdFailed(err, "lpush")
	}

	// Save in cache
	err = tx.Set(getOpKey(op.Id), bs, operationTTL).Err()
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

func (operationAPI *operationAPIService) CreateOperation(
	ctx context.Context, createReq *operation.CreateOperationRequest,
) (*operation.Operation, error) {
	// Request must not be nil
	if createReq == nil {
		return nil, errs.NilObject("CreateOperationRequest")
	}

	// Authenticate request
	err := operationAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	err = validateOperation(createReq.Operation)
	if err != nil {
		return nil, err
	}

	createReq.Operation.Id = uuid.New().String()

	err = operationAPI.saveOperation(ctx, createReq.Operation)
	if err != nil {
		return nil, err
	}

	return createReq.Operation, nil
}

func (operationAPI *operationAPIService) UpdateOperation(
	ctx context.Context, updateReq *operation.UpdateOperationRequest,
) (*operation.Operation, error) {
	// Request must not be nil
	if updateReq == nil {
		return nil, errs.NilObject("UpdateOperationRequest")
	}

	// Authorization
	err := operationAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case updateReq.OperationId == "":
		err = errs.MissingField("operarion id")
	case updateReq.Result == "":
		err = errs.MissingField("operation result")
	case updateReq.Status == operation.OperationStatus_OPERATION_STATUS_UNKNOWN:
		err = errs.MissingField("operation status")
	}
	if err != nil {
		return nil, err
	}

	// Get the operation
	opStr, err := operationAPI.redisDB.Get(getOpKey(updateReq.OperationId)).Result()
	switch {
	case err == nil:
	case errors.Is(err, redis.Nil):
		return nil, errs.WrapMessagef(codes.NotFound, "operation with id %v not found", updateReq.OperationId)
	default:
		return nil, errs.RedisCmdFailed(err, "get")
	}

	opPB := &operation.Operation{}

	// Proto unmarshal
	err = proto.Unmarshal([]byte(opStr), opPB)
	if err != nil {
		return nil, errs.FromProtoUnMarshal(err, "operation")
	}

	opPB.Status = updateReq.Status
	opPB.Result = updateReq.Result

	// Proto marshal
	bs, err := proto.Marshal(opPB)
	if err != nil {
		return nil, errs.FromProtoMarshal(err, "operation")
	}

	// Save updated operation
	err = operationAPI.redisDB.Set(getOpKey(updateReq.OperationId), bs, operationTTL).Err()
	if err != nil {
		return nil, err
	}

	return opPB, nil
}

func (operationAPI *operationAPIService) DeleteOperation(
	ctx context.Context, delReq *operation.DeleteOperationRequest,
) (*empty.Empty, error) {
	// Request must not be nil
	if delReq == nil {
		return nil, errs.NilObject("DeleteOperationRequest")
	}

	// Authorize actor
	_, err := operationAPI.authAPI.AuthorizeActor(ctx, delReq.UserId)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case delReq.UserId == "":
		err = errs.MissingField("user id")
	case delReq.OperationId == "":
		err = errs.MissingField("operation id")
	}
	if err != nil {
		return nil, err
	}

	ops, err := operationAPI.redisDB.LRange(getUserOpList(delReq.UserId), 0, -1).Result()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "hdel")
	}

	for i, op := range ops {
		if op == delReq.OperationId {
			ops = append(ops[:i], ops[i+1:]...)
		}
	}

	err = operationAPI.redisDB.Del(getUserOpList(delReq.UserId)).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "del")
	}

	err = operationAPI.redisDB.Del(getOpKey(delReq.OperationId)).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "del")
	}

	if len(ops) > 0 {
		err = operationAPI.redisDB.LPush(getUserOpList(delReq.UserId), ops).Err()
		if err != nil {
			return nil, errs.RedisCmdFailed(err, "lpush")
		}
	}

	return &emptypb.Empty{}, nil
}

const defaultPageSize = 20

func (operationAPI *operationAPIService) ListOperations(
	ctx context.Context, listReq *operation.ListOperationsRequest,
) (*operation.ListOperationsResponse, error) {
	// Request must not be nil
	if listReq == nil {
		return nil, errs.NilObject("ListOperationsRequest")
	}

	// Authentication
	_, err := operationAPI.authAPI.AuthorizeActor(ctx, listReq.UserId)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case listReq.UserId == "":
		err = errs.MissingField("user id")
	}
	if err != nil {
		return nil, err
	}

	pageSize := listReq.GetPageSize()
	if pageSize <= 0 || pageSize > defaultPageSize {
		pageSize = defaultPageSize
	}

	var id int64
	pageToken := listReq.GetPageToken()
	if pageToken != "" {
		ids, err := operationAPI.hasher.DecodeInt64WithError(listReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		id = int64(ids[0])
	}

	// Get operations ids
	opKeys, err := operationAPI.redisDB.LRange(getUserOpList(listReq.UserId), id, int64(pageSize)+id).Result()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "lrange")
	}

	operationsPB := make([]*operation.Operation, 0, len(opKeys))

	for _, val := range opKeys {
		// Get operation
		val, err := operationAPI.redisDB.Get(getOpKey(val)).Result()
		if err != nil {
			return nil, errs.RedisCmdFailed(err, "get")
		}

		opPB := &operation.Operation{}
		// Unmarshal operation
		err = proto.Unmarshal([]byte(val), opPB)
		if err != nil {
			return nil, errs.FromProtoUnMarshal(err, "operation")
		}

		operationsPB = append(operationsPB, opPB)
	}

	var token = pageToken
	if int(pageSize) == len(operationsPB) {
		// Next page token
		token, err = operationAPI.hasher.EncodeInt64([]int64{id + 1})
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to generate page token")
		}
	}

	return &operation.ListOperationsResponse{
		Operations:    operationsPB,
		NextPageToken: token,
	}, nil
}

func (operationAPI *operationAPIService) GetOperation(
	ctx context.Context, getReq *operation.GetOperationRequest,
) (*operation.Operation, error) {
	// Request must not be nil
	if getReq == nil {
		return nil, errs.NilObject("GetOperationRequest")
	}

	// Authentication
	err := operationAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case getReq.OperationId == "":
		err = errs.MissingField("operation id")
	}
	if err != nil {
		return nil, err
	}

	// Get operation
	val, err := operationAPI.redisDB.Get(getOpKey(getReq.OperationId)).Result()
	switch {
	case err == nil:
	case errors.Is(err, redis.Nil):
		return nil, errs.WrapMessagef(codes.NotFound, "operation with id %v not found", getReq.OperationId)
	default:
		return nil, errs.RedisCmdFailed(err, "get")
	}

	opPB := &operation.Operation{}
	// Unmarshal operation
	err = proto.Unmarshal([]byte(val), opPB)
	if err != nil {
		return nil, errs.FromProtoUnMarshal(err, "operation")
	}

	return opPB, nil
}
