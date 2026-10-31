package tournament

import (
	"gorm.io/gorm"
)

type TournamentHistory struct {
	gorm.Model
	AccountId          string `gorm:"column:account_id"`
	EventId            string `gorm:"column:event_id"`
	TeamId             string `gorm:"column:team_id"`
	TeamElimsStatIndex string `gorm:"column:team_elims_stat_index"`
	PlacementStatIndex string `gorm:"column:placement_stat_index"`
	VictoryStatIndex   string `gorm:"column:victory_stat_index"`
	Window             int    `gorm:"column:window"`
	Season             int    `gorm:"column:season;default:0"`
}

func (TournamentHistory) TableName() string {
	return "tournament_history"
}
