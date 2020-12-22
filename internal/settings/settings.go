package settings

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gidyon/micro/pkg/grpc/auth"
	"github.com/gidyon/micro/utils/encryption"
	"github.com/gidyon/micro/utils/errs"
	"github.com/gidyon/services/pkg/api/settings"
	"github.com/speps/go-hashids"
	"google.golang.org/grpc/grpclog"
	"gorm.io/gorm"
)

type settingsAPIServer struct {
	settings.UnimplementedSettingsAPIServer
	hasher  *hashids.HashID
	authAPI auth.API
	*Options
}

// Options contains parameters passed to NewOperationAPIService
type Options struct {
	SQLDB         *gorm.DB
	Logger        grpclog.LoggerV2
	JWTSigningKey []byte
}

// NewSettingsAPI creates a new instance of settings API server
func NewSettingsAPI(ctx context.Context, opt *Options) (settings.SettingsAPIServer, error) {
	// Validation
	var err error
	switch {
	case ctx == nil:
		err = errs.NilObject("context")
	case opt == nil:
		err = errs.NilObject("options")
	case opt.SQLDB == nil:
		err = errs.NilObject("sql db")
	case opt.Logger == nil:
		err = errs.NilObject("logger")
	case opt.JWTSigningKey == nil:
		err = errs.NilObject("jwt key")
	}
	if err != nil {
		return nil, err
	}

	authAPI, err := auth.NewAPI(opt.JWTSigningKey, "Operation API", "users")
	if err != nil {
		return nil, err
	}

	hasher, err := encryption.NewHasher(string(opt.JWTSigningKey))
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash id: %v", err)
	}

	settingsAPI := &settingsAPIServer{
		hasher:  hasher,
		authAPI: authAPI,
		Options: opt,
	}

	return settingsAPI, nil
}

// ValidateSetting validates a setitng resource
func ValidateSetting(settingPB *settings.Setting) error {
	var err error
	switch {
	case settingPB == nil:
		err = errs.NilObject("setting")
	case settingPB.Key == "":
		err = errs.MissingField("setting key")
	case settingPB.Value == "":
		err = errs.MissingField("setting value")
	}
	return err
}

func (settingsAPI *settingsAPIServer) UpdateSetting(
	ctx context.Context, updateReq *settings.UpdateSettingRequest,
) (*settings.UpdateSettingResponse, error) {
	// Authentication
	err := settingsAPI.authAPI.AuthenticateRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Validation
	switch {
	case updateReq == nil:
		return nil, errs.NilObject("UpdateSettingRequest")
	case updateReq.OwnerId == "":
		return nil, errs.NilObject("owner id")
	case updateReq.Settings == nil:
		return nil, errs.NilObject("settings")
	default:
		for _, setting := range updateReq.Settings {
			err = ValidateSetting(setting)
			if err != nil {
				return nil, err
			}
		}
	}

	// get user settings from db
	userSettingsDB := &Model{}

	err = settingsAPI.SQLDB.First(userSettingsDB, "owner_id=?", updateReq.OwnerId).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errs.DoesNotExist("settings", updateReq.OwnerId)
	default:
		return nil, errs.FailedToFind("settings", err)
	}

	userSettingsPB, err := GetSettingsPB(userSettingsDB)
	if err != nil {
		return nil, err
	}

	// Apply settings update
	for key, settingsPB := range updateReq.Settings {
		userSettingsPB.Settings[key].Domain = updateWithDefault(settingsPB.Domain, userSettingsPB.Settings[key].Domain)
		userSettingsPB.Settings[key].Value = updateWithDefault(settingsPB.Value, userSettingsPB.Settings[key].Value)
	}

	userSettingsDB, err = GetSettingsDB(userSettingsPB)
	if err != nil {
		return nil, err
	}

	// Update settings
	err = settingsAPI.SQLDB.Select("settings").Updates(userSettingsDB).Error
	if err != nil {
		return nil, errs.FailedToUpdate("settings", err)
	}

	// Get updated resource
	userSettingsPB, err = GetSettingsPB(userSettingsDB)
	if err != nil {
		return nil, err
	}

	return &settings.UpdateSettingResponse{
		Settings: userSettingsPB.Settings,
	}, nil
}

func updateWithDefault(val, def string) string {
	if val != "" {
		return val
	}
	return def
}

func (settingsAPI *settingsAPIServer) GetSettings(
	ctx context.Context, getReq *settings.GetSettingsRequest,
) (*settings.GetSettingsResponse, error) {
	// Authentication
	_, err := settingsAPI.authAPI.AuthorizeActorOrGroups(ctx, getReq.GetOwnerId(), auth.AdminGroup())
	if err != nil {
		return nil, err
	}

	var ownerID int
	// Validation
	switch {
	case getReq == nil:
		return nil, errs.NilObject("GetSettingsRequest")
	case getReq.OwnerId == "":
		return nil, errs.MissingField("owner id")
	default:
		ownerID, err = strconv.Atoi(getReq.OwnerId)
		if err != nil {
			return nil, errs.IncorrectVal("owner id")
		}
	}

	userSettingsDB := &Model{}

	err = settingsAPI.SQLDB.First(userSettingsDB, "owner_id=?", getReq.OwnerId).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Create settings for user
		settingsAPI.SQLDB.Create(&Model{OwnerID: uint(ownerID)})
		return nil, errs.DoesNotExist("settings", getReq.OwnerId)
	default:
		return nil, errs.FailedToFind("settings", err)
	}

	userSettingsPB, err := GetSettingsPB(userSettingsDB)
	if err != nil {
		return nil, err
	}

	settingsPB := make(map[string]*settings.Setting, 0)

	for _, settingPB := range userSettingsPB.Settings {
		if getReq.Domain != "" && settingPB.Domain == getReq.Domain {
			settingsPB[settingPB.Key] = settingPB
			continue
		}
		settingsPB[settingPB.Key] = settingPB
	}

	return &settings.GetSettingsResponse{
		Settings: settingsPB,
	}, nil
}
