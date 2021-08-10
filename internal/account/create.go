package account

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/dbutil"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/micro/v2/utils/mdutil"
	"github.com/gidyon/services/pkg/api/messaging"

	"github.com/gidyon/services/pkg/api/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func fixPhone(phone string) string {
	phone = strings.TrimPrefix(phone, "+")
	if strings.HasPrefix(phone, "7") {
		phone = fmt.Sprint("254", phone)
	}
	if strings.HasPrefix(phone, "07") {
		phone = fmt.Sprint("254", phone[1:])
	}
	return phone
}

func (accountAPI *accountAPIServer) CreateAccount(
	ctx context.Context, createReq *account.CreateAccountRequest,
) (*account.CreateAccountResponse, error) {
	// Request should not be nil
	if createReq == nil {
		return nil, errs.NilObject("CreateAccountRequest")
	}

	var err error

	accountPB := createReq.GetAccount()
	if accountPB == nil {
		return nil, errs.NilObject("Account")
	}

	// Validation
	switch {
	case createReq.ProjectId == "":
		err = errs.MissingField("project id")
	case accountPB.Group == "":
		err = errs.MissingField("group")
	case accountPB.Names == "":
		err = errs.MissingField("names")
	case accountPB.Phone == "" && accountPB.Email == "":
		err = errs.MissingField("phone and email and hiduma id")
	case createReq.GetByAdmin() && createReq.AdminId == "":
		err = errs.MissingField("admin id")
	}
	if err != nil {
		return nil, err
	}

	// Check if account already exists
	existRes, err := accountAPI.ExistAccount(ctx, &account.ExistAccountRequest{
		Email:     accountPB.Email,
		Phone:     accountPB.Phone,
		ProjectId: createReq.ProjectId,
	})
	if err != nil {
		return nil, err
	}

	// Fails if account already exists
	if existRes.Exists {
		return nil, errs.WrapMessagef(
			codes.AlreadyExists,
			"account with %s already exists",
			strings.Join(existRes.ExistingFields, " and "),
		)
	}

	accountDB, err := GetAccountDB(accountPB)
	if err != nil {
		return nil, err
	}

	// Fix phone number
	accountDB.Phone = fixPhone(accountDB.Phone)

	accountState := account.AccountState_INACTIVE

	if createReq.GetByAdmin() {
		// Authenticate the admin
		p, err := accountAPI.AuthAPI.AuthorizeGroup(ctx, accountAPI.AuthAPI.AdminGroups()...)
		if err != nil {
			return nil, err
		}
		if p.ID != createReq.AdminId {
			dev := (os.Getenv("MODE") == "development")
			if !dev {
				return nil, errs.WrapMessage(codes.Unauthenticated, "token id and admin id do not match")
			}
		}
		accountState = account.AccountState_ACTIVE
	}

	accountDB.AccountState = accountState.String()

	accountDB.ProjectID = createReq.ProjectId

	accountPrivate := createReq.GetPrivateAccount()
	if accountPrivate != nil {
		accountDB.SecurityAnswer = accountPrivate.GetSecurityQuestion()
		// Store password as encrypted
		accountDB.SecurityAnswer = accountPrivate.GetSecurityAnswer()
		if accountPrivate.Password != "" {
			newPass, err := genHash(accountPrivate.GetPassword())
			if err != nil {
				return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to generate hash password")
			}
			accountDB.Password = newPass
		}
	}

	// Start a transaction
	tx := accountAPI.SQLDBWrites.Begin(&sql.TxOptions{
		Isolation: sql.IsolationLevel(0),
	})
	defer func() {
		if err := recover(); err != nil {
			accountAPI.Logger.Errorf("recovering from panic: %v", err)
		}
	}()

	if tx.Error != nil {
		tx.Rollback()
		return nil, errs.FailedToBeginTx(err)
	}

	err = tx.Create(accountDB).Error
	switch {
	case err == nil:
	default:
		emailOrPhone := func(err error) (string, string) {
			if strings.Contains(strings.ToLower(err.Error()), "email") {
				return "email", accountDB.Email
			}
			if strings.Contains(strings.ToLower(err.Error()), "phone") {
				return "phone", accountDB.Phone
			}
			return "id", fmt.Sprint(accountDB.AccountID)
		}

		if dbutil.IsDuplicate(err) {
			// Upsert must be true
			if createReq.GetUpdateOnly() && createReq.GetByAdmin() {
				// Update account instead
				err = accountAPI.SQLDBWrites.Table(accountsTable).Updates(accountDB).Error
				if err != nil {
					tx.Rollback()
					return nil, errs.FailedToUpdate("account", err)
				}
				break
			}

			tx.Rollback()
			return nil, errs.DuplicateField(emailOrPhone(err))
		}

		tx.Rollback()
		return nil, errs.SQLQueryFailed(err, "CREATE")
	}

	accountID := fmt.Sprint(accountDB.AccountID)

	// Commit transaction
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, errs.FailedToCommitTx(err)
	}

	if !createReq.GetUpdateOnly() && createReq.Notify {
		// Generate jwt token with expiration of 6 hours
		jwtToken, err := accountAPI.AuthAPI.GenToken(ctx, &auth.Payload{
			ID:           accountID,
			Names:        accountPB.Names,
			PhoneNumber:  accountPB.Phone,
			EmailAddress: accountPB.Email,
		}, time.Now().Add(time.Duration(6*time.Hour)))
		if err != nil {
			return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to generate token")
		}

		// Send method
		sendMethods := func() []messaging.SendMethod {
			if accountPB.Email != "" {
				return []messaging.SendMethod{messaging.SendMethod_EMAIL}
			}
			if accountPB.Phone != "" {
				return []messaging.SendMethod{messaging.SendMethod_SMSV2}
			}
			return []messaging.SendMethod{messaging.SendMethod_EMAIL, messaging.SendMethod_SMSV2}
		}()

		appName := firstVal(createReq.GetSender().GetAppName(), createReq.GetSmsAuth().GetAppName(), accountAPI.AppName)

		// CreateAccount message
		messagePB := &messaging.Message{
			UserId:      accountID,
			Title:       fmt.Sprintf("%s Account created successfully", appName),
			Data:        fmt.Sprintf("Hello %s. Your %s account was created successfully, but you'll need to verify and activate the account", accountDB.Names, appName),
			Link:        fmt.Sprintf("%s?token=%s?&account_id=%s", accountAPI.activationURL, jwtToken, accountID),
			Save:        true,
			Type:        messaging.MessageType_REMINDER,
			SendMethods: sendMethods,
		}

		if createReq.GetByAdmin() {
			messagePB = &messaging.Message{
				UserId: accountID,
				Title:  fmt.Sprintf("%s Account created successfully by Admin", appName),
				Data: fmt.Sprintf(
					"Hello %s. %s account has been created successfully by the administrator. You can now sign in to your account.",
					accountDB.Names, appName,
				),
				Save:        true,
				Type:        messaging.MessageType_REMINDER,
				SendMethods: sendMethods,
			}
		}

		md := metadata.Pairs(auth.Header(), fmt.Sprintf("%s %s", auth.Scheme(), jwtToken))

		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		// Send message
		_, err = accountAPI.MessagingClient.SendMessage(mdutil.AddMD(ctx, md), &messaging.SendMessageRequest{
			Message: messagePB,
			SmsAuth: createReq.GetSmsAuth(),
			Sender:  createReq.GetSender(),
		})
		if err != nil {
			accountAPI.Logger.Errorf("error while sending account creation message: %v", err)
		}
	}

	return &account.CreateAccountResponse{
		AccountId: accountID,
	}, nil
}

func firstVal(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
