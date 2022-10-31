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
	ctx context.Context, req *account.SignInRequest,
) (*account.SignInResponse, error) {
	// Request should not be nil
	if req == nil {
		return nil, errs.NilObject("SignInRequest")
	}

	// Validation
	var err error
	switch {
	case req.Username == "":
		err = errs.MissingField("username")
	case req.Password == "":
		err = errs.MissingField("password")
	case req.ProjectId == "":
		err = errs.MissingField("project id")
	}
	if err != nil {
		return nil, err
	}

	// Check whtether account exist
	db := &Account{}

	// Query for user with email or phone or huduma id
	err = accountAPI.SQLDBWrites.First(
		db, "(phone=? OR email=?) AND project_id=?", req.Username, req.Username, req.ProjectId,
	).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		emailOrPhone := func() string {
			if strings.Contains(req.Username, "@") {
				return "email " + req.Username
			}
			if strings.Contains(req.Username, "+") {
				return "phone " + req.Username
			}
			return "username " + req.Username
		}
		return nil, errs.WrapMessagef(codes.NotFound, "account with %s not found", emailOrPhone())
	default:
		return nil, errs.SQLQueryFailed(err, "LOGIN")
	}

	// If no password set in account
	if db.Password == "" {
		return nil, errs.WrapMessage(
			codes.PermissionDenied, "account has no password; please request new password",
		)
	}

	pb, err := AccountProto(db)
	if err != nil {
		return nil, err
	}

	// Check that account is not blocked
	if pb.State == account.AccountState_BLOCKED {
		return nil, errs.WrapMessage(codes.FailedPrecondition, "account blocked")
	}

	// Check if password match if they logged in with Phone or Email
	err = compareHash(db.Password, req.Password)
	if err != nil {
		return nil, errs.WrapMessage(codes.Internal, "wrong password")
	}

	// Update last login
	err = accountAPI.SQLDBWrites.Model(db).Update("last_login", time.Now()).Error
	if err != nil {
		return nil, errs.WrapMessage(codes.Internal, "failed to update last login")
	}

	return accountAPI.updateSession(ctx, db, req.GetGroup())
}

func (accountAPI *accountAPIServer) updateSession(
	ctx context.Context, db *Account, signInGroup string,
) (*account.SignInResponse, error) {
	var (
		accountID    = fmt.Sprint(db.AccountID)
		refreshToken = uuid.New().String()
		token        string
		err          error
	)

	// Secondary groups
	secondaryGroups := make([]string, 0)
	if len(db.SecondaryGroups) != 0 {
		err = json.Unmarshal(db.SecondaryGroups, &secondaryGroups)
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
		for _, group := range append(secondaryGroups, db.PrimaryGroup) {
			group := strings.ToUpper(strings.TrimSpace(group))
			if group == signInGroup {
				found = true
				// Generates JWT
				token, err = accountAPI.AuthAPI.GenToken(ctx, &auth.Payload{
					ID:           fmt.Sprint(db.AccountID),
					Names:        db.Names,
					Group:        signInGroup,
					ProjectID:    db.ProjectID,
					EmailAddress: db.Email,
					PhoneNumber:  db.Phone,
					Roles:        secondaryGroups,
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
		signInGroup = db.PrimaryGroup
		// Generate JWT
		token, err = accountAPI.AuthAPI.GenToken(ctx, &auth.Payload{
			ID:           accountID,
			Names:        db.Names,
			Group:        db.PrimaryGroup,
			ProjectID:    db.ProjectID,
			EmailAddress: db.Email,
			PhoneNumber:  db.Phone,
			Roles:        secondaryGroups,
		}, time.Now().Add(time.Duration(dur)*time.Minute))
		if err != nil {
			return nil,
				errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to generate token")
		}
	}

	// Set refresh token ~ Needs cleanup
	err = accountAPI.RedisDBWrites.SAdd(ctx, refreshTokenSet(), refreshToken, 0).Err()
	if err != nil {
		return nil, errs.WrapErrorWithCodeAndMsg(codes.Internal, err, "failed to set refresh token")
	}

	// Get account
	pb, err := AccountProto(db)
	if err != nil {
		return nil, err
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
		State:           account.AccountState(account.AccountState_value[db.AccountState]),
		Group:           signInGroup,
		SecondaryGroups: secondaryGroups,
		Account:         pb,
	}, nil
}
