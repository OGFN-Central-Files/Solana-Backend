package tournament

type TournamentScoring struct {
	Stat  string                   `gorm:"column:stat"`
	Type  string                   `gorm:"column:type"`
	Value []TournamentScoringTiers `gorm:"column:value;serializer:json"`
}
