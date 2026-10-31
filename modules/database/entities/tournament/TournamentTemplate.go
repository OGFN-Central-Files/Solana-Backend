package tournament

import (
	"solana/classes/tournament"

	"gorm.io/gorm"
)

type TournamentTemplate struct {
	gorm.Model
	EventId    string                         `gorm:"column:event_id;primaryKey"`
	TemplateId string                         `gorm:"column:template_id"`
	Playlist   string                         `gorm:"column:playlist"`
	MatchLimit int                            `gorm:"column:match_limit"`
	Scoring    []tournament.TournamentScoring `gorm:"column:scoring;serializer:json"`
	Payout     tournament.TournamentPayout    `gorm:"column:payout;serializer:json"`
	Season     int                            `gorm:"column:season"`
}

func (TournamentTemplate) TableName() string {
	return "tournament_templates"
}
