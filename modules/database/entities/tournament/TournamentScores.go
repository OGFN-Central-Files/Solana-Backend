package tournament

import (
	"gorm.io/gorm"
)

type TournamentScores struct {
	gorm.Model
	AccountId string `gorm:"column:account_id"`
	Type      string `gorm:"column:type"`
	Value     int    `gorm:"column:value"`
	Season    int    `gorm:"column:season"`
}

func (TournamentScores) TableName() string {
	return "tournament_scores"
}
