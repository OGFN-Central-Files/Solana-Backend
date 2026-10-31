package contentpages

import (
	"gorm.io/gorm"
)

type TournamentInfo struct {
	gorm.Model
	TitleColor           string `gorm:"column:title_color"`
	LoadingScreenImage   string `gorm:"column:loading_screen_image"`
	BackgroundTextColor  string `gorm:"column:background_text_color"`
	BackgroundRightColor string `gorm:"column:background_right_color"`
	PosterBackImage      string `gorm:"column:poster_back_image"`
	Type                 string `gorm:"column:_type"`
	PinScoreRequirement  string `gorm:"column:pin_score_requirement"`
	PinEarnedText        string `gorm:"column:pin_earned_text"`
	TournamentDisplayId  string `gorm:"column:tournament_display_id"`
	HighlightColor       string `gorm:"column:highlight_color"`
	ScheduleInfo         string `gorm:"column:schedule_info"`
	PrimaryColor         string `gorm:"column:primary_color"`
	FlavorDescription    string `gorm:"column:flavor_description"`
	PosterFrontImage     string `gorm:"column:poster_front_image"`
	ShortFormatTitle     string `gorm:"column:short_format_title"`
	TitleLine2           string `gorm:"column:title_line_2"`
	TitleLine1           string `gorm:"column:title_line_1"`
	ShadowColor          string `gorm:"column:shadow_color"`
	DetailsDescription   string `gorm:"column:details_description"`
	BackgroundLeftColor  string `gorm:"column:background_left_color"`
	LongFormatTitle      string `gorm:"column:long_format_title"`
	PosterFadeColor      string `gorm:"column:poster_fade_color"`
	SecondaryColor       string `gorm:"column:secondary_color"`
	PlaylistTileImage    string `gorm:"column:playlist_tile_image"`
	BaseColor            string `gorm:"column:base_color"`
}

func (TournamentInfo) TableName() string {
	return "tournament_info"
}
