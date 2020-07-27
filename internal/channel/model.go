package channel

import (
	"fmt"

	"github.com/gidyon/services/pkg/api/channel"
	"github.com/jinzhu/gorm"
)

const channelsTable = "channels"

// Channel is a bulk channel
type Channel struct {
	Title       string `gorm:"type:varchar(50);not null"`
	Description string `gorm:"type:text;not null"`
	OwnerID     string `gorm:"type:varchar(50);not null"`
	Subscribers int32  `gorm:"type:int(10);not null"`
	gorm.Model
}

// TableName returns the table name of the channel
func (*Channel) TableName() string {
	return channelsTable
}

// GetChannelPB gets the proto message equivalence from a channel model
func GetChannelPB(channelDB *Channel) (*channel.Channel, error) {
	channelPB := &channel.Channel{
		Id:                fmt.Sprint(channelDB.ID),
		Title:             channelDB.Title,
		Description:       channelDB.Description,
		OwnerId:           channelDB.OwnerID,
		CreateTimeSeconds: int32(channelDB.CreatedAt.Unix()),
		Subscribers:       channelDB.Subscribers,
	}
	return channelPB, nil
}

// GetChannelDB gets the database model of a channel proto message
func GetChannelDB(channelPB *channel.Channel) (*Channel, error) {
	channelDB := &Channel{
		Title:       channelPB.Title,
		Description: channelPB.Description,
		OwnerID:     channelPB.OwnerId,
		Subscribers: channelPB.Subscribers,
	}

	return channelDB, nil
}
