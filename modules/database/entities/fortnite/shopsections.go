package fortnite

import (
	"time"

	"gorm.io/gorm"
)

type ShopSections struct {
	gorm.Model
	Section string    `gorm:"column:section"`
	Date    time.Time `gorm:"column:date"`
}

func (ShopSections) TableName() string {
	return "shopsections"
}
