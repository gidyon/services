package account

import (
	"fmt"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/utils/errs"
	"gorm.io/gorm"
)

const accountsTable = "accounts"

// Account contains profile information stored in the database
type Account struct {
	Email            string `gorm:"index:query_index;type:varchar(50);unique_index;not null"`
	Phone            string `gorm:"index:query_index;type:varchar(50);unique_index;not null"`
	ExternalID       string `gorm:"index:query_index;type:varchar(50);unique_index;not null"`
	DeviceToken      string `gorm:"type:varchar(256)"`
	Names            string `gorm:"type:varchar(50);not null"`
	BirthDate        string `gorm:"type:varchar(30);"`
	Gender           string `gorm:"index:query_index;type:varchar(10);default:'unknown'"`
	Nationality      string `gorm:"type:varchar(50);default:'Kenyan'"`
	ProfileURL       string `gorm:"type:varchar(256)"`
	LinkedAccounts   string `gorm:"type:varchar(256)"`
	SecurityQuestion string `gorm:"type:varchar(50)"`
	SecurityAnswer   string `gorm:"type:varchar(50)"`
	Password         string `gorm:"type:text"`
	PrimaryGroup     string `gorm:"index:query_index;type:varchar(50);not null"`
	SecondaryGroups  []byte `gorm:"type:json"`
	AccountState     string `gorm:"index:query_index;type:enum('BLOCKED','ACTIVE', 'INACTIVE');not null;default:'INACTIVE'"`
	gorm.Model
}

// TableName is the name of the tables
func (u *Account) TableName() string {
	return accountsTable
}

// AfterCreate is a callback after creating object
func (u *Account) AfterCreate(scope *gorm.Scope) error {
	accountID := fmt.Sprint(u.ID)
	var err error

	if u.Email == "" {
		err = scope.DB().Model(u).Update("email", accountID).Error
		if err != nil {
			return err
		}
	}
	if u.Phone == "" {
		err = scope.DB().Model(u).Update("phone", accountID).Error
		if err != nil {
			return err
		}
	}
	if u.ExternalID == "" {
		err = scope.DB().Model(u).Update("external_id", accountID).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// AfterFind will reset email and phone to their zero value if they equal the user id
func (u *Account) AfterFind() (err error) {
	accountID := fmt.Sprint(u.ID)
	if u.Email == accountID {
		u.Email = ""
	}
	if u.Phone == accountID {
		u.Phone = ""
	}
	if u.ExternalID == accountID {
		u.ExternalID = ""
	}
	return
}

// GetAccountPB converts account db model to protobuf Account message
func GetAccountPB(accountDB *Account) (*account.Account, error) {
	if accountDB == nil {
		return nil, errs.NilObject("account")
	}

	var state = account.AccountState(account.AccountState_value[accountDB.AccountState])
	if accountDB.DeletedAt != nil {
		state = account.AccountState_DELETED
	}

	accountPB := &account.Account{
		AccountId:      fmt.Sprint(accountDB.ID),
		ExternalId:     accountDB.ExternalID,
		Email:          accountDB.Email,
		Phone:          accountDB.Phone,
		DeviceToken:    accountDB.DeviceToken,
		Names:          accountDB.Names,
		BirthDate:      accountDB.BirthDate,
		Gender:         accountDB.Gender,
		Nationality:    accountDB.Nationality,
		ProfileUrl:     accountDB.ProfileURL,
		LinkedAccounts: accountDB.LinkedAccounts,
		Group:          accountDB.PrimaryGroup,
		State:          state,
	}

	return accountPB, nil
}

// GetAccountDB converts protobuf Account message to account db model
func GetAccountDB(accountPB *account.Account) (*Account, error) {
	if accountPB == nil {
		return nil, errs.NilObject("account")
	}

	accountDB := &Account{
		Email:          accountPB.Email,
		Phone:          accountPB.Phone,
		ExternalID:     accountPB.ExternalId,
		DeviceToken:    accountPB.DeviceToken,
		Names:          accountPB.Names,
		BirthDate:      accountPB.BirthDate,
		Gender:         accountPB.Gender,
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
			AccountId:  accountPB.AccountId,
			Email:      accountPB.Email,
			Phone:      accountPB.Phone,
			ExternalId: accountPB.ExternalId,
			Names:      accountPB.Names,
			Group:      accountPB.Group,
			State:      accountPB.State,
		}
	default:
		return accountPB
	}
}
