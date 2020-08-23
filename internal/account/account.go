package account

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/gorilla/securecookie"
	"gorm.io/gorm"

	"google.golang.org/grpc/grpclog"

	"github.com/gidyon/services/internal/pkg/fauth"
	"github.com/gidyon/services/pkg/api/messaging"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/dbutil"
	"github.com/gidyon/services/pkg/utils/errs"
	"github.com/gidyon/services/pkg/utils/mdutil"
	"github.com/gidyon/services/pkg/utils/templateutil"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/speps/go-hashids"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const templateName = "base"

type cookier interface {
	Decode(string, string, interface{}) error
	Encode(string, interface{}) (string, error)
}

type accountAPIServer struct {
	activationURL      string
	appName            string
	sqlDB              *gorm.DB
	redisDB            *redis.Client
	logger             grpclog.LoggerV2
	authAPI            auth.Interface
	messagingClient    messaging.MessagingClient
	firebaseAuthClient fauth.FirebaseAuthClient
	tpl                *template.Template
	hasher             *hashids.HashID
	cookier            cookier
	setCookie          func(context.Context, string) error
}

// Options contain parameters for NewAccountAPI
type Options struct {
	AppName         string
	TemplatesDir    string
	ActivationURL   string
	JWTSigningKey   []byte
	SQLDB           *gorm.DB
	RedisDB         *redis.Client
	SecureCookie    *securecookie.SecureCookie
	Logger          grpclog.LoggerV2
	MessagingClient messaging.MessagingClient
	FirebaseAuth    fauth.FirebaseAuthClient
}

func newHasher(salt string) (*hashids.HashID, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 30

	return hashids.NewWithData(hd)
}

// NewAccountAPI creates an account API singleton
func NewAccountAPI(ctx context.Context, opt *Options) (account.AccountAPIServer, error) {
	// Validation
	var err error
	switch {
	case ctx == nil:
		err = errs.NilObject("context")
	case opt == nil:
		err = errs.NilObject("options")
	case opt.AppName == "":
		err = errs.MissingField("app name")
	case opt.TemplatesDir == "":
		err = errs.MissingField("templates directory")
	case opt.ActivationURL == "":
		err = errs.MissingField("activation url")
	case opt.JWTSigningKey == nil:
		err = errs.MissingField("jwt token")
	case opt.SQLDB == nil:
		err = errs.NilObject("sql db")
	case opt.RedisDB == nil:
		err = errs.NilObject("redis dB")
	case opt.SecureCookie == nil:
		err = errs.NilObject("secure cookie")
	case opt.Logger == nil:
		err = errs.NilObject("logger")
	case opt.MessagingClient == nil:
		err = errs.NilObject("messaging client")
	case opt.FirebaseAuth == nil:
		err = errs.NilObject("firebase auth")
	}
	if err != nil {
		return nil, err
	}

	// Auth API
	authAPI, err := auth.NewAPI(opt.JWTSigningKey, "Accounts Service", "users")
	if err != nil {
		return nil, err
	}

	// Pagination hasher
	hasher, err := newHasher(string(opt.JWTSigningKey))
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to generate hash id")
	}

	// Account API
	accountAPI := &accountAPIServer{
		activationURL:      opt.ActivationURL,
		appName:            opt.AppName,
		sqlDB:              opt.SQLDB,
		redisDB:            opt.RedisDB,
		logger:             opt.Logger,
		authAPI:            authAPI,
		messagingClient:    opt.MessagingClient,
		firebaseAuthClient: opt.FirebaseAuth,
		hasher:             hasher,
		cookier:            opt.SecureCookie,
		setCookie: func(ctx context.Context, cookie string) error {
			err := grpc.SetHeader(ctx, metadata.Pairs("set-cookie", cookie))
			if err != nil {
				return errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to set cookie")
			}
			err = grpc.SetHeader(ctx, metadata.Pairs("access-control-expose-headers", "set-cookie"))
			if err != nil {
				return errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to expose set-cookie header")
			}
			return nil
		},
	}

	// Read template files from directory
	tFiles, err := templateutil.ReadFiles(opt.TemplatesDir)
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to read template files in directory")
	}

	// Parse template
	accountAPI.tpl, err = templateutil.ParseTemplate(tFiles...)
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to parse template")
	}

	// Perform auto migration
	err = accountAPI.sqlDB.AutoMigrate(&Account{})
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to automigrate accounts table")
	}

	// Create a full text search index
	err = dbutil.CreateFullTextIndex(accountAPI.sqlDB, accountsTable, "email", "phone")
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to create full text index")
	}

	return accountAPI, nil
}

func (accountAPI *accountAPIServer) SignInExternal(
	ctx context.Context, signInReq *account.SignInExternalRequest,
) (*account.SignInResponse, error) {
	// Request must not be nil
	if signInReq == nil {
		return nil, errs.NilObject("SignInExternalRequest")
	}

	// Validation
	var err error
	switch {
	case signInReq.Account == nil:
		err = errs.NilObject("account")
	case signInReq.Account.Names == "":
		err = errs.MissingField("names")
	case signInReq.Account.Email == "" && signInReq.Account.Phone == "":
		err = errs.MissingField("email and phone")
	case signInReq.AuthToken == "":
		err = errs.MissingField("auth token")
	}
	if err != nil {
		return nil, err
	}

	// Verify ID token
	_, err = accountAPI.firebaseAuthClient.VerifyIDToken(ctx, signInReq.AuthToken)
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to verify firebase ID token")
	}

	accountDB := &Account{}

	shouldUpdateUser := true

	var ID uint

	// Get user
	switch {
	case signInReq.Account.Email != "":
		err = accountAPI.sqlDB.First(accountDB, "email=?", signInReq.Account.Email).Error
	case signInReq.Account.Phone != "":
		err = accountAPI.sqlDB.First(accountDB, "phone=?", signInReq.Account.Phone).Error
	}
	switch {
	case err == nil:
		ID = accountDB.ID
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Create user
		accountDB, err = GetAccountDB(signInReq.Account)
		if err != nil {
			return nil, err
		}
		accountDB.AccountState = account.AccountState_ACTIVE.String()
		err = accountAPI.sqlDB.Create(accountDB).Error
		if err != nil {
			return nil, errs.FailedToSave("account", err)
		}
		shouldUpdateUser = false
		ID = accountDB.ID
	default:
		return nil, errs.FailedToSave("account", err)
	}

	if shouldUpdateUser {
		err = accountAPI.sqlDB.Table(accountsTable).Where("id = ?", ID).
			Updates(accountDB).Error
		if err != nil {
			return nil, errs.FailedToUpdate("account", err)
		}
	}

	return accountAPI.updateSession(ctx, accountDB, "")
}

func (accountAPI *accountAPIServer) RefreshSession(
	ctx context.Context, req *account.RefreshSessionRequest,
) (*account.SignInResponse, error) {
	// Request must not be nil
	if req == nil {
		return nil, errs.NilObject("RefreshSessionRequest")
	}

	// Validation
	var ID int
	var err error
	switch {
	case req.RefreshToken == "":
		return nil, errs.NilObject("refresh token")
	case req.AccountId == "":
		return nil, errs.NilObject("account id")
	default:
		ID, err = strconv.Atoi(req.AccountId)
		if err != nil {
			return nil, errs.IncorrectVal("account id")
		}
	}

	// Ensure that refresh token already exists
	ok, err := accountAPI.redisDB.SIsMember(refreshTokenSet(), req.RefreshToken).Result()
	switch {
	case err == nil:
		if !ok {
			return nil, errs.WrapMessage(codes.Unauthenticated, "not signed in")
		}
	case errors.Is(err, redis.Nil) || !ok:
		return nil, errs.WrapMessage(codes.Unauthenticated, "not signed in")
	default:
		return nil, errs.RedisCmdFailed(err, "get")
	}

	accountDB := &Account{}
	err = accountAPI.sqlDB.First(accountDB, "id=?", ID).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.DoesNotExist("account", req.AccountId)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	switch {
	case accountDB.AccountState == account.AccountState_BLOCKED.String():
		return nil, errs.WrapMessage(codes.PermissionDenied, "account is blocked")
	case accountDB.AccountState == account.AccountState_DELETED.String():
		return nil, errs.WrapMessage(codes.PermissionDenied, "account is deleted")
	}

	return accountAPI.updateSession(ctx, accountDB, req.AccountGroup)
}

func (accountAPI *accountAPIServer) ActivateAccount(
	ctx context.Context, activateReq *account.ActivateAccountRequest,
) (*account.ActivateAccountResponse, error) {
	// Request must not be nil
	if activateReq == nil {
		return nil, errs.NilObject("ActivateAccountRequest")
	}

	// Validation 1
	var ID int
	var err error
	switch {
	case activateReq.Token == "":
		return nil, errs.MissingField("token")
	case activateReq.AccountId == "":
		return nil, errs.MissingField("account id")
	default:
		ID, err = strconv.Atoi(activateReq.AccountId)
		if err != nil {
			return nil, errs.IncorrectVal("account id")
		}
	}

	// Retrieve token claims
	payload, err := accountAPI.authAPI.AuthorizeActorOrGroups(
		auth.AddTokenMD(ctx, activateReq.Token), activateReq.Token, auth.AdminGroup(),
	)
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.Unauthenticated, err, "failed to authorize request")
	}

	// Validation 2
	dev := (strings.ToLower(os.Getenv("MODE")) == "development")
	switch {
	case payload.ID == "":
		if !dev {
			return nil, errs.MissingField("token id")
		}
	}

	// Compare if account id matches or if activated by admin
	isOwner := payload.ID == activateReq.AccountId
	isAdmin := payload.Group == auth.AdminGroup()
	if !isOwner && !isAdmin {
		if !dev {
			switch {
			case !isAdmin:
				return nil, errs.WrapMessage(codes.PermissionDenied, "not admin user")
			case !isOwner:
				return nil, errs.TokenCredentialNotMatching("account id")
			}
		}
	}

	// Check that account exists
	if errors.Is(accountAPI.sqlDB.Select("account_state").
		First(&Account{}, "id=?", ID).Error, gorm.ErrRecordNotFound) {
		return nil, errs.DoesNotExist("account", activateReq.AccountId)
	}

	// Update the model of the user to activate their account
	err = accountAPI.sqlDB.Table(accountsTable).Where("id=?", ID).
		Update("account_state", account.AccountState_ACTIVE.String()).Error
	if err != nil {
		return nil, errs.FailedToUpdate("account", err)
	}

	return &account.ActivateAccountResponse{}, nil
}

func (accountAPI *accountAPIServer) UpdateAccount(
	ctx context.Context, updateReq *account.UpdateAccountRequest,
) (*empty.Empty, error) {
	// Request must not be nil
	if updateReq == nil {
		return nil, errs.NilObject("UpdateRequest")
	}

	// Authorization
	_, err := accountAPI.authAPI.AuthorizeActorOrGroups(ctx, updateReq.GetAccount().GetAccountId(), auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case updateReq.Account == nil:
		return nil, errs.NilObject("Account")
	case updateReq.Account.AccountId == "":
		return nil, errs.MissingField("AccountID")
	default:
		_, err = strconv.Atoi(updateReq.Account.AccountId)
		if err != nil {
			return nil, errs.WrapMessage(codes.InvalidArgument, "account id is incorrect")
		}
	}

	// GetAccount the account details from database
	accountDB := &Account{}
	err = accountAPI.sqlDB.Select("account_state").
		First(accountDB, "id=?", updateReq.Account.AccountId).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.DoesNotExist("account", updateReq.Account.AccountId)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	// Check that account is not blocked
	if accountDB.AccountState == account.AccountState_BLOCKED.String() {
		return nil, errs.WrapMessage(codes.PermissionDenied, "account is blocked")
	}

	accountDBX, err := GetAccountDB(updateReq.Account)
	if err != nil {
		return nil, err
	}

	// Update the model; omit "id", "primary_group", "account_state"
	err = accountAPI.sqlDB.Model(accountDBX).
		Omit("id", "primary_group", "account_state", "password", "security_answer", "security_question").
		Where("id=?", updateReq.Account.AccountId).
		Updates(accountDBX).Error
	switch {
	case err == nil:
	default:
		emailOrPhone := func(err error) string {
			if strings.Contains(strings.ToLower(err.Error()), "email") {
				return "email " + accountDBX.Email
			}
			if strings.Contains(strings.ToLower(err.Error()), "phone") {
				return "phone " + accountDBX.Phone
			}
			return "id " + fmt.Sprint(accountDBX.ID)
		}
		// Check if duplicate
		if dbutil.IsDuplicate(err) {
			return nil, errs.DoesExist("account", emailOrPhone(err))
		}
	}

	return &empty.Empty{}, nil
}

func updateToken(accountID string) string {
	return "updatetoken:" + accountID
}

func (accountAPI *accountAPIServer) RequestChangePrivateAccount(
	ctx context.Context, req *account.RequestChangePrivateAccountRequest,
) (*account.RequestChangePrivateAccountResponse, error) {
	// Request must not be nil
	if req == nil {
		return nil, errs.NilObject("RequestChangePrivateAccountRequest")
	}

	// Authentication
	err := accountAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case req.Payload == "":
		return nil, errs.MissingField("payload")
	case req.FallbackUrl == "":
		return nil, errs.MissingField("fallback url")
	case req.SendMethod == messaging.SendMethod_SEND_METHOD_UNSPECIFIED:
		return nil, errs.WrapMessage(codes.InvalidArgument, "send method is unspecified")
	}

	// GetAccount the user from database
	accountDB := &Account{}
	err = accountAPI.sqlDB.
		Find(accountDB, "email=? OR phone=?", req.Payload, req.Payload).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		emailOrPhone := func(err error) string {
			if strings.Contains(strings.ToLower(err.Error()), "email") {
				return "email " + req.Payload
			}
			if strings.Contains(strings.ToLower(err.Error()), "phone") {
				return "phone " + req.Payload
			}
			return "email or phone " + req.Payload
		}
		return nil, errs.WrapMessagef(codes.NotFound, "account with %s does not exist", emailOrPhone(err))
	default:
		return nil, errs.FailedToFind("account", err)
	}

	accountID := fmt.Sprint(accountDB.ID)

	// Authorize the actor
	_, err = accountAPI.authAPI.AuthorizeActorOrGroups(ctx, accountID, auth.AdminGroup())
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to authorize actor")
	}

	uniqueNumber := rand.Intn(499999) + 500000

	// Set token with expiration of 6 hours
	err = accountAPI.redisDB.Set(updateToken(accountID), uniqueNumber, time.Duration(time.Hour*6)).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "SET")
	}

	// GetAccount jwt
	jwtToken, err := accountAPI.authAPI.GenToken(ctx, &auth.Payload{
		ID:    accountID,
		Names: accountDB.Names,
	}, time.Now().Add(6*time.Hour))
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to generate token")
	}

	link := fmt.Sprintf("%s?token=%s&id=%s&passphrase=%d", req.FallbackUrl, jwtToken, accountID, uniqueNumber)

	// Send message
	_, err = accountAPI.messagingClient.SendMessage(mdutil.AddFromCtx(ctx), &messaging.Message{
		UserId: accountID,
		Title:  "Reset Account Credentials",
		Data: fmt.Sprintf(
			"You requested to change your account security credentials. Reset token is %d.", uniqueNumber,
		),
		Link:        link,
		Save:        true,
		Type:        messaging.MessageType_ALERT,
		SendMethods: []messaging.SendMethod{req.SendMethod},
		Details: map[string]string{
			"origin": "Accounts API",
			"app":    accountAPI.appName,
		},
	})
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to send message")
	}

	return &account.RequestChangePrivateAccountResponse{
		ResponseMessage: "reset token was sent to " + req.Payload,
	}, nil
}

func (accountAPI *accountAPIServer) UpdatePrivateAccount(
	ctx context.Context, updatePrivateReq *account.UpdatePrivateAccountRequest,
) (*empty.Empty, error) {
	// Request must not be nil
	if updatePrivateReq == nil {
		return nil, errs.NilObject("UpdatePrivateRequest")
	}

	// Authorization
	_, err := accountAPI.authAPI.AuthorizeActorOrGroups(ctx, updatePrivateReq.AccountId, auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case updatePrivateReq.AccountId == "":
		return nil, errs.MissingField("account id")
	case updatePrivateReq.PrivateAccount == nil:
		return nil, errs.NilObject("private account")
	case updatePrivateReq.ChangeToken == "":
		return nil, errs.MissingField("change token")
	default:
		ID, err = strconv.Atoi(updatePrivateReq.AccountId)
		if err != nil {
			return nil, errs.IncorrectVal("account id")
		}
	}

	// GetAccount the account details from database
	accountDB := &Account{}
	err = accountAPI.sqlDB.Select("account_state").First(accountDB, "id=?", ID).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.DoesNotExist("account", updatePrivateReq.AccountId)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	// Check that account is not blocked
	if accountDB.AccountState == account.AccountState_BLOCKED.String() {
		return nil, errs.WrapMessage(codes.PermissionDenied, "account not active")
	}

	// Hash the password if not empty
	if updatePrivateReq.PrivateAccount.Password != "" {
		// Lets get the update token
		token, err := accountAPI.redisDB.Get(updateToken(updatePrivateReq.AccountId)).Result()
		switch {
		case err == nil:
		case err == redis.Nil:
			return nil, errs.WrapMessage(codes.NotFound, "update token not found")
		default:
			return nil, errs.RedisCmdFailed(err, "get token")
		}

		if token != updatePrivateReq.ChangeToken {
			return nil, errs.WrapMessage(codes.InvalidArgument, "token is incorrect")
		}

		// Passwords must be similar
		if updatePrivateReq.PrivateAccount.ConfirmPassword != updatePrivateReq.PrivateAccount.Password {
			return nil, errs.WrapMessage(codes.InvalidArgument, "passwords do not match")
		}

		updatePrivateReq.PrivateAccount.Password, err = genHash(updatePrivateReq.PrivateAccount.Password)
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to generate password hash")
		}
	}

	// Create database model of the new account
	privateDB := &Account{
		SecurityQuestion: updatePrivateReq.PrivateAccount.SecurityQuestion,
		SecurityAnswer:   updatePrivateReq.PrivateAccount.SecurityAnswer,
		Password:         updatePrivateReq.PrivateAccount.Password,
	}

	// Update the model
	err = accountAPI.sqlDB.Model(privateDB).Where("id=?", ID).Updates(privateDB).Error
	if err != nil {
		return nil, errs.FailedToUpdate("account", err)
	}

	return &empty.Empty{}, nil
}

func (accountAPI *accountAPIServer) DeleteAccount(
	ctx context.Context, delReq *account.DeleteAccountRequest,
) (*empty.Empty, error) {
	// Request must not be nil
	if delReq == nil {
		return nil, errs.NilObject("DeleteAccountRequest")
	}

	// Authorization
	_, err := accountAPI.authAPI.AuthorizeActorOrGroups(ctx, delReq.AccountId, auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	// Check account id is provided
	var ID int
	switch {
	case delReq.AccountId == "":
		return nil, errs.MissingField("AccountID")
	default:
		ID, err = strconv.Atoi(delReq.AccountId)
		if err != nil {
			return nil, errs.WrapMessage(codes.InvalidArgument, "incorrect account id")
		}
	}

	// Get the account details from database
	accountDB := &Account{}
	err = accountAPI.sqlDB.Select("account_state").First(accountDB, "id=?", ID).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.DoesNotExist("account", delReq.AccountId)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	// Check that account is not blocked
	if accountDB.AccountState == account.AccountState_BLOCKED.String() {
		return nil, errs.WrapMessage(codes.PermissionDenied, "account is blocked")
	}

	// Soft delete their account
	err = accountAPI.sqlDB.Delete(accountDB, "id=?", ID).Error
	if err != nil {
		return nil, errs.FailedToDelete("account", err)
	}

	return &empty.Empty{}, nil
}

func (accountAPI *accountAPIServer) GetAccount(
	ctx context.Context, getReq *account.GetAccountRequest,
) (*account.Account, error) {
	// Request must not be nil
	if getReq == nil {
		return nil, errs.NilObject("GetAccountRequest")
	}

	// Authorization
	payload, err := accountAPI.authAPI.AuthorizeActorOrGroups(ctx, getReq.AccountId, auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	// Validation
	var ID int
	switch {
	case getReq.AccountId == "":
		return nil, errs.MissingField("account id")
	default:
		ID, err = strconv.Atoi(getReq.AccountId)
		if err != nil {
			return nil, errs.WrapMessage(codes.InvalidArgument, "incorrect accoint id")
		}
	}

	// GetAccount account from database
	accountDB := &Account{}

	if getReq.Priviledge {
		if payload.Group == auth.AdminGroup() {
			err = accountAPI.sqlDB.Unscoped().First(accountDB, "id=?", ID).Error
		} else {
			err = accountAPI.sqlDB.First(accountDB, "id=?", ID).Error
		}
	} else {
		err = accountAPI.sqlDB.First(accountDB, "id=?", ID).Error
	}
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.DoesNotExist("account", getReq.AccountId)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	// Account should not be blocked
	if accountDB.AccountState == account.AccountState_BLOCKED.String() && !getReq.Priviledge {
		return nil, errs.WrapMessage(codes.PermissionDenied, "account is blocked")
	}

	accountPB, err := GetAccountPB(accountDB)
	if err != nil {
		return nil, err
	}

	return GetAccountPBView(accountPB, getReq.GetView()), nil
}

func (accountAPI *accountAPIServer) BatchGetAccounts(
	ctx context.Context, batchReq *account.BatchGetAccountsRequest,
) (*account.BatchGetAccountsResponse, error) {
	return nil, nil
}

func (accountAPI *accountAPIServer) GetLinkedAccounts(
	ctx context.Context, getReq *account.GetLinkedAccountsRequest,
) (*account.GetLinkedAccountsResponse, error) {
	return nil, nil
}

func (accountAPI *accountAPIServer) ExistAccount(
	ctx context.Context, existReq *account.ExistAccountRequest,
) (*account.ExistAccountResponse, error) {
	// Request must not be nil
	if existReq == nil {
		return nil, errs.NilObject("ExistAccountRequest")
	}

	// Authenticate the request
	err := accountAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	var (
		externalID = existReq.GetExternalId()
		email      = existReq.GetEmail()
		phone      = existReq.GetPhone()
	)

	// Validation
	if email == "" && phone == "" && externalID == "" {
		return nil, errs.MissingField("email, phone or external id")
	}

	accountDB := &Account{}

	// Query for account with email or phone
	err = accountAPI.sqlDB.Select("email,phone,external_id").
		First(accountDB, "email=? OR phone=? OR external_id=?", email, phone, externalID).Error
	switch {
	case err == nil:
		// Account exist
		return &account.ExistAccountResponse{
			Exists: true,
		}, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Account dosn't exist
		return &account.ExistAccountResponse{
			Exists: false,
		}, nil
	default:
		return nil, errs.FailedToFind("account", err)
	}
}

const defaultPageSize = 20

func (accountAPI *accountAPIServer) ListAccounts(
	ctx context.Context, listReq *account.ListAccountsRequest,
) (*account.Accounts, error) {
	// Request must not be nil
	if listReq == nil {
		return nil, errs.NilObject("ListRequest")
	}

	// Authenticate the request
	payload, err := accountAPI.authAPI.AuthenticateRequestV2(ctx)
	if err != nil {
		return nil, err
	}

	// Parse page size and page token
	pageSize := listReq.GetPageSize()
	if pageSize <= 0 || pageSize > defaultPageSize {
		pageSize = defaultPageSize
	}

	var id uint

	// Get last id from page token
	pageToken := listReq.GetPageToken()
	if pageToken != "" {
		ids, err := accountAPI.hasher.DecodeInt64WithError(listReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		id = uint(ids[0])
	}

	accountsDB := make([]*Account, 0, pageSize)

	// Apply filter criterias
	db := generateWhereCondition(accountAPI.sqlDB, listReq.GetListCriteria())

	if payload.Group == auth.AdminGroup() {
		db = db.Unscoped()
	}

	// Order by ID
	db = db.Limit(int(pageSize)).Order("id DESC")

	err = db.Find(&accountsDB).Error
	switch {
	case err == nil:
	default:
		return nil, errs.FailedToFind("accounts", err)
	}

	accountsPB := make([]*account.Account, 0, len(accountsDB))

	for _, accountDB := range accountsDB {
		accountPB, err := GetAccountPB(accountDB)
		if err != nil {
			return nil, err
		}
		accountsPB = append(accountsPB, GetAccountPBView(accountPB, listReq.GetView()))
		id = accountDB.ID
	}

	var token string
	if int(pageSize) == len(accountsDB) {
		// Next page token
		token, err = accountAPI.hasher.EncodeInt64([]int64{int64(id)})
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to generate page token")
		}
	}

	return &account.Accounts{
		NextPageToken: token,
		Accounts:      accountsPB,
	}, nil
}

// Searches for accounts
func (accountAPI *accountAPIServer) SearchAccounts(
	ctx context.Context, searchReq *account.SearchAccountsRequest,
) (*account.Accounts, error) {
	// Request must not be nil
	if searchReq == nil {
		return nil, errs.NilObject("SearchRequest")
	}

	// Authenticate the request
	payload, err := accountAPI.authAPI.AuthenticateRequestV2(ctx)
	if err != nil {
		return nil, err
	}

	// For empty queries
	if searchReq.Query == "" {
		return &account.Accounts{
			Accounts: []*account.Account{},
		}, nil
	}

	// Parse page size and page token
	pageSize := searchReq.GetPageSize()
	if pageSize <= 0 || pageSize > defaultPageSize {
		pageSize = defaultPageSize
	}

	var ID uint

	// Get last id from page token
	pageToken := searchReq.GetPageToken()
	if pageToken != "" {
		ids, err := accountAPI.hasher.DecodeInt64WithError(searchReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		ID = uint(ids[0])
	}

	accountsDB := make([]*Account, 0, pageSize)

	// Apply filter criterias
	db := generateWhereCondition(accountAPI.sqlDB, searchReq.GetSearchCriteria())

	if payload.Group == auth.AdminGroup() {
		db = db.Unscoped()
	}

	// Order by ID
	db = db.Limit(int(pageSize)).Order("id DESC")

	parsedQuery := dbutil.ParseQuery(searchReq.Query)

	if searchReq.SearchLinkedAccounts {
		err = db.Find(&accountsDB, "MATCH(email, phone, linked_accounts) AGAINST(? IN BOOLEAN MODE)", parsedQuery).
			Error
	} else {
		err = db.Find(&accountsDB, "MATCH(email, phone) AGAINST(? IN BOOLEAN MODE)", parsedQuery).
			Error
	}
	switch {
	case err == nil:
	default:
		return nil, errs.FailedToFind("accounts", err)
	}

	accountsPB := make([]*account.Account, 0, len(accountsDB))

	for _, accountDB := range accountsDB {
		accountPB, err := GetAccountPB(accountDB)
		if err != nil {
			return nil, err
		}
		accountsPB = append(accountsPB, GetAccountPBView(accountPB, searchReq.GetView()))
		ID = accountDB.ID
	}

	var token string
	if int(pageSize) == len(accountsDB) {
		// Next page token
		token, err = accountAPI.hasher.EncodeInt64([]int64{int64(ID)})
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to generate page token")
		}
	}

	return &account.Accounts{
		NextPageToken: token,
		Accounts:      accountsPB,
	}, nil
}

func generateWhereCondition(db *gorm.DB, criteria *account.Criteria) *gorm.DB {
	if criteria == nil || !criteria.Filter {
		return db
	}

	// Filter by account state
	switch {
	case criteria.ShowActiveAccounts:
		db = db.Where("account_state = ?", account.AccountState_ACTIVE.String())
	case criteria.ShowInactiveAccounts:
		db = db.Where("account_state = ?", account.AccountState_INACTIVE.String())
	case criteria.ShowBlockedAccounts:
		db = db.Where("account_state = ?", account.AccountState_BLOCKED.String())
	}

	// Filter by gender
	switch {
	case criteria.ShowFemales:
		db = db.Where("gender = ?", "female")
	case criteria.ShowMales:
		db = db.Where("gender = ?", "male")
	}

	// Filter by date
	if criteria.FilterCreationDate {
		nowSecs := time.Now().Unix()
		switch {
		case criteria.CreatedFrom > 0 && criteria.CreatedUntil > 0 && criteria.CreatedFrom < criteria.CreatedUntil:
			db = db.Where(
				"UNIX_TIMESTAMP(created_at) BETWEEN ? AND ?",
				criteria.CreatedFrom, criteria.CreatedUntil,
			)
		case criteria.CreatedUntil > 0:
			db = db.Where(
				"UNIX_TIMESTAMP(created_at) < ?", criteria.CreatedUntil,
			)
		case criteria.CreatedFrom > 0:
			if criteria.CreatedFrom < nowSecs {
				db = db.Where(
					"UNIX_TIMESTAMP(created_at) > ?", criteria.CreatedFrom,
				)
			}
		}
	}

	// Filter by primary_groups
	if criteria.FilterAccountGroups {
		db = db.Where("primary_group IN (?)", criteria.GetGroups())
	}

	return db
}
