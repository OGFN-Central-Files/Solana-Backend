package tournament

import (
	"solana/classes/tournament"

	"gorm.io/gorm"
)

type TournamentEvent struct {
	gorm.Model
	EventId       string                        `gorm:"column:event_id;primaryKey"`
	DisplayId     string                        `gorm:"column:display_id"`
	Begins        string                        `gorm:"column:begins"`
	Ends          string                        `gorm:"column:ends"`
	Playlist      string                        `gorm:"column:playlist"`
	MinAccountLvl int                           `gorm:"column:min_account_lvl"`
	Windows       []tournament.TournamentWindow `gorm:"column:windows;serializer:json"`
	Season        int                           `gorm:"column:season"`
}

func (TournamentEvent) TableName() string {
	return "tournament_events"
}
