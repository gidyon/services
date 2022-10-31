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

	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/micro/v2/utils/mdutil"
	"github.com/gidyon/micro/v2/utils/templateutil"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/api/messaging"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func (accountAPI *accountAPIServer) validateAdmin(
	ctx context.Context, adminID string,
) error {
	// Get admin
	admin := &Account{}
	err := accountAPI.SQLDBWrites.Unscoped().Select("account_state,primary_group").
		First(admin, "account_id=?", adminID).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return errs.WrapMessagef(codes.NotFound, "admin account with id: %s doesn't exist", adminID)
	default:
		accountAPI.Logger.Errorln(err)
		return errs.WrapMessage(codes.Internal, "getting admin failed")
	}

	// Admin account must be active
	if admin.AccountState != account.AccountState_ACTIVE.String() {
		return errs.WrapMessage(codes.PermissionDenied, "admin account not active")
	}

	return nil
}

func (accountAPI *accountAPIServer) getUser(
	ctx context.Context, userID string,
) (*Account, error) {

	// Get user
	db := &Account{}
	err := accountAPI.SQLDBWrites.Unscoped().First(db, "account_id=?", userID).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.WrapMessagef(codes.NotFound, "account with id: %s doesn't exist", userID)
	default:
		return nil, errs.SQLQueryFailed(err, "SELECT")
	}

	return db, nil
}

func (accountAPI *accountAPIServer) AdminUpdateAccount(
	ctx context.Context, req *account.AdminUpdateAccountRequest,
) (*empty.Empty, error) {

	// Validation
	switch {
	case req == nil:
		return nil, errs.MissingField("request")
	case req.AccountId == "":
		return nil, errs.MissingField("account id")
	case req.AdminId == "":
		return nil, errs.MissingField("admin id")
	case req.UpdateOperation == account.UpdateOperation_UPDATE_OPERATION_INSPECIFIED:
		return nil, errs.WrapMessage(codes.InvalidArgument, "update operation is uknown")
	default:
		_, err := strconv.Atoi(req.AdminId)
		if err != nil {
			return nil, errs.WrapMessage(codes.InvalidArgument, "incorrect admin id")
		}
		_, err = strconv.Atoi(req.AccountId)
		if err != nil {
			return nil, errs.WrapMessage(codes.InvalidArgument, "incorrect account id")
		}
	}

	// Authorize the admin
	_, err := accountAPI.AuthAPI.AuthorizeAdminStrict(ctx, req.AdminId)
	if err != nil {
		return nil, err
	}

	err = accountAPI.validateAdmin(ctx, req.AdminId)
	if err != nil {
		return nil, err
	}

	db, err := accountAPI.getUser(ctx, req.AccountId)
	if err != nil {
		return nil, err
	}

	var (
		fullName    = db.Names
		messageType messaging.MessageType
		title       string
		data        string
		link        string
	)

	// Start a transaction
	err = accountAPI.SQLDBWrites.Transaction(func(tx *gorm.DB) error {

		tx = tx.Model(&Account{}).Where("account_id=?", req.AccountId)

		switch req.UpdateOperation {
		case account.UpdateOperation_UNDELETE:
			err = tx.Update("deleted_at", nil).Error
			if err != nil {
				return errs.WrapMessage(codes.Internal, "failed to delete account")
			}
			messageType = messaging.MessageType_INFO
			title = "Your Account Has Been Restored"
			data = fmt.Sprintf(
				"Hello %s, we are glad to inform you that your account has been restored", fullName,
			)

		case account.UpdateOperation_DELETE:
			err = tx.Update("deleted_at", time.Now()).Error
			if err != nil {

				return errs.WrapMessage(codes.Internal, "failed to undelete account")
			}
			messageType = messaging.MessageType_ALERT
			title = "Your Account Has Been Deleted"
			data = fmt.Sprintf(
				"Hello %s, we are sad to inform you that your account has been scheduled for deletion", fullName,
			)

		case account.UpdateOperation_UNBLOCK:
			// The state must be blocked in order to unblock it
			if db.AccountState != account.AccountState_BLOCKED.String() {
				return errs.WrapMessage(codes.FailedPrecondition, "account is not blocked")
			}
			err = tx.Update("account_state", account.AccountState_ACTIVE.String()).Error
			if err != nil {
				return errs.WrapMessage(codes.Internal, "failed to unblock account")
			}
			messageType = messaging.MessageType_INFO
			title = "Your Account Has Been Unblock"
			data = fmt.Sprintf(
				"Hello %s, we are glad to inform you that your account has been unblocked", fullName,
			)

		case account.UpdateOperation_BLOCK:
			// The state must be active in order to block it
			if db.AccountState != account.AccountState_ACTIVE.String() {
				return errs.WrapMessage(codes.FailedPrecondition, "account is not active")
			}
			err = tx.Update("account_state", account.AccountState_BLOCKED.String()).Error
			if err != nil {
				return errs.WrapMessage(codes.Internal, "failed to block account")
			}
			messageType = messaging.MessageType_ALERT
			title = "Your Account Has Been Blocked"
			data = fmt.Sprintf(
				"Hello %s, we are sad to inform you that your account has been blocked", fullName,
			)

		case account.UpdateOperation_CHANGE_GROUP:
			if len(req.Payload) == 0 {
				return errs.MissingField("payload")
			}
			// The state must be active in order to change group it
			if db.AccountState != account.AccountState_ACTIVE.String() {
				return errs.WrapMessage(codes.FailedPrecondition, "account to change group is not active")
			}
			// Update the model
			bs, err := json.Marshal(req.Payload)
			if err != nil {
				return errs.WrapMessage(codes.Internal, "failed to json unmarshal")
			}
			err = tx.Update("secondary_groups", bs).Error
			if err != nil {
				return errs.WrapMessage(codes.Internal, "failed to update secondary groups")
			}
			messageType = messaging.MessageType_INFO
			title = "Your Account Has Group Has Been Changed"
			data = fmt.Sprintf(
				"Hello %s, your account has been added to the following groups %s", fullName, req.Payload,
			)

		case account.UpdateOperation_CHANGE_PRIMARY_GROUP:
			if len(req.Payload) == 0 {
				return errs.MissingField("payload")
			}
			// The state must be active in order to change group it
			if db.AccountState != account.AccountState_ACTIVE.String() {
				return errs.WrapMessage(codes.FailedPrecondition, "account to change group is not active")
			}
			// Update the model
			err = tx.Update("primary_group", req.Payload[0]).Error
			if err != nil {
				return errs.WrapMessage(codes.Internal, "failed to update primary group")
			}
			messageType = messaging.MessageType_INFO
			title = fmt.Sprintf("Your Group Has Been Changed To %s", req.Payload[0])
			data = fmt.Sprintf(
				"Hello %s, your account group has been updated to %s", fullName, req.Payload[0],
			)

		case account.UpdateOperation_ADMIN_ACTIVATE:
			err = tx.Update("account_state", account.AccountState_ACTIVE.String()).Error
			if err != nil {
				return errs.WrapMessage(codes.Internal, "failed to activate account")
			}
			messageType = messaging.MessageType_INFO
			title = "Your Account Has Been Activated"
			data = fmt.Sprintf(
				"Hello %s. Your account has been activated by the administrator", fullName,
			)

		case account.UpdateOperation_PASSWORD_RESET:
			if len(req.Payload) == 0 {
				return errs.MissingField("payload")
			}
			newPass, err := genHash(req.Payload[0])
			if err != nil {
				return errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to generate hash password")
			}
			err = tx.Update("password", newPass).Error
			if err != nil {
				return errs.WrapMessage(codes.Internal, "failed to update password")
			}
			messageType = messaging.MessageType_INFO
			title = "Your Account Pasword Has Been Updated"
			data = fmt.Sprintf(
				"Hello %s. Your account password has been updated by the administrator. <br>New password is: %s", fullName, req.Payload[0],
			)

		case account.UpdateOperation_GROUP_ID:
			if len(req.Payload) == 0 {
				return errs.MissingField("payload")
			}
			err = tx.Update("group_id", req.Payload[0]).Error
			if err != nil {
				return errs.WrapMessage(codes.Internal, "failed to update group id")
			}

		case account.UpdateOperation_PARENT_ID:
			if len(req.Payload) == 0 {
				return errs.MissingField("payload")
			}
			err = tx.Update("parent_id", req.Payload[0]).Error
			if err != nil {
				return errs.WrapMessage(codes.Internal, "failed to update parent id")
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if req.Notify {
		// Email template
		emailContent := templateutil.EmailData{
			Names:        db.Names,
			AccountID:    req.AccountId,
			AppName:      firstVal(req.GetSender().GetAppName(), accountAPI.AppName),
			Reason:       req.Reason,
			TemplateName: templateName,
		}

		var emailData string
		if accountAPI.tpl != nil {
			content := bytes.NewBuffer(make([]byte, 0, 64))
			err = accountAPI.tpl.ExecuteTemplate(content, templateName, emailContent)
			if err != nil {

				return nil, errs.WrapMessage(codes.Internal, "failed to exucute template")
			}
			emailData = content.String()
		} else {
			emailData = data
		}

		ctx, cancel := context.WithTimeout(mdutil.AddFromCtx(ctx), 5*time.Second)
		defer cancel()

		// Send message to inform necessary audience
		_, err = accountAPI.MessagingClient.SendMessage(ctx, &messaging.SendMessageRequest{
			Message: &messaging.Message{
				UserId:      req.AccountId,
				Title:       title,
				Data:        data,
				EmailData:   emailData,
				Link:        link,
				Save:        true,
				Type:        messageType,
				SendMethods: []messaging.SendMethod{messaging.SendMethod_SMSV2, messaging.SendMethod_EMAIL},
			},
			Sender:          req.GetSender(),
			SmsAuth:         req.GetSmsAuth(),
			SmsCredentialId: req.SmsCredentialId,
			FetchSmsAuth:    req.FetchSmsAuth,
		}, grpc.WaitForReady(true))
		if err != nil {
			accountAPI.Logger.Errorf("error while sending account changed message: %v", err)
		}
	}

	return &empty.Empty{}, nil
}
