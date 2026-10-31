package accounts

import (
	"gorm.io/gorm"
)

type Friends struct {
	gorm.Model
	Created   string `gorm:"column:created"`
	AccountId string `gorm:"column:account_id;size:255"`
	FriendId  string `gorm:"column:friend_id;size:255"`
	Alias     string `gorm:"column:alias"`
	Direction string `gorm:"column:direction"`
	Status    string `gorm:"column:status"`
}

func (Friends) TableName() string {
	return "friends"
}
