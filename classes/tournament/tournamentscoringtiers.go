package tournament

type TournamentScoringTiers struct {
	Value    int  `gorm:"column:value"`
	Points   int  `gorm:"column:points"`
	Multiply bool `gorm:"column:multiply"`
}
