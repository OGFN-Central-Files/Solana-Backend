package launcher

import (
	"encoding/json"
	"net/http"
	"strings"

	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"

	"github.com/gin-gonic/gin"
)

func GetAccountLauncher(c *gin.Context) {
	db := database.Get()

	authHeader := c.GetHeader("Authorization")
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	var token accounts.Tokens
	if err := db.Where("token = ?", tokenString).First(&token).Error; err != nil {
		utilities.Authentication.AuthenticationFailed().Apply(c.Writer)
		return
	}

	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", token.AccountID).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().Apply(c.Writer)
		return
	}

	var athenaItems []accounts.Items
	if err := db.Where("account_id = ? AND profile_id = ? AND is_stat = ?", user.AccountID, "athena", true).Find(&athenaItems).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	hypeValue := 0

	var vbucks accounts.Items
	vbucksQuantity := 0
	if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ?",
		user.AccountID, "common_core", "Currency:MtxPurchased").First(&vbucks).Error; err == nil {
		vbucksQuantity = vbucks.Quantity
	}

	athenaStats := make(map[string]interface{})
	var favoriteCharacter string

	for _, item := range athenaItems {
		if item.IsStat && item.Value != "" {
			var value interface{}
			if err := json.Unmarshal([]byte(item.Value), &value); err == nil {
				athenaStats[item.TemplateID] = value

				if item.TemplateID == "favorite_character" {
					favoriteCharacter = value.(string)
				}
			}
		}
	}

	filteredStats := make(map[string]interface{})
	for key, value := range athenaStats {
		if key != "stash" &&
			key != "rested_xp_overflow" &&
			key != "mfa_reward_claimed" &&
			key != "permissions" {
			filteredStats[key] = value
		}
	}

	fav := ""
	if favoriteCharacter != "" {
		fav = strings.ToLower(favoriteCharacter)
		fav = strings.Replace(fav, "athenacharacter:", "", -1)
		fav = strings.Replace(fav, "athenabackpack:", "", -1)
		fav = strings.Replace(fav, "athenapickaxe:", "", -1)
		fav = strings.Replace(fav, "athenaglider:", "", -1)
		fav = strings.Replace(fav, "athenadance:", "", -1)
	}

	if fav == "" {
		fav = "CID_001_Athena_Commando_F_Default"
	}
	filteredStats["favorite_character"] = fav

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"accountId":      user.AccountID,
			"displayName":    user.DisplayName,
			"banned":         user.Banned,
			"email":          user.Email,
			"discordId":      user.DiscordID,
			"profilePicture": user.ProfilePicture,
			"role":           user.Role,
		},
		"hype":        hypeValue,
		"athena":      filteredStats,
		"common_core": gin.H{"VBucks": vbucksQuantity},
	})
}
