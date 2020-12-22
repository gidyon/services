package account

import (
	"fmt"
	"time"

	"github.com/gidyon/micro/utils/errs"
	"github.com/gidyon/services/pkg/api/account"
	"gorm.io/gorm"
)

const accountsTable = "accounts"

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
	Nationality      string `gorm:"type:varchar(50);default:'Kenyan'"`
	ProfileURL       string `gorm:"type:varchar(256)"`
	LinkedAccounts   string `gorm:"type:varchar(256)"`
	SecurityQuestion string `gorm:"type:varchar(50)"`
	SecurityAnswer   string `gorm:"type:varchar(50)"`
	Password         string `gorm:"type:text"`
	PrimaryGroup     string `gorm:"index;type:varchar(50);not null"`
	SecondaryGroups  []byte `gorm:"type:json"`
	AccountState     string `gorm:"index;type:enum('BLOCKED','ACTIVE', 'INACTIVE');not null;default:'INACTIVE'"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt
}

// TableName is the name of the tables
func (u *Account) TableName() string {
	return accountsTable
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

// AfterFind will reset email and phone to their zero value if they equal the user id
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

// GetAccountPB converts account db model to protobuf Account message
func GetAccountPB(accountDB *Account) (*account.Account, error) {
	if accountDB == nil {
		return nil, errs.NilObject("account")
	}

	accountState := account.AccountState(account.AccountState_value[accountDB.AccountState])

	if accountDB.DeletedAt.Valid {
		accountState = account.AccountState_DELETED
	}

	accountPB := &account.Account{
		AccountId:      fmt.Sprint(accountDB.AccountID),
		ProjectId:      accountDB.ProjectID,
		Email:          accountDB.Email,
		Phone:          accountDB.Phone,
		DeviceToken:    accountDB.DeviceToken,
		Names:          accountDB.Names,
		BirthDate:      accountDB.BirthDate,
		Gender:         account.Account_Gender(account.Account_Gender_value[accountDB.Gender]),
		Nationality:    accountDB.Nationality,
		ProfileUrl:     accountDB.ProfileURL,
		LinkedAccounts: accountDB.LinkedAccounts,
		Group:          accountDB.PrimaryGroup,
		State:          accountState,
	}

	return accountPB, nil
}

// GetAccountDB converts protobuf Account message to account db model
func GetAccountDB(accountPB *account.Account) (*Account, error) {
	if accountPB == nil {
		return nil, errs.NilObject("account")
	}

	accountDB := &Account{
		ProjectID:      accountPB.ProjectId,
		Email:          accountPB.Email,
		Phone:          accountPB.Phone,
		DeviceToken:    accountPB.DeviceToken,
		Names:          accountPB.Names,
		BirthDate:      accountPB.BirthDate,
		Gender:         accountPB.Gender.String(),
		Nationality:    accountPB.Nationality,
		ProfileURL:     accountPB.ProfileUrl,
		LinkedAccounts: accountPB.LinkedAccounts,
		PrimaryGroup:   accountPB.Group,
		AccountState:   accountPB.State.String(),
	}

	return accountDB, nil
}

// GetAccountPBView returns the appropriate view
func GetAccountPBView(accountPB *account.Account, view account.AccountView) *account.Account {
	if accountPB == nil {
		return accountPB
	}
	switch view {
	case account.AccountView_SEARCH_VIEW, account.AccountView_LIST_VIEW:
		return &account.Account{
			AccountId: accountPB.AccountId,
			Email:     accountPB.Email,
			Phone:     accountPB.Phone,
			ProjectId: accountPB.ProjectId,
			Names:     accountPB.Names,
			Group:     accountPB.Group,
			State:     accountPB.State,
		}
	default:
		return accountPB
	}
}
