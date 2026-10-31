package fortnite

import (
	"gorm.io/gorm"
)

type Battlepass struct {
	gorm.Model
	RewardIndex                int    `gorm:"column:rewardIndex;primaryKey"`
	AssetPathName              string `gorm:"column:assetPathName"`
	SubPathString              string `gorm:"column:subPathString"`
	TemplateId                 string `gorm:"column:templateId"`
	Quantity                   int    `gorm:"column:quantity"`
	RewardGiftBox_GiftBoxToUse string `gorm:"column:giftBoxToUse"`
	StringAssetType            string `gorm:"column:stringAssetType"`
	StringData                 string `gorm:"column:stringData"`
	IsChaseReward              bool   `gorm:"column:isChaseReward"`
	RewardType                 string `gorm:"column:rewardType"`
}

func (Battlepass) TableName() string {
	return "battlepass"
}
