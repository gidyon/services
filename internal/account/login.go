package account

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/account"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

func refreshTokenSet() string {
	year, month, day := time.Now().Date()
	return fmt.Sprintf("refreshtokens:%d-%d-%d", year, month, day)
}

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
	case signInReq.ProjectId == "":
		err = errs.MissingField("project id")
	}
	if err != nil {
		return nil, err
	}

	// Check whtether account exist
	accountDB := &Account{}

	// Query for user with email or phone or huduma id
	err = accountAPI.SQLDBWrites.Select("account_id,names,primary_group,account_state,password").First(
		accountDB, "(phone=? OR email=?) AND project_id=?", signInReq.Username, signInReq.Username, signInReq.ProjectId,
	).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
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

	accountPB, err := GetAccountPB(accountDB)
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

	return accountAPI.updateSession(ctx, accountDB, signInReq.GetGroup())
}

func (accountAPI *accountAPIServer) updateSession(
	ctx context.Context, accountDB *Account, signInGroup string,
) (*account.SignInResponse, error) {
	var (
		accountID    = fmt.Sprint(accountDB.AccountID)
		refreshToken = uuid.New().String()
		token        string
		err          error
	)

	// Secondary groups
	secondaryGroups := make([]string, 0)
	if len(accountDB.SecondaryGroups) != 0 {
		err = json.Unmarshal(accountDB.SecondaryGroups, &secondaryGroups)
		if err != nil {
			return nil, errs.WrapErrorWithMsg(err, "failed to json unmarshal")
		}
	}

	signInGroup = strings.ToUpper(signInGroup)

	durStr := os.Getenv("TOKEN_EXPIRATION_MINUTES")
	dur, err := strconv.Atoi(durStr)
	if err != nil {
		dur = 30
	}

	if signInGroup != "" {
		var found bool
		for _, group := range append(secondaryGroups, accountDB.PrimaryGroup) {
			group := strings.ToUpper(strings.TrimSpace(group))
			if group == signInGroup {
				found = true
				// Generates JWT
				token, err = accountAPI.AuthAPI.GenToken(ctx, &auth.Payload{
					ID:        fmt.Sprint(accountDB.AccountID),
					Names:     accountDB.Names,
					Group:     signInGroup,
					ProjectID: accountDB.ProjectID,
				}, time.Now().Add(time.Duration(dur)*time.Minute))
				if err != nil {
					return nil,
						errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to generate token")
				}
				break
			}
		}
		if !found {
			return nil,
				errs.WrapMessagef(codes.InvalidArgument, "group %s not associated with the account", signInGroup)
		}
	} else {
		signInGroup = accountDB.PrimaryGroup
		// Generate JWT
		token, err = accountAPI.AuthAPI.GenToken(ctx, &auth.Payload{
			ID:        accountID,
			Names:     accountDB.Names,
			Group:     accountDB.PrimaryGroup,
			ProjectID: accountDB.ProjectID,
		}, time.Now().Add(time.Duration(dur)*time.Minute))
		if err != nil {
			return nil,
				errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to generate token")
		}
	}

	// Set refresh token
	err = accountAPI.RedisDBWrites.SAdd(ctx, refreshTokenSet(), refreshToken, 0).Err()
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to set refresh token")
	}

	// Set Cookie in response header
	encoded, err := accountAPI.cookier.Encode(auth.JWTCookie(), token)
	if err == nil {
		// JWT cookie
		cookie := &http.Cookie{
			Name:     auth.JWTCookie(),
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(time.Hour * 8760),
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		}
		err = accountAPI.setCookie(ctx, cookie.String())
		if err != nil {
			return nil, err
		}

		// Refresh token
		cookie.Name = auth.RefreshCookie()
		cookie.Value = refreshToken
		cookie.HttpOnly = false
		err = accountAPI.setCookie(ctx, cookie.String())
		if err != nil {
			return nil, err
		}

		// Acccount ID Cookie
		cookie.Name = auth.AccountIDCookie()
		cookie.Value = accountID
		cookie.HttpOnly = false
		err = accountAPI.setCookie(ctx, cookie.String())
		if err != nil {
			return nil, err
		}
	}

	// Return token
	return &account.SignInResponse{
		AccountId:       accountID,
		Token:           token,
		RefreshToken:    refreshToken,
		State:           account.AccountState(account.AccountState_value[accountDB.AccountState]),
		Group:           signInGroup,
		SecondaryGroups: secondaryGroups,
	}, nil
}
