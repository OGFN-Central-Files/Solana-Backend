package accounts

import (
	"gorm.io/gorm"
)

type Items struct {
	gorm.Model
	UUID       string `gorm:"column:uuid;primaryKey"`
	AccountID  string `gorm:"column:account_id;index"`
	ProfileID  string `gorm:"column:profile_id;index"`
	TemplateID string `gorm:"column:template_id"`
	Value      string `gorm:"column:value"`
	Quantity   int    `gorm:"column:quantity"`
	IsStat     bool   `gorm:"column:is_stat;default:false"`
}

func (Items) TableName() string {
	return "items"
}
