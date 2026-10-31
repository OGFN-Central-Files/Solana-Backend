package tournament

type TournamentPayoutRank struct {
	Threshold int                     `gorm:"column:threshold"`
	Value     []TournamentPayoutValue `gorm:"column:value;serializer:json"`
}
