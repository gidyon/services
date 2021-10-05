package sms

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/messaging/sms"
	"gorm.io/gorm"
)

type SenderCredential struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	ProjectID  string    `gorm:"index;unique;type:varchar(50);not null"`
	Credential []byte    `gorm:"type:json"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt
}

func (*SenderCredential) TableName() string {
	tableName := os.Getenv("SENDERID_CREDENTIALS_TABLE")
	if tableName != "" {
		return tableName
	}
	return "sms_senderid_credentials"
}

func SenderCredentialProto(db *SenderCredential) (*sms.SenderCredential, error) {
	pb := &sms.SenderCredential{
		CredentialId: fmt.Sprint(db.ID),
		ProjectId:    db.ProjectID,
		Auth:         &sms.SMSAuth{},
	}

	if len(db.Credential) != 0 {
		err := json.Unmarshal(db.Credential, pb.Auth)
		if err != nil {
			return nil, errs.FromJSONUnMarshal(err, "credential")
		}

	}

	return pb, nil
}

func SenderCredentialModel(pb *sms.SenderCredential) (*SenderCredential, error) {
	db := &SenderCredential{
		ProjectID:  pb.ProjectId,
		Credential: []byte{},
	}

	if pb.Auth != nil {
		bs, err := json.Marshal(pb.Auth)
		if err != nil {
			return nil, errs.FromJSONMarshal(err, "auth")
		}
		db.Credential = bs
	}

	return db, nil
}
