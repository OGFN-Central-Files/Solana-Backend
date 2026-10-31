package solaris

import (
	"gorm.io/gorm"
)

type Updater struct {
	gorm.Model
	Version     string `gorm:"column:version;primaryKey"`
	PublishDate string `gorm:"column:pub_date"`
	Url         string `gorm:"column:url"`
	Signature   string `gorm:"column:signature"`
	Notes       string `gorm:"column:notes"`
}

func (Updater) TableName() string {
	return "launcher_updater"
}
