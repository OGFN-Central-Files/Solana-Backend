package tournament

type TournamentPayout struct {
	Value     []TournamentPayoutRank `gorm:"column:value;serializer:json"`
	Rolling   bool                   `gorm:"column:rolling"`
	ScoreId   *string                `gorm:"column:score_id"`
	ScoreType string                 `gorm:"column:score_type"`
}
