package fortnite

import (
	"gorm.io/gorm"
)

type MatchStats struct {
	gorm.Model
	AccountId        string `gorm:"column:account_id"`
	Kills            int    `gorm:"column:kills"`
	Team             string `gorm:"column:team;type:jsonb"`
	Xp               int    `gorm:"column:xp"`
	ChallengeUpdates string `gorm:"column:challenge_updates"`
	Position         int    `gorm:"column:position"`
	Playlist         string `gorm:"column:playlist"`
}

func (MatchStats) TableName() string {
	return "match_stats"
}
