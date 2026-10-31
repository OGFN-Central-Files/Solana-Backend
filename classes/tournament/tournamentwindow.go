package tournament

type TournamentWindow struct {
	WindowId        string `gorm:"column:window_id"`
	Round           int    `gorm:"column:round"`
	TBD             bool   `gorm:"column:tbd"`
	CanLiveSpectate bool   `gorm:"column:can_live_spectate"`
	Start           string `gorm:"column:start"`
	End             string `gorm:"column:end"`
}
