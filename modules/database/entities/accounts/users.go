package accounts

import (
	"time"

	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	Email                   string              `gorm:"column:email;uniqueIndex" validate:"email"`
	Password                string              `gorm:"column:password"`
	AccountID               string              `gorm:"column:accountId"`
	DiscordID               string              `gorm:"column:discordId"`
	Banned                  bool                `gorm:"column:banned;default:false"`
	HasFL                   bool                `gorm:"column:hasFL;default:false"`
	Created                 time.Time           `gorm:"column:created"`
	DisplayName             string              `gorm:"column:displayName"`
	Role                    string              `gorm:"column:role"`
	BanHistory              []map[string]string `gorm:"column:ban_history;type:jsonb"`
	IsServer                bool                `gorm:"column:is_server;default:false"`
	PartyIndex              int                 `gorm:"column:party_index;default:0"`
	LastLoginTime           string              `gorm:"column:lastLoginTime"`
	LastLoginIP             string              `gorm:"column:lastLoginIP"`
	HWID                    string              `gorm:"column:hwid"`
	ProfilePicture          string              `gorm:"column:profilePicture"`
	DisplayNameChanges      int                 `gorm:"column:displayNameChanges;default:0"`
	MatchmakingBannedUntil  string              `gorm:"column:matchmakingBannedUntil"`
	MatchmakingBannedSince  string              `gorm:"column:matchmakingBannedSince"`
	MatchmakingBannedReason string              `gorm:"column:matchmakingBannedReason"`
	HasGold                 bool                `gorm:"column:hasGold;default:false"`
	HasPlat                 bool                `gorm:"column:hasPlat;default:false"`
	HasDiamond              bool                `gorm:"column:hasDiamond;default:false"`
}

func (Users) TableName() string {
	return "users"
}
