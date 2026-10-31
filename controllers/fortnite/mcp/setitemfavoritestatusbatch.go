package mcp

import (
	"encoding/json"
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"
	"time"

	"github.com/gin-gonic/gin"
)

// first thing i did that didnt use json looked at the other files to figure this out I'm still learning
func SetItemFavoriteStatusBatch(c *gin.Context) {
	accountId := c.Param("accountId")
	profileId := c.Query("profileId")
	db := database.Get()

	var body struct {
		ItemIds       []string `json:"itemIds"`
		ItemFavStatus []bool   `json:"itemFavStatus"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utilities.MCP.InvalidPayload().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	if profileId == "" {
		utilities.MCP.ProfileNotFound().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")

	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", accountId).First(&user).Error; err != nil {
		utilities.MCP.ProfileNotFound().
			WithMessageVar([]string{accountId}).
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	var profile accounts.Profiles
	if err := db.Where("account_id = ? AND profile_id = ?", accountId, profileId).First(&profile).Error; err != nil {
		utilities.MCP.ProfileNotFound().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	profileChanges := make([]map[string]interface{}, 0)

	for i := 0; i < len(body.ItemIds) && i < len(body.ItemFavStatus); i++ {
		itemId := body.ItemIds[i]
		isFavorite := body.ItemFavStatus[i]

		var item accounts.Items
		if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
			itemId, accountId, profileId).First(&item).Error; err != nil {
			continue
		}

		var itemValue map[string]interface{}
		if err := json.Unmarshal([]byte(item.Value), &itemValue); err != nil {
			continue
		}

		itemValue["favorite"] = isFavorite
		newValue, _ := json.Marshal(itemValue)
		item.Value = string(newValue)
		db.Save(&item)

		profileChanges = append(profileChanges, map[string]interface{}{
			"changeType":     "itemAttrChanged",
			"itemId":         itemId,
			"attributeName":  "favorite",
			"attributeValue": isFavorite,
		})
	}

	profile.Revision++
	db.Save(&profile)

	response := mcp.DefaultMCPResponse{
		ProfileRevision:            profile.Revision,
		ProfileId:                  profileId,
		ProfileChangesBaseRevision: profile.Revision - 1,
		ProfileCommandRevision:     profile.Revision,
		ServerTime:                 now,
		ResponseVersion:            1,
		ProfileChanges:             profileChanges,
	}

	c.JSON(200, response)
}
