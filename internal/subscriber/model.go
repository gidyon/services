package subscriber

import (
	"time"

	"github.com/gidyon/services/pkg/api/account"

	"github.com/gidyon/services/pkg/api/subscriber"
)

const subscribersTable = "subscribers"

// Subscriber is model for subscribers
type Subscriber struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	UserID    string `gorm:"index;type:varchar(20);not null"`
	Channel   string `gorm:"index;type:varchar(50);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// TableName returns the name of the table
func (*Subscriber) TableName() string {
	return subscribersTable
}

// GetSubscriberPB creates a proto message from subscriber model
func GetSubscriberPB(userPB *account.Account, channels []string) (*subscriber.Subscriber, error) {
	subscriberPB := &subscriber.Subscriber{
		SubscriberId: userPB.AccountId,
		Email:        userPB.Email,
		Phone:        userPB.Phone,
		DeviceToken:  userPB.DeviceToken,
		Channels:     channels,
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
