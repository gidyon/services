package account

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/auth"
	"github.com/gidyon/services/pkg/utils/errs"
	"github.com/jinzhu/gorm"
	"google.golang.org/grpc/codes"
)

func (accountAPI *accountAPIServer) SignIn(
	ctx context.Context, signInReq *account.SignInRequest,
) (*account.SignInResponse, error) {
	// Request should not be nil
	if signInReq == nil {
		return nil, errs.NilObject("SignInRequest")
	}

	// Validation
	var err error
	switch {
	case signInReq.Username == "":
		err = errs.MissingField("username")
	case signInReq.Password == "":
		err = errs.MissingField("password")
	}
	if err != nil {
		return nil, err
	}

	accountDB := &Account{}

	// Query for user with email or phone or huduma id
	err = accountAPI.sqlDB.Select("id,names,primary_group,account_state,password").First(
		accountDB, "phone=? OR email=?", signInReq.Username, signInReq.Username,
	).Error
	switch {
	case err == nil:
	case gorm.IsRecordNotFoundError(err):
		emailOrPhone := func() string {
			if strings.Contains(signInReq.Username, "@") {
				return "email " + signInReq.Username
			}
			if strings.Contains(signInReq.Username, "+") {
				return "phone " + signInReq.Username
			}
			return "username " + signInReq.Username
		}
		return nil, errs.WrapMessagef(codes.NotFound, "account with %s not found", emailOrPhone())
	default:
		return nil, errs.SQLQueryFailed(err, "LOGIN")
	}

	// If no password set in account
	if accountDB.Password == "" {
		return nil, errs.WrapMessage(
			codes.PermissionDenied, "account has no password; please request new password",
		)
	}

	accountPB, err := getAccountPB(accountDB)
	if err != nil {
		return nil, err
	}

	// Check that account is not blocked
	if accountPB.State == account.AccountState_BLOCKED {
		return nil, errs.WrapMessage(codes.FailedPrecondition, "account blocked")
	}

	// Check if password match if they logged in with Phone or Email
	err = compareHash(accountDB.Password, signInReq.Password)
	if err != nil {
		return nil, errs.WrapMessage(codes.Internal, "wrong password")
	}

	// Secondary groups
	secondaryGroups := make([]string, 0)
	if len(accountDB.SecondaryGroups) != 0 {
		err = json.Unmarshal(accountDB.SecondaryGroups, &secondaryGroups)
		if err != nil {
			return nil, errs.WrapErrorWithMsg(err, "failed to json unmarshal")
		}
	}

	var token string
	signInGroup := strings.ToUpper(signInReq.GetGroup())
	if signInGroup != "" {
		var found bool
		for _, group := range append(secondaryGroups, accountDB.PrimaryGroup) {
			group := strings.ToUpper(strings.TrimSpace(group))
			if group == signInGroup {
				found = true
				// Generates the token with claims from profile object
				token, err = accountAPI.authAPI.GenToken(ctx, &auth.Payload{
					ID:           fmt.Sprint(accountDB.ID),
					Names:        accountDB.Names,
					PhoneNumber:  accountDB.Phone,
					EmailAddress: accountDB.Email,
					Group:        signInGroup,
				}, 0)
				if err != nil {
					return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to generate token")
				}
				break
			}
		}
		if !found {
			return nil,
				errs.WrapMessagef(codes.InvalidArgument, "group %s not associated with the account", signInGroup)
		}
	}

	// Set Cookie in response header
	encoded, err := accountAPI.cookier.Encode(auth.CookieName(), token)
	if err == nil {
		cookie := &http.Cookie{
			Name:     auth.CookieName(),
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
		}
		err = accountAPI.setCookie(ctx, cookie.String())
		if err != nil {
			return nil, err
		}
	}

	// Return token
	return &account.SignInResponse{
		Token:     token,
		AccountId: fmt.Sprint(accountDB.ID),
		State:     accountPB.State,
		Group:     signInGroup,
	}, nil
}
