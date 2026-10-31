package fortnite

import (
	"gorm.io/gorm"
)

type CreatorCodes struct {
	gorm.Model
	Uuid    string `gorm:"column:uuid;primaryKey"`
	Created string `gorm:"column:created"`
	Code    string `gorm:"column:code"`
	Owner   string `gorm:"column:owner"`
	Usage   int    `gorm:"column:usage"`
}

func (CreatorCodes) TableName() string {
	return "creator_codes"
}
