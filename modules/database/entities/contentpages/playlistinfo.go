package contentpages

import (
	"gorm.io/gorm"
)

type PlaylistInfo struct {
	gorm.Model
	Type          string `gorm:"column:_type"`
	Image         string `gorm:"column:image"`
	PlaylistName  string `gorm:"column:playlist_name"`
	Hidden        bool   `gorm:"column:hidden"`
	SpecialBorder string `gorm:"column:special_border"`
}

func (PlaylistInfo) TableName() string {
	return "playlist_info"
}
