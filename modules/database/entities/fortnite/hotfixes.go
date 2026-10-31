package fortnite

import (
	"gorm.io/gorm"
)

type Hotfixes struct {
	gorm.Model
	Name    string `gorm:"column:name;uniqueIndex"`
	Value   string `gorm:"column:value"`
	Enabled bool   `gorm:"column:enabled"`
}

func (Hotfixes) TableName() string {
	return "hotfixes"
}
