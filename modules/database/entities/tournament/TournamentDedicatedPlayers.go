package tournament

import (
	"gorm.io/gorm"
)

type TournamentDedicatedPlayers struct {
	gorm.Model
	AccountId      string `gorm:"column:account_id"`
	TeamId         string `gorm:"column:team_id"`
	TeamAccountIds string `gorm:"column:team_account_ids"`
	SessionHistory string `gorm:"column:session_history"`
	EventId        string `gorm:"column:event_id"`
	Window         int    `gorm:"column:window"`
	PointsEarned   int    `gorm:"column:points_earned"`
	Season         int    `gorm:"column:season"`
}

func (TournamentDedicatedPlayers) TableName() string {
	return "tournament_dedicated_players"
}
