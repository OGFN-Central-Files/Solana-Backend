package accounts

import (
	"gorm.io/gorm"
)

type Secrets struct {
	gorm.Model
	Secret    string `gorm:"column:secret;uniqueIndex"`
	AccountID string `gorm:"column:account_id"`
}

func (Secrets) TableName() string {
	return "secrets"
}
