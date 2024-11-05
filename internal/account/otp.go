package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/messaging"
	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

func otpKey(accountID string) string {
	return "otplogin:" + accountID
}

func (accountAPI *accountAPIServer) RequestSignInOTP(
	ctx context.Context, req *account.RequestSignInOTPRequest,
) (*empty.Empty, error) {
	var err error

	// Validation
	switch {
	case req == nil:
		return nil, errs.NilObject("RequestChangePrivateAccountRequest")
	case req.Username == "":
		return nil, errs.MissingField("username")
	}

	// GetAccount the user from database
	db := &Account{}
	err = accountAPI.SQLDBWrites.
		First(db, "account_id = ?", req.Username).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.WrapMessagef(codes.NotFound, "account with id %s does not exist", req.Username)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	accountID := fmt.Sprint(db.AccountID)

	uniqueNumber := randomdata.Number(100000, 999999)

	// Set token with expiration of 5 minutes
	err = accountAPI.RedisDBWrites.Set(
		ctx, otpKey(accountID), uniqueNumber, time.Duration(5*time.Minute),
	).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "SET")
	}

	// Generate token
	jwt, err := accountAPI.AuthAPI.GenToken(ctx, &auth.Payload{
		ID:           accountID,
		ProjectID:    db.ProjectID,
		Names:        db.Names,
		EmailAddress: db.Email,
		PhoneNumber:  db.Phone,
		Group:        db.PrimaryGroup,
	}, time.Now().Add(10*time.Minute))
	if err != nil {
		return nil, err
	}

	// Outgoing context
	ctxExt := metadata.NewOutgoingContext(ctx, metadata.Pairs(auth.Header(), fmt.Sprintf("Bearer %s", jwt)))

	data := fmt.Sprintf(`Login OTP is %v \n\nExpires in 10 minutes`, uniqueNumber)

	if req.Project != "" {
		data = fmt.Sprintf(`Login OTP for %s. \n\nOTP is %d \nExpires in 10 minutes`, req.Project, uniqueNumber)
	}

	// // Send message
	// _, err = accountAPI.MessagingClient.SendMessage(ctxExt, &messaging.SendMessageRequest{
	// 	Message: &messaging.Message{
	// 		UserId:      accountID,
	// 		Title:       "OTP Login",
	// 		Data:        data,
	// 		Save:        true,
	// 		Type:        messaging.MessageType_INFO,
	// 		SendMethods: []messaging.SendMethod{messaging.SendMethod_SMSV2},
	// 	},
	// 	SmsAuth:         req.GetSmsAuth(),
	// 	SmsCredentialId: req.SmsCredentialId,
	// 	FetchSmsAuth:    req.FetchSmsAuth,
	// })
	// if err != nil {
	// 	return nil, errs.WrapErrorWithMsg(err, "failed to send otp to phone")
	// }

	// Send message
	_, err = accountAPI.MessagingClient.SendMessage(ctxExt, &messaging.SendMessageRequest{
		Message: &messaging.Message{
			UserId:      accountID,
			Title:       "OTP Login",
			Data:        data,
			Save:        true,
			Type:        messaging.MessageType_INFO,
			SendMethods: []messaging.SendMethod{messaging.SendMethod_SMSV2},
		},
		SmsAuth:         req.GetSmsAuth(),
		SmsCredentialId: req.SmsCredentialId,
		FetchSmsAuth:    req.FetchSmsAuth,
	})
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to send otp to phone")
	}

	return &emptypb.Empty{}, nil
}

const maxTrials = 4

var (
	blockedState = account.AccountState_BLOCKED.String()
)

func (accountAPI *accountAPIServer) SignInOTP(
	ctx context.Context, req *account.SignInOTPRequest,
) (*account.SignInResponse, error) {
	var err error

	// Validation
	switch {
	case req == nil:
		return nil, errs.NilObject("sign in otp request")
	case req.Username == "":
		return nil, errs.MissingField("username")
	case req.Otp == "":
		return nil, errs.MissingField("otp")
	}

	// Get the user from database
	db := &Account{}
	err = accountAPI.SQLDBWrites.
		First(db, "account_id = ?", req.Username).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.WrapMessage(codes.NotFound, "account does not exist")
	default:
		return nil, errs.FailedToFind("account", err)
	}

	if db.AccountState == blockedState {
		return nil, errs.WrapMessage(codes.PermissionDenied, "account is blocked")
	}

	trialsKey := fmt.Sprintf("accounts:otptrials:%d", db.AccountID)

	// Increment trials by 1
	trials, err := accountAPI.RedisDBWrites.Incr(ctx, trialsKey).Result()
	switch {
	case err == nil:
	case errors.Is(err, redis.Nil):
	default:
		return nil, errs.RedisCmdFailed(err, "ICR")
	}

	// Check if exceed trials
	if trials > maxTrials {
		// Block the account
		err = accountAPI.SQLDBWrites.Model(db).Update("account_state", account.AccountState_BLOCKED.String()).Error
		if err != nil {
			accountAPI.Logger.Errorln(err)
			return nil, errs.WrapMessage(codes.Internal, "failed to block account")
		}

		// Delete key
		err = accountAPI.RedisDBWrites.Del(ctx, trialsKey).Err()
		if err != nil {
			return nil, errs.RedisCmdFailed(err, "DEL")
		}

		return nil, errs.WrapMessage(codes.PermissionDenied, "account is blocked due to too many attempts.")
	}

	accountID := fmt.Sprint(db.AccountID)

	// Get otp
	otp, err := accountAPI.RedisDBWrites.Get(ctx, otpKey(accountID)).Result()
	switch {
	case err == nil:
	case errors.Is(err, redis.Nil):
		return nil, errs.WrapMessage(codes.DeadlineExceeded, "Login OTP expired, request another OTP")
	default:
		return nil, errs.RedisCmdFailed(err, "GET")
	}

	// Compare otp
	if otp != req.Otp {
		return nil, errs.WrapMessage(codes.Unauthenticated, "Login OTP do not match")
	}

	// Delete key
	err = accountAPI.RedisDBWrites.Del(ctx, trialsKey).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "DEL")
	}

	return accountAPI.updateSession(ctx, db, req.Group)
}

func (accountAPI *accountAPIServer) RequestActivateAccountOTP(
	ctx context.Context, req *account.RequestActivateAccountOTPRequest,
) (*empty.Empty, error) {
	var err error

	// Validation
	switch {
	case req == nil:
		return nil, errs.NilObject("RequestChangePrivateAccountRequest")
	case req.AccountId == "":
		return nil, errs.MissingField("account id")
		// case req.SmsAuth == nil:
		// 	return nil, errs.MissingField("sms auth")
	}

	// Retrieve token claims
	_, err = accountAPI.AuthAPI.AuthorizeActorOrGroup(ctx, req.AccountId, accountAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.Unauthenticated, err, "failed to authorize request")
	}

	// GetAccount the user from database
	db := &Account{}
	err = accountAPI.SQLDBWrites.
		First(db, "account_id = ?", req.AccountId).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.WrapMessagef(codes.NotFound, "account with id %s does not exist", req.AccountId)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	accountID := fmt.Sprint(db.AccountID)

	uniqueNumber := randomdata.Number(100000, 999999)

	// Set token with expiration of 5 minutes
	err = accountAPI.RedisDBWrites.Set(
		ctx, otpKey(accountID), uniqueNumber, time.Duration(5*time.Minute),
	).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "SET")
	}

	// Generate token
	jwt, err := accountAPI.AuthAPI.GenToken(ctx, &auth.Payload{
		ID:           accountID,
		ProjectID:    db.ProjectID,
		Names:        db.Names,
		EmailAddress: db.Email,
		PhoneNumber:  db.Phone,
		Group:        db.PrimaryGroup,
	}, time.Now().Add(10*time.Minute))
	if err != nil {
		return nil, err
	}

	// Outgoing context
	ctxExt := metadata.NewOutgoingContext(ctx, metadata.Pairs(auth.Header(), fmt.Sprintf("Bearer %s", jwt)))

	data := fmt.Sprintf(`Account verification OTP for %s. \n\nOTP is %d \nExpires in 10 minutes`, db.ProjectID, uniqueNumber)

	// Send message
	_, err = accountAPI.MessagingClient.SendMessage(ctxExt, &messaging.SendMessageRequest{
		Message: &messaging.Message{
			UserId:      accountID,
			Title:       "Account Verification OTP",
			Data:        data,
			Save:        true,
			Type:        messaging.MessageType_INFO,
			SendMethods: []messaging.SendMethod{messaging.SendMethod_SMSV2},
		},
		SmsAuth:         req.GetSmsAuth(),
		SmsCredentialId: req.SmsCredentialId,
		FetchSmsAuth:    req.FetchSmsAuth,
	})
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to send otp to phone")
	}

	return &emptypb.Empty{}, nil
}

func (accountAPI *accountAPIServer) ActivateAccountOTP(
	ctx context.Context, req *account.ActivateAccountOTPRequest,
) (*account.ActivateAccountResponse, error) {
	var err error
	// Validation
	switch {
	case req == nil:
		return nil, errs.NilObject("RequestChangePrivateAccountRequest")
	case req.AccountId == "":
		return nil, errs.MissingField("account id")
	case req.Otp == "":
		return nil, errs.MissingField("OTP code")
	}

	// Authorization
	_, err = accountAPI.AuthAPI.AuthorizeActorOrGroup(ctx, req.AccountId, accountAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.Unauthenticated, err, "failed to authorize request")
	}

	// Get the user from database
	db := &Account{}
	err = accountAPI.SQLDBWrites.
		First(db, "account_id = ?", req.AccountId).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.WrapMessage(codes.NotFound, "account does not exist")
	default:
		return nil, errs.FailedToFind("account", err)
	}

	if db.AccountState == blockedState {
		return nil, errs.WrapMessage(codes.PermissionDenied, "account is blocked")
	}

	// Get otp
	otp, err := accountAPI.RedisDBWrites.Get(ctx, otpKey(req.AccountId)).Result()
	switch {
	case err == nil:
	case errors.Is(err, redis.Nil):
		return nil, errs.WrapMessage(codes.DeadlineExceeded, "Verification OTP expired, request another OTP")
	default:
		return nil, errs.RedisCmdFailed(err, "GET")
	}

	// Compare otp
	if otp != req.Otp {
		return nil, errs.WrapMessage(codes.Unauthenticated, "Verification OTP do not match")
	}

	// Update the model of the user to activate their account
	err = accountAPI.SQLDBWrites.Table(accountsTable).Where("account_id=?", db.AccountID).
		Update("account_state", account.AccountState_ACTIVE.String()).Error
	if err != nil {
		return nil, errs.FailedToUpdate("account", err)
	}

	return &account.ActivateAccountResponse{}, nil
}
