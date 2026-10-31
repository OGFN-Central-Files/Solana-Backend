package fortnite

import (
	"gorm.io/gorm"
)

type MMPayloads struct {
	gorm.Model
	AccountId string `gorm:"column:account_id"`
	SessionId string `gorm:"column:session_id"`
	Data      string `gorm:"column:data"`
	MatchId   string `gorm:"column:match_id"`
}

func (MMPayloads) TableName() string {
	return "mmpayloads"
}
