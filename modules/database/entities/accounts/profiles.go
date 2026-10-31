package accounts

import (
	"gorm.io/gorm"
)

type Profiles struct {
	gorm.Model
	AccountID string `gorm:"column:account_id;index"`
	ProfileID string `gorm:"column:profile_id;index"`
	Revision  int    `gorm:"column:revision;default:0"`
}

func (Profiles) TableName() string {
	return "profiles"
}
