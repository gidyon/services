package account

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/account"
	"gorm.io/gorm"
)

const defaultAccountsTable = "accounts"

var accountsTable = ""

// Account contains profile information stored in the database
type Account struct {
	AccountID        uint   `gorm:"primaryKey;autoIncrement"`
	ProjectID        string `gorm:"index;type:varchar(50);not null"`
	Email            string `gorm:"index;type:varchar(50);not null"`
	Phone            string `gorm:"index;type:varchar(50);not null"`
	DeviceToken      string `gorm:"type:varchar(256)"`
	Names            string `gorm:"type:varchar(50);not null"`
	BirthDate        string `gorm:"type:varchar(30);"`
	Gender           string `gorm:"index;type:enum('GENDER_UNSPECIFIED', 'MALE', 'FEMALE');default:'GENDER_UNSPECIFIED';not null"`
	IDNumber         string `gorm:"index;type:varchar(15)"`
	Profession       string `gorm:"type:varchar(50)"`
	Residence        string `gorm:"type:varchar(100)"`
	Nationality      string `gorm:"type:varchar(50);default:'Kenyan'"`
	ProfileURL       string `gorm:"type:varchar(256)"`
	ParentId         string `gorm:"type:varchar(50)"`
	LinkedAccounts   string `gorm:"type:varchar(256)"`
	SecurityQuestion string `gorm:"type:varchar(50)"`
	SecurityAnswer   string `gorm:"type:varchar(50)"`
	Password         string `gorm:"type:text"`
	PrimaryGroup     string `gorm:"index;type:varchar(50);not null"`
	SecondaryGroups  []byte `gorm:"type:json"`
	AccountState     string `gorm:"index;type:enum('BLOCKED','ACTIVE', 'INACTIVE');not null;default:'INACTIVE'"`
	LastLogin        *time.Time
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt
}

// TableName is the name of the tables
func (u *Account) TableName() string {
	if accountsTable != "" {
		return accountsTable
	}
	return defaultAccountsTable
}

// AfterCreate is a callback after creating object
func (u *Account) AfterCreate(tx *gorm.DB) error {
	accountID := fmt.Sprint(u.AccountID)
	var err error

	if u.Email == "" {
		err = tx.Model(u).Update("email", accountID).Error
		if err != nil {
			return err
		}
	}
	if u.Phone == "" {
		err = tx.Model(u).Update("phone", accountID).Error
		if err != nil {
			return err
		}
	}
	if u.ProjectID == "" {
		err = tx.Model(u).Update("project_id", accountID).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// AfterFind will reset email and phone to their zero value if they equal the accoint id
func (u *Account) AfterFind(tx *gorm.DB) (err error) {
	accountID := fmt.Sprint(u.AccountID)
	if u.Email == accountID {
		u.Email = ""
	}
	if u.Phone == accountID {
		u.Phone = ""
	}
	if u.ProjectID == accountID {
		u.ProjectID = ""
	}
	return
}

// AccountProto converts account db model to protobuf Account message
func AccountProto(db *Account) (*account.Account, error) {
	if db == nil {
		return nil, errs.NilObject("account")
	}

	accountState := account.AccountState(account.AccountState_value[db.AccountState])

	if db.DeletedAt.Valid {
		accountState = account.AccountState_DELETED
	}

	// Secondary groups
	secondaryGroups := make([]string, 0)
	if len(db.SecondaryGroups) != 0 {
		err := json.Unmarshal(db.SecondaryGroups, &secondaryGroups)
		if err != nil {
			return nil, errs.WrapErrorWithMsg(err, "failed to json unmarshal")
		}
	}

	pb := &account.Account{
		AccountId:       fmt.Sprint(db.AccountID),
		ProjectId:       db.ProjectID,
		Email:           db.Email,
		Phone:           db.Phone,
		DeviceToken:     db.DeviceToken,
		Names:           db.Names,
		BirthDate:       db.BirthDate,
		Gender:          account.Account_Gender(account.Account_Gender_value[db.Gender]),
		Nationality:     db.Nationality,
		Residence:       db.Residence,
		Profession:      db.Profession,
		IdNumber:        db.IDNumber,
		ProfileUrl:      db.ProfileURL,
		LinkedAccounts:  db.LinkedAccounts,
		CreatedAt:       db.CreatedAt.Format(time.RFC3339),
		Group:           db.PrimaryGroup,
		State:           accountState,
		SecondaryGroups: secondaryGroups,
	}

	if db.LastLogin != nil {
		pb.LastLogin = db.LastLogin.Format(time.RFC3339)
	}

	return pb, nil
}

// AccountModel converts protobuf Account message to account db model
func AccountModel(pb *account.Account) (*Account, error) {
	if pb == nil {
		return nil, errs.NilObject("account")
	}

	db := &Account{
		ProjectID:      pb.ProjectId,
		Email:          pb.Email,
		Phone:          pb.Phone,
		DeviceToken:    pb.DeviceToken,
		Names:          pb.Names,
		BirthDate:      pb.BirthDate,
		Gender:         pb.Gender.String(),
		Nationality:    pb.Nationality,
		Residence:      pb.Residence,
		IDNumber:       pb.IdNumber,
		Profession:     pb.Profession,
		ProfileURL:     pb.ProfileUrl,
		LinkedAccounts: pb.LinkedAccounts,
		PrimaryGroup:   pb.Group,
		AccountState:   pb.State.String(),
	}

	if len(pb.SecondaryGroups) > 0 {
		bs, err := json.Marshal(pb.SecondaryGroups)
		if err != nil {
			return nil, errs.FromJSONMarshal(err, "secondary group")
		}
		db.SecondaryGroups = bs
	}

	return db, nil
}

// AccountProtoView returns the appropriate view
func AccountProtoView(pb *account.Account, view account.AccountView) *account.Account {
	if pb == nil {
		return pb
	}
	switch view {
	case account.AccountView_SEARCH_VIEW, account.AccountView_LIST_VIEW:
		return &account.Account{
			AccountId: pb.AccountId,
			Email:     pb.Email,
			Phone:     pb.Phone,
			ProjectId: pb.ProjectId,
			Names:     pb.Names,
			Group:     pb.Group,
			State:     pb.State,
		}
	default:
		return pb
	}
}
