package settings

import (
	"encoding/json"
	"time"

	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/settings"
	"gorm.io/gorm"
)

const settingsTable = "settings"

// Model for settings
type Model struct {
	OwnerID   uint   `gorm:"primarykey"`
	Settings  []byte `gorm:"type:json"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName is table that stores user settings
func (*Model) TableName() string {
	return settingsTable
}

// GetSettingsPB converts settings model to protobuf message
func GetSettingsPB(settingsDB *Model) (*settings.Settings, error) {
	if settingsDB == nil {
		return nil, errs.NilObject("settings")
	}

	settingsPB := &settings.Settings{
		Settings: make(map[string]*settings.Setting),
	}

	err := json.Unmarshal(settingsDB.Settings, &settingsPB.Settings)
	if err != nil {
		return nil, errs.FromJSONUnMarshal(err, "settings")
	}

	return settingsPB, nil
}

// GetSettingsDB converts settings model to protobuf settings
func GetSettingsDB(settingsPB *settings.Settings) (*Model, error) {
	if settingsPB == nil {
		return nil, errs.NilObject("settings")
	}

	bs, err := json.Marshal(settingsPB.Settings)
	if err != nil {
		return nil, errs.FromJSONMarshal(err, "settings")
	}

	settingsDB := &Model{
		Settings: bs,
	}

	return settingsDB, nil
}
