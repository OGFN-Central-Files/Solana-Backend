package accounts

import (
	"gorm.io/gorm"
)

type Stats struct {
	gorm.Model
	AccountID     string `gorm:"column:account_id;primaryKey"`
	MatchesPlayed int    `gorm:"column:matches_played;default:0"`
	Solos         string `gorm:"column:solos"`
	Duos          string `gorm:"column:duos"`
	Squads        string `gorm:"column:squads"`
}

func (Stats) TableName() string {
	return "stats"
}
