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
	"gorm.io/hints"

	"google.golang.org/grpc/grpclog"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/dbutil"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/micro/v2/utils/mdutil"
	"github.com/gidyon/micro/v2/utils/templateutil"
	"github.com/gidyon/services/internal/pkg/fauth"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/messaging"

	redis "github.com/go-redis/redis/v8"
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
	account.UnimplementedAccountAPIServer
	activationURL string
	tpl           *template.Template
	cookier       cookier
	setCookie     func(context.Context, string) error
	*Options
}

// Options contain parameters for NewAccountAPI
type Options struct {
	AppName            string
	EmailDisplayName   string
	DefaultEmailSender string
	TemplatesDir       string
	ActivationURL      string
	PaginationHasher   *hashids.HashID
	AuthAPI            auth.API
	SQLDBWrites        *gorm.DB
	SQLDBReads         *gorm.DB
	RedisDBWrites      *redis.Client
	RedisDBReads       *redis.Client
	SecureCookie       *securecookie.SecureCookie
	Logger             grpclog.LoggerV2
	MessagingClient    messaging.MessagingClient
	FirebaseAuth       fauth.FirebaseAuthClient
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
	case opt.EmailDisplayName == "":
		err = errs.MissingField("email display name")
	case opt.TemplatesDir == "":
		err = errs.MissingField("templates directory")
	case opt.ActivationURL == "":
		err = errs.MissingField("activation url")
	case opt.PaginationHasher == nil:
		err = errs.MissingField("pagination PaginationHasher")
	case opt.AuthAPI == nil:
		err = errs.MissingField("authentication API")
	case opt.SQLDBWrites == nil:
		err = errs.NilObject("sql writes db")
	case opt.SQLDBReads == nil:
		err = errs.NilObject("sql reads db")
	case opt.RedisDBWrites == nil:
		err = errs.NilObject("redis writes db")
	case opt.RedisDBReads == nil:
		err = errs.NilObject("redis reads db")
	case opt.SecureCookie == nil:
		err = errs.NilObject("secure cookie")
	case opt.Logger == nil:
		err = errs.NilObject("Logger")
	case opt.MessagingClient == nil:
		err = errs.NilObject("messaging client")
	case opt.FirebaseAuth == nil:
		err = errs.NilObject("firebase auth")
	}
	if err != nil {
		return nil, err
	}

	// Account API
	accountAPI := &accountAPIServer{
		activationURL: opt.ActivationURL,
		cookier:       opt.SecureCookie,
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
		Options: opt,
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
	if !accountAPI.SQLDBWrites.Migrator().HasTable(accountsTable) {
		err = accountAPI.SQLDBWrites.AutoMigrate(&Account{})
		if err != nil {
			return nil, errs.WrapErrorWithMsg(err, "failed to automigrate accounts table")
		}
	}

	if !accountAPI.SQLDBWrites.Migrator().HasIndex(&Account{}, dbutil.FullTextIndex) {
		// Create a full text search index
		err = dbutil.CreateFullTextIndex(accountAPI.SQLDBWrites, accountsTable, "names", "email", "phone", "linked_accounts")
		if err != nil {
			return nil, errs.WrapErrorWithMsg(err, "failed to create full text index")
		}
	}

	return accountAPI, nil
}

func (accountAPI *accountAPIServer) SignInExternal(
	ctx context.Context, signInReq *account.SignInExternalRequest,
) (*account.SignInResponse, error) {
	// Validation
	var err error
	switch {
	case signInReq == nil:
		err = errs.NilObject("sign in request")
	case signInReq.ProjectId == "":
		err = errs.MissingField("project id")
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
	_, err = accountAPI.FirebaseAuth.VerifyIDToken(ctx, signInReq.AuthToken)
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to verify firebase ID token")
	}

	var (
		accountDB = &Account{}
		ID        uint
	)

	// Get user
	switch {
	case signInReq.Account.Email != "":
		err = accountAPI.SQLDBWrites.First(accountDB, "email=?", signInReq.Account.Email).Error
	case signInReq.Account.Phone != "":
		err = accountAPI.SQLDBWrites.First(accountDB, "phone=?", signInReq.Account.Phone).Error
	}
	switch {
	case err == nil:
		ID = accountDB.AccountID
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Create user
		signInReq.Account.ProjectId = signInReq.ProjectId
		accountDB, err = GetAccountDB(signInReq.Account)
		if err != nil {
			return nil, err
		}
		accountDB.AccountState = account.AccountState_ACTIVE.String()
		err = accountAPI.SQLDBWrites.Create(accountDB).Error
		if err != nil {
			return nil, errs.FailedToSave("account", err)
		}
		return accountAPI.updateSession(ctx, accountDB, "")
	default:
		return nil, errs.FailedToSave("account", err)
	}

	// Update account
	err = accountAPI.SQLDBWrites.Table(accountsTable).Where("account_id= ?", ID).
		Updates(accountDB).Error
	if err != nil {
		return nil, errs.FailedToUpdate("account", err)
	}

	return accountAPI.updateSession(ctx, accountDB, "")
}

func (accountAPI *accountAPIServer) RefreshSession(
	ctx context.Context, req *account.RefreshSessionRequest,
) (*account.SignInResponse, error) {
	var (
		ID  int
		err error
	)
	// Validation
	switch {
	case req == nil:
		return nil, errs.NilObject("RefreshSessionRequest")
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
	ok, err := accountAPI.RedisDBWrites.SIsMember(ctx, refreshTokenSet(), req.RefreshToken).Result()
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
	err = accountAPI.SQLDBWrites.First(accountDB, "account_id=?", ID).Error
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
	var (
		ID  int
		err error
	)
	// Validation 1
	switch {
	case activateReq == nil:
		return nil, errs.NilObject("ActivateAccountRequest")
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
	payload, err := accountAPI.AuthAPI.AuthorizeActorOrGroup(
		auth.AddTokenMD(ctx, activateReq.Token), activateReq.AccountId, accountAPI.AuthAPI.AdminGroups()...,
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

	// Compare if account account_id matches or if activated by admin
	isOwner := payload.ID == activateReq.AccountId
	isAdmin := accountAPI.AuthAPI.IsAdmin(payload.Group)
	if isOwner == false && isAdmin == false {
		if !dev {
			switch {
			case isAdmin == false:
				return nil, errs.WrapMessage(codes.PermissionDenied, "not admin user")
			case isOwner == false:
				return nil, errs.TokenCredentialNotMatching("account id")
			}
		}
	}

	// Check that account exists
	if errors.Is(accountAPI.SQLDBWrites.Select("account_state").
		First(&Account{}, "account_id=?", ID).Error, gorm.ErrRecordNotFound) {
		return nil, errs.DoesNotExist("account", activateReq.AccountId)
	}

	// Update the model of the user to activate their account
	err = accountAPI.SQLDBWrites.Table(accountsTable).Where("account_id=?", ID).
		Update("account_state", account.AccountState_ACTIVE.String()).Error
	if err != nil {
		return nil, errs.FailedToUpdate("account", err)
	}

	return &account.ActivateAccountResponse{}, nil
}

func emailOrPhone(err error, accountDB *Account) string {
	if strings.Contains(strings.ToLower(err.Error()), "email") {
		return "email " + accountDB.Email
	}
	if strings.Contains(strings.ToLower(err.Error()), "phone") {
		return "phone " + accountDB.Phone
	}
	return fmt.Sprintf("id %v", accountDB.AccountID)
}

func (accountAPI *accountAPIServer) UpdateAccount(
	ctx context.Context, updateReq *account.UpdateAccountRequest,
) (*empty.Empty, error) {
	var err error
	// Validation
	switch {
	case updateReq == nil:
		return nil, errs.NilObject("UpdateRequest")
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

	// Authorization
	payload, err := accountAPI.AuthAPI.AuthorizeActorOrGroup(ctx, updateReq.GetAccount().GetAccountId(), accountAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, err
	}

	// GetAccount the account details from database
	accountDB := &Account{}
	err = accountAPI.SQLDBWrites.Select("account_state").
		First(accountDB, "account_id=?", updateReq.Account.AccountId).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.DoesNotExist("account", updateReq.Account.AccountId)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	// Check that account is not blocked or deleted
	switch {
	case accountDB.AccountState == account.AccountState_BLOCKED.String(),
		accountDB.AccountState == account.AccountState_DELETED.String():
		return nil, errs.WrapMessage(codes.PermissionDenied, "account is blocked or deleted")
	}

	accountDBX, err := GetAccountDB(updateReq.Account)
	if err != nil {
		return nil, err
	}

	if accountAPI.AuthAPI.IsAdmin(payload.Group) == false {
		// Update the model; omit "id", "primary_group", "account_state" and "security profile"
		err = accountAPI.SQLDBWrites.Model(accountDBX).
			Omit("id", "primary_group", "account_state", "password", "security_answer", "security_question").
			Where("account_id=?", updateReq.Account.AccountId).
			Updates(accountDBX).Error
	} else {
		err = accountAPI.SQLDBWrites.Model(accountDBX).
			Where("account_id=?", updateReq.Account.AccountId).
			Updates(accountDBX).Error
	}
	if err != nil {
		return nil, errs.FailedToUpdate("account", err)
	}

	return &empty.Empty{}, nil
}

func updateToken(accountID string) string {
	return "updatetoken:" + accountID
}

func (accountAPI *accountAPIServer) RequestChangePrivateAccount(
	ctx context.Context, req *account.RequestChangePrivateAccountRequest,
) (*account.RequestChangePrivateAccountResponse, error) {
	var err error

	// Validation
	switch {
	case req == nil:
		return nil, errs.NilObject("RequestChangePrivateAccountRequest")
	case req.Payload == "":
		return nil, errs.MissingField("payload")
	case req.FallbackUrl == "":
		return nil, errs.MissingField("fallback url")
	case req.SendMethod == messaging.SendMethod_SEND_METHOD_UNSPECIFIED:
		return nil, errs.WrapMessage(codes.InvalidArgument, "send method is unspecified")
	}

	// GetAccount the user from database
	accountDB := &Account{}
	err = accountAPI.SQLDBWrites.
		First(accountDB, "email=? OR phone=?", req.Payload, req.Payload).Error
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

	accountID := fmt.Sprint(accountDB.AccountID)

	// Authorize the actor
	_, err = accountAPI.AuthAPI.AuthorizeActorOrGroup(ctx, accountID, accountAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, errs.WrapErrorWithMsg(err, "failed to authorize actor")
	}

	uniqueNumber := rand.Intn(499999) + 500000

	// Set token with expiration of 6 hours
	err = accountAPI.RedisDBWrites.Set(
		ctx, updateToken(accountID), uniqueNumber, time.Duration(time.Hour*6),
	).Err()
	if err != nil {
		return nil, errs.RedisCmdFailed(err, "SET")
	}

	// GetAccount jwt
	jwtToken, err := accountAPI.AuthAPI.GenToken(ctx, &auth.Payload{
		ID:    accountID,
		Names: accountDB.Names,
	}, time.Now().Add(6*time.Hour))
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to generate token")
	}

	link := fmt.Sprintf("%s?token=%s&account_id=%s&passphrase=%d", req.FallbackUrl, jwtToken, accountID, uniqueNumber)

	ctx, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 5*time.Second)
	defer cancel()

	// Send message
	_, err = accountAPI.MessagingClient.SendMessage(ctx, &messaging.Message{
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
			"app_name":     firstVal(req.GetSender().GetAppName(), accountAPI.AppName, "Accounts API"),
			"display_name": firstVal(req.GetSender().GetEmailDisplayName(), accountAPI.EmailDisplayName, "Accounts API"),
			"sender":       firstVal(req.GetSender().GetEmailSender(), accountAPI.DefaultEmailSender),
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
	var err error

	// Validation
	var ID int
	switch {
	case updatePrivateReq == nil:
		return nil, errs.NilObject("UpdatePrivateRequest")
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

	// Authorization
	_, err = accountAPI.AuthAPI.AuthorizeActorOrGroup(ctx, updatePrivateReq.AccountId, accountAPI.AuthAPI.AdminGroups()...)
	if err != nil {
		return nil, err
	}

	// GetAccount the account details from database
	accountDB := &Account{}
	err = accountAPI.SQLDBWrites.Select("account_state").First(accountDB, "account_id=?", ID).Error
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

	// Lets get the update token
	token, err := accountAPI.RedisDBWrites.Get(ctx, updateToken(updatePrivateReq.AccountId)).Result()
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

	// Hash the password if not empty
	if updatePrivateReq.PrivateAccount.Password != "" {
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
	err = accountAPI.SQLDBWrites.Model(privateDB).Where("account_id=?", ID).Updates(privateDB).Error
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
	_, err := accountAPI.AuthAPI.AuthorizeActorOrGroup(ctx, delReq.AccountId, accountAPI.AuthAPI.AdminGroups()...)
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
	err = accountAPI.SQLDBWrites.Select("account_state").First(accountDB, "account_id=?", ID).Error
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
	err = accountAPI.SQLDBWrites.Delete(accountDB, "account_id=?", ID).Error
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
	payload, err := accountAPI.AuthAPI.AuthorizeActorOrGroup(ctx, getReq.AccountId, accountAPI.AuthAPI.AdminGroups()...)
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
		if accountAPI.AuthAPI.IsAdmin(payload.Group) {
			err = accountAPI.SQLDBWrites.Unscoped().First(accountDB, "account_id=?", ID).Error
		} else {
			err = accountAPI.SQLDBWrites.First(accountDB, "account_id=?", ID).Error
		}
	} else {
		err = accountAPI.SQLDBWrites.First(accountDB, "account_id=?", ID).Error
	}
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.DoesNotExist("account", getReq.AccountId)
	default:
		return nil, errs.FailedToFind("account", err)
	}

	// Account should not be deleted
	if accountDB.DeletedAt.Valid && !getReq.Priviledge {
		return nil, errs.DoesExist("account", getReq.AccountId)
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
	err := accountAPI.AuthAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	var (
		projectID = existReq.GetProjectId()
		email     = existReq.GetEmail()
		phone     = existReq.GetPhone()
	)

	// Validation
	switch {
	case projectID == "":
		return nil, errs.MissingField("project id")
	case email == "" && phone == "":
		return nil, errs.MissingField("email, phone or external id")
	}

	// Fix phone
	phone = fixPhone(phone)

	accountDB := &Account{}

	// Query for account with email or phone
	err = accountAPI.SQLDBWrites.Select("account_id,email,phone").
		First(accountDB, "(phone=? OR email=?) AND project_id=?", phone, email, projectID).Error
	switch {
	case err == nil:
		existingFields := make([]string, 0)
		if accountDB.Email == email {
			existingFields = append(existingFields, "email")
		}
		if accountDB.Phone == phone {
			existingFields = append(existingFields, "phone")
		}
		// Account exist
		return &account.ExistAccountResponse{
			Exists:         true,
			AccountId:      fmt.Sprint(accountDB.AccountID),
			ExistingFields: existingFields,
		}, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Account doesn't exist
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
	payload, err := accountAPI.AuthAPI.AuthenticateRequestV2(ctx)
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
		ids, err := accountAPI.PaginationHasher.DecodeInt64WithError(listReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(
				codes.InvalidArgument, err, "failed to parse page token",
			)
		}
		id = uint(ids[0])
	}

	// Apply filter criterias
	db := generateWhereCondition(accountAPI.SQLDBReads, listReq.GetListCriteria()).Debug()

	// For admins
	for _, group := range accountAPI.AuthAPI.AdminGroups() {
		if payload.Group == group {
			db = db.Unscoped()
			break
		}
	}

	// ID filter
	if id > 0 {
		db = db.Where("account_id<?", id)
	}

	// Apply project filter
	if payload.ProjectID != "" {
		db = db.Where("project_id=?", payload.ProjectID)
	}

	// Order by ID
	db = db.Limit(int(pageSize) + 1).Order("account_id DESC").Clauses(hints.ForceIndex("PRIMARY").ForOrderBy())

	accountsDB := make([]*Account, 0, pageSize+1)

	err = db.Find(&accountsDB).Error
	switch {
	case err == nil:
	default:
		return nil, errs.FailedToFind("accounts", err)
	}

	accountsPB := make([]*account.Account, 0, len(accountsDB))
	pageSize2 := int(pageSize)

	for i, accountDB := range accountsDB {
		accountPB, err := GetAccountPB(accountDB)
		if err != nil {
			return nil, err
		}

		if i == pageSize2 {
			break
		}

		accountsPB = append(accountsPB, GetAccountPBView(accountPB, listReq.GetView()))
		id = accountDB.AccountID
	}

	var token string
	if len(accountsDB) > pageSize2 {
		// Next page token
		token, err = accountAPI.PaginationHasher.EncodeInt64([]int64{int64(id)})
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
	payload, err := accountAPI.AuthAPI.AuthenticateRequestV2(ctx)
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
		ids, err := accountAPI.PaginationHasher.DecodeInt64WithError(searchReq.GetPageToken())
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to parse page token")
		}
		ID = uint(ids[0])
	}

	accountsDB := make([]*Account, 0, pageSize)

	// Apply filter criterias
	db := generateWhereCondition(accountAPI.SQLDBReads, searchReq.GetSearchCriteria())

	// For admins
	for _, group := range accountAPI.AuthAPI.AdminGroups() {
		if payload.Group == group {
			db = db.Unscoped()
			break
		}
	}

	// Apply project project
	if payload.ProjectID != "" {
		db = db.Where("project_id=?", payload.ProjectID)
	}

	// Order by ID
	db = db.Limit(int(pageSize)).Order("account_id DESC")

	parsedQuery := dbutil.ParseQuery(searchReq.Query)

	// "names", "email", "phone", "linked_accounts"
	err = db.Find(&accountsDB, "MATCH(names, email, phone, linked_accounts) AGAINST(? IN BOOLEAN MODE)", parsedQuery).
		Error
	switch {
	case err == nil:
	default:
		return nil, errs.FailedToFind("accounts", err)
	}

	accountsPB := make([]*account.Account, 0, len(accountsDB))
	pageSize2 := int(pageSize)

	for i, accountDB := range accountsDB {
		accountPB, err := GetAccountPB(accountDB)
		if err != nil {
			return nil, err
		}

		if pageSize2 == i {
			break
		}

		accountsPB = append(accountsPB, GetAccountPBView(accountPB, searchReq.GetView()))
		ID = accountDB.AccountID
	}

	var token string
	if len(accountsDB) > pageSize2 {
		// Next page token
		token, err = accountAPI.PaginationHasher.EncodeInt64([]int64{int64(ID)})
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
		db = db.Where("gender = ?", account.Account_FEMALE.String())
	case criteria.ShowMales:
		db = db.Where("gender = ?", account.Account_MALE.String())
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
