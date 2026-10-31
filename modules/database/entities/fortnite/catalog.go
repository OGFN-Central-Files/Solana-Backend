package fortnite

import (
	"gorm.io/gorm"
)

type Catalog struct {
	gorm.Model
	Created    string `gorm:"column:created"`
	Storefront string `gorm:"column:storefront"`
	OfferId    string `gorm:"column:offerId;uniqueIndex"`
	Name       string `gorm:"column:name"`
	TemplateId string `gorm:"column:templateId"`
	Data       string `gorm:"column:data"`
	Category   string `gorm:"column:category"`
}

func (Catalog) TableName() string {
	return "catalog"
}
