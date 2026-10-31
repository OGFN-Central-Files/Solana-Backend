package tournament

import (
	"gorm.io/gorm"
)

type TournamentTokens struct {
	gorm.Model
	AccountId string `gorm:"column:account_id"`
	Token     string `gorm:"column:token"`
	Season    int    `gorm:"column:season"`
}

func (TournamentTokens) TableName() string {
	return "tournament_tokens"
}
