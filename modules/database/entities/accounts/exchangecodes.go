package accounts

import (
	"gorm.io/gorm"
)

type ExchangeCodes struct {
	gorm.Model
	AccountId string `gorm:"column:account_id"`
	Code      string `gorm:"column:code;primaryKey"`
	Type      string `gorm:"column:type"`
	Created   string `gorm:"column:created"`
}

func (ExchangeCodes) TableName() string {
	return "exchange_codes"
}
