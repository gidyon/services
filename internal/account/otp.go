package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/micro/v2/utils/mdutil"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/messaging"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

func (accountAPI *accountAPIServer) RequestOTP(
	ctx context.Context, req *account.RequestOTPRequest,
) (*empty.Empty, error) {
	var err error

	// Validation
	switch {
	case req == nil:
		return nil, errs.NilObject("RequestChangePrivateAccountRequest")
	case req.Phone == "":
		return nil, errs.MissingField("phone")
	}

	// GetAccount the user from database
	accountDB := &Account{}
	err = accountAPI.SQLDBWrites.
		First(accountDB, "phone=?", req.Phone).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.WrapMessagef(codes.NotFound, "account with phone %s does not exist", req.Phone)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	accountID := fmt.Sprint(accountDB.AccountID)

	uniqueNumber := randomdata.Number(100000, 999999)

	// Set token with expiration of 5 minutes
	err = accountAPI.RedisDBWrites.Set(
		ctx, updateToken(accountID), uniqueNumber, time.Duration(5*time.Minute),
	).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "SET")
	}

	// Update token
	ctxExt := mdutil.AddFromCtx(ctx)

	// Send message
	_, err = accountAPI.MessagingClient.SendMessage(ctxExt, &messaging.SendMessageRequest{
		Message: &messaging.Message{
			UserId:      accountID,
			Title:       "OTP Login",
			Data:        fmt.Sprintf("Login OTP code is %v", uniqueNumber),
			Save:        true,
			Type:        messaging.MessageType_INFO,
			SendMethods: []messaging.SendMethod{messaging.SendMethod_SMSV2},
		},
		SmsAuth: req.GetSmsAuth(),
	})
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to send otp to phone")
	}

	return &emptypb.Empty{}, nil
}

func (accountAPI *accountAPIServer) SignInOTP(
	ctx context.Context, req *account.SignInOTPRequest,
) (*account.SignInResponse, error) {
	var err error

	// Validation
	switch {
	case req == nil:
		return nil, errs.NilObject("RequestChangePrivateAccountRequest")
	case req.Phone == "":
		return nil, errs.MissingField("phone")
	case req.Otp == "":
		return nil, errs.MissingField("otp")
	}

	// Get the user from database
	accountDB := &Account{}
	err = accountAPI.SQLDBWrites.
		First(accountDB, "phone=?", req.Phone).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.WrapMessagef(codes.NotFound, "account with phone %s does not exist", req.Phone)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	accountID := fmt.Sprint(accountDB.AccountID)

	// Get otp
	otp, err := accountAPI.RedisDBWrites.Get(ctx, updateToken(accountID)).Result()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "GET")
	}

	// Compare otp
	if otp != req.Otp {
		return nil, errs.WrapMessage(codes.Unauthenticated, "otp do not match")
	}

	accountPB, err := GetAccountPB(accountDB)
	if err != nil {
		return nil, err
	}

	// Check that account is not blocked
	if accountPB.State == account.AccountState_BLOCKED {
		return nil, errs.WrapMessage(codes.FailedPrecondition, "account blocked")
	}

	return accountAPI.updateSession(ctx, accountDB, req.Group)
}
