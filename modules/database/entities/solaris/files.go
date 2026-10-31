package solaris

import (
	"gorm.io/gorm"
)

type Files struct {
	gorm.Model
	Url  string `gorm:"column:url;primaryKey"`
	Name string `gorm:"column:name"`
}

func (Files) TableName() string {
	return "files"
}
