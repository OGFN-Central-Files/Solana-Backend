package fortnite

import (
	"time"

	"gorm.io/gorm"
)

type PastShops struct {
	gorm.Model
	Date time.Time `gorm:"column:date"`
}

func (PastShops) TableName() string {
	return "pastshops"
}

func GetLastProcessedShop(db *gorm.DB) time.Time {
	var lastShop PastShops
	result := db.Order("date DESC").First(&lastShop)
	if result.Error != nil {
		return time.Time{}
	}
	return lastShop.Date
}

func SaveProcessedShop(db *gorm.DB, date time.Time) {
	db.Create(&PastShops{Date: date})
}

func IsShopProcessed(db *gorm.DB, date time.Time) bool {
	var shop PastShops
	currentDay := time.Now().Format("2006-01-02")

	result := db.Where("date = ? AND DATE(created_at) = ?", date.Format("2006-01-02"), currentDay).First(&shop)
	return result.RowsAffected > 0
}
