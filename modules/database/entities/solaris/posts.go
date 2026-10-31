package solaris

import (
	"gorm.io/gorm"
)

type Posts struct {
	gorm.Model
	Id       int    `gorm:"column:id;primaryKey"`
	Title    string `gorm:"column:title"`
	Date     string `gorm:"column:date"`
	Content  string `gorm:"column:content"`
	Features string `gorm:"column:features"`
}

func (Posts) TableName() string {
	return "launcher_posts"
}
