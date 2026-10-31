package accounts

import (
	"gorm.io/gorm"
)

type Quests struct {
	gorm.Model
	UUID       string `gorm:"column:uuid;primaryKey"`
	AccountID  string `gorm:"column:account_id;index"`
	ProfileID  string `gorm:"column:profile_id;index"`
	TemplateID string `gorm:"column:template_id"`
	Value      string `gorm:"column:value"`
	Daily      bool   `gorm:"column:daily;default:false"`
	Season     int    `gorm:"column:season;default:0"`
}

func (Quests) TableName() string {
	return "quests"
}
