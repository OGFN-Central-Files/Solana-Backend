package tournament

type TournamentPayoutValue struct {
	Quantity   int    `gorm:"column:quantity"`
	RewardMode string `gorm:"column:reward_mode"`
	RewardType string `gorm:"column:reward_type"`
	Value      string `gorm:"column:value"`
}
