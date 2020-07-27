package subscriber

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gidyon/services/pkg/api/account"
	"github.com/gidyon/services/pkg/utils/errs"

	"github.com/gidyon/services/pkg/api/subscriber"
)

const subscribersTable = "subscribers"

// Subscriber is model for subscribers
type Subscriber struct {
	ID        uint   `gorm:"primary_key"`
	Channels  []byte `gorm:"type:json"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// TableName returns the name of the table
func (*Subscriber) TableName() string {
	return subscribersTable
}

// GetSubscriberPB creates a proto message from subscriber model
func GetSubscriberPB(subscriberDB *Subscriber, userPB *account.Account) (*subscriber.Subscriber, error) {
	subscriberPB := &subscriber.Subscriber{
		SubscriberId: fmt.Sprint(subscriberDB.ID),
		Email:        userPB.Email,
		Phone:        userPB.Phone,
		DeviceToken:  userPB.DeviceToken,
		Channels:     []*subscriber.Channel{},
	}

	// safe json unmarshal
	if len(subscriberDB.Channels) > 0 {
		err := json.Unmarshal(subscriberDB.Channels, &subscriberPB.Channels)
		if err != nil {
			return nil, errs.FromJSONUnMarshal(err, "channels")
		}
	}

	return subscriberPB, nil
}

// GetSubscriberDB creates a subscriber model from a proto message
// func GetSubscriberDB(subscriberPB *subscriber.Subscriber) (*Subscriber, error) {
// 	ID, err := strconv.Atoi(subscriberPB.SubscriberId)
// 	if err != nil {
// 		return nil, errs.WrapErrorWithCodeAndMsg(codes.InvalidArgument, err, "failed to retrieve ID")
// 	}

// 	subscriberDB := &Subscriber{
// 		ID:       uint(ID),
// 		Channels: make([]byte, 0),
// 	}

// 	// safe json marshal
// 	if len(subscriberPB.Channels) != 0 {
// 		data, err := json.Marshal(subscriberPB.Channels)
// 		if err != nil {
// 			return nil, errs.FromJSONMarshal(err, "channels")
// 		}
// 		subscriberDB.Channels = data
// 	}

// 	return subscriberDB, nil
// }
