package solaris

import (
	"time"

	"gorm.io/gorm"
)

type Banners struct {
	gorm.Model
	BannerId       string     `gorm:"column:banner_id;primaryKey"`
	Title          string     `gorm:"column:title;size:100;not null"`
	Subtitle       string     `gorm:"column:subtitle;size:100"`
	Description    string     `gorm:"column:description;size:500;not null"`
	BackgroundType string     `gorm:"column:background_type;size:10;not null;default:image"`
	BackgroundUrl  string     `gorm:"column:background_url;size:500;not null"`
	OverlayColor   string     `gorm:"column:overlay_color;size:10"`
	OverlayOpacity float64    `gorm:"column:overlay_opacity"`
	ThemePrimary   string     `gorm:"column:theme_primary;size:10;not null"`
	ThemeSecondary string     `gorm:"column:theme_secondary;size:10;not null"`
	ThemeText      string     `gorm:"column:theme_text;size:10;not null"`
	ThemeAccent    string     `gorm:"column:theme_accent;size:10"`
	CtaText        string     `gorm:"column:cta_text;size:50;not null"`
	CtaUrl         string     `gorm:"column:cta_url;size:500"`
	CtaType        string     `gorm:"column:cta_type;size:20;not null;default:internal"`
	IsPremium      bool       `gorm:"column:is_premium"`
	Priority       int        `gorm:"column:priority"`
	ExpireAt       *time.Time `gorm:"column:expire_at"`
	Tags           string     `gorm:"column:tags;size:255"`
	Enabled        bool       `gorm:"column:enabled;default:true"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"`
}

func (Banners) TableName() string {
	return "news_banners"
}
