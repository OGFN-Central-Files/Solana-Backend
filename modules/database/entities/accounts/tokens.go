package accounts

import (
	"gorm.io/gorm"
)

type Tokens struct {
	gorm.Model
	AccountID   string   `gorm:"column:account_id;index"`
	Token       string   `gorm:"column:token;primaryKey"`
	Type        string   `gorm:"column:type"`
	Permissions []string `gorm:"column:permissions;type:text[];default:'{}'"`
}

func (Tokens) TableName() string {
	return "tokens"
}
