package account

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/messaging"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/errs"
	"github.com/gidyon/services/pkg/utils/mdutil"
	"github.com/gidyon/services/pkg/utils/templateutil"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func (accountAPI *accountAPIServer) validateAdminUpdateAccountRequest(
	ctx context.Context, changeReq *account.AdminUpdateAccountRequest,
) (*Account, error) {
	// Request should not be nil
	if changeReq == nil {
		return nil, errs.NilObject("AdminUpdateAccountRequest")
	}

	// Authorize the admin
	_, err := accountAPI.authAPI.AuthorizeGroups(ctx, auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	accountID := changeReq.GetAccountId()
	adminID := changeReq.GetAdminId()

	// Validation
	var ID, ID2 int
	switch {
	case accountID == "":
		return nil, errs.MissingField("account id")
	case adminID == "":
		return nil, errs.MissingField("admin id")
	case changeReq.UpdateOperation == account.UpdateOperation_UPDATE_OPERATION_INSPECIFIED:
		return nil, errs.WrapMessage(codes.InvalidArgument, "update operation is uknown")
	default:
		ID, err = strconv.Atoi(adminID)
		if err != nil {
			return nil, errs.WrapMessage(codes.InvalidArgument, "incorrect admin id")
		}
		ID2, err = strconv.Atoi(accountID)
		if err != nil {
			return nil, errs.WrapMessage(codes.InvalidArgument, "incorrect account id")
		}
	}

	// Get admin
	admin := &Account{}
	err = accountAPI.sqlDBWrites.Unscoped().Select("account_state,primary_group").
		First(admin, "account_id=?", ID).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.WrapMessagef(codes.NotFound, "admin account with id: %s doesn't exist", accountID)
	default:
		return nil, errs.SQLQueryFailed(err, "GET")
	}

	// Admin account must be active
	if admin.AccountState != account.AccountState_ACTIVE.String() {
		return nil, errs.WrapMessage(codes.PermissionDenied, "account not active")
	}

	// Admin must be admin
	if admin.PrimaryGroup != auth.AdminGroup() {
		return nil, errs.WrapMessage(codes.PermissionDenied, "only admins allowed")
	}

	// Get user
	accountDB := &Account{}
	err = accountAPI.sqlDBWrites.Unscoped().First(accountDB, "account_id=?", ID2).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.WrapMessagef(codes.NotFound, "account with id: %s doesn't exist", accountID)
	default:
		return nil, errs.SQLQueryFailed(err, "SELECT")
	}

	return accountDB, nil
}

func (accountAPI *accountAPIServer) AdminUpdateAccount(
	ctx context.Context, updateReq *account.AdminUpdateAccountRequest,
) (*empty.Empty, error) {
	// Validate the request, super admin credentials and account owner
	accountDB, err := accountAPI.validateAdminUpdateAccountRequest(ctx, updateReq)
	if err != nil {
		return nil, err
	}

	var (
		fullName    = accountDB.Names
		messageType messaging.MessageType
		title       string
		data        string
		link        string
	)

	// Start a transaction
	tx := accountAPI.sqlDBWrites.Begin()
	defer func() {
		if err := recover(); err != nil {
			accountAPI.logger.Errorln(err)
		}
	}()

	if tx.Error != nil {
		return nil, errs.FailedToBeginTx(err)
	}

	// Update the model
	tx = tx.Unscoped().Model(&Account{}).Where("account_id=?", updateReq.AccountId)

	switch updateReq.UpdateOperation {
	case account.UpdateOperation_UNDELETE:
		err = tx.Update("deleted_at", nil).Error
		if err != nil {
			tx.Rollback()
			return nil, errs.WrapErrorWithMsg(err, "failed to delete account")
		}
		messageType = messaging.MessageType_INFO
		title = "Your Account Has Been Restored"
		data = fmt.Sprintf(
			"Hello %s, we are glad to inform you that your account has been restored", fullName,
		)

	case account.UpdateOperation_DELETE:
		err = tx.Update("deleted_at", time.Now()).Error
		if err != nil {
			tx.Rollback()
			return nil, errs.WrapErrorWithMsg(err, "failed to undelete account")
		}
		messageType = messaging.MessageType_ALERT
		title = "Your Account Has Been Deleted"
		data = fmt.Sprintf(
			"Hello %s, we are sad to inform you that your account has been deleted", fullName,
		)

	case account.UpdateOperation_UNBLOCK:
		// The state must be blocked in order to unblock it
		if accountDB.AccountState != account.AccountState_BLOCKED.String() {
			tx.Rollback()
			return nil, errs.WrapMessage(codes.FailedPrecondition, "account is not blocked")
		}
		err = tx.Update("account_state", account.AccountState_ACTIVE.String()).Error
		if err != nil {
			tx.Rollback()
			return nil, errs.WrapErrorWithMsg(err, "failed to unblock account")
		}
		messageType = messaging.MessageType_INFO
		title = "Your Account Has Been Unblock"
		data = fmt.Sprintf(
			"Hello %s, we are glad to inform you that your account has been unblocked", fullName,
		)

	case account.UpdateOperation_BLOCK:
		// The state must be active in order to block it
		if accountDB.AccountState != account.AccountState_ACTIVE.String() {
			tx.Rollback()
			return nil, errs.WrapMessage(codes.FailedPrecondition, "account is not active")
		}
		err = tx.Update("account_state", account.AccountState_BLOCKED.String()).Error
		if err != nil {
			tx.Rollback()
			return nil, errs.WrapErrorWithMsg(err, "failed to block account")
		}
		messageType = messaging.MessageType_ALERT
		title = "Your Account Has Been Blocked"
		data = fmt.Sprintf(
			"Hello %s, we are sad to inform you that your account has been blocked", fullName,
		)

	case account.UpdateOperation_CHANGE_GROUP:
		// The state must be active in order to change group it
		if accountDB.AccountState != account.AccountState_ACTIVE.String() {
			tx.Rollback()
			return nil, errs.WrapMessage(codes.FailedPrecondition, "account is not active")
		}
		// Update the model
		bs, err := json.Marshal(updateReq.Payload)
		if err != nil {
			tx.Rollback()
			return nil, errs.WrapErrorWithMsg(err, "failed to json unmarshal")
		}
		err = tx.Update("secondary_groups", bs).Error
		if err != nil {
			tx.Rollback()
			return nil, errs.WrapErrorWithMsg(err, "failed to update secondary groups")
		}
		messageType = messaging.MessageType_INFO
		title = "Your Account Has Group Has Been Changed"
		data = fmt.Sprintf(
			"Hello %s, you've been added to the following groups %s", fullName, updateReq.Payload,
		)

	case account.UpdateOperation_ADMIN_ACTIVATE:
		err = tx.Update("account_state", account.AccountState_ACTIVE.String()).Error
		if err != nil {
			tx.Rollback()
			return nil, errs.WrapErrorWithMsg(err, "failed to update secondary groups")
		}
		messageType = messaging.MessageType_INFO
		title = "Your Account Has Been Activated"
		data = fmt.Sprintf(
			"Hello %s. Your account has been activated by the administrator", fullName,
		)
	}

	// Email template
	emailContent := templateutil.EmailData{
		Names:        accountDB.Names,
		AccountID:    updateReq.AccountId,
		AppName:      accountAPI.appName,
		Reason:       updateReq.Reason,
		TemplateName: templateName,
	}

	content := bytes.NewBuffer(make([]byte, 0, 64))
	err = accountAPI.tpl.ExecuteTemplate(content, templateName, emailContent)
	if err != nil {
		tx.Rollback()
		return nil, errs.WrapErrorWithMsg(err, "failed to exucute template")
	}

	// Send message to inform necessary audience
	_, err = accountAPI.messagingClient.SendMessage(mdutil.AddFromCtx(ctx), &messaging.Message{
		UserId:      updateReq.AccountId,
		Title:       title,
		Data:        data,
		Link:        link,
		Save:        true,
		Type:        messageType,
		SendMethods: []messaging.SendMethod{messaging.SendMethod_SMS, messaging.SendMethod_EMAIL},
		Details: map[string]string{
			"email_body": content.String(),
		},
	}, grpc.WaitForReady(true))
	if err != nil {
		tx.Rollback()
		return nil, errs.WrapErrorWithMsg(err, "failed to send message")
	}

	// Commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, errs.WrapErrorWithMsg(err, "failed to commit transation")
	}

	return &empty.Empty{}, nil
}
