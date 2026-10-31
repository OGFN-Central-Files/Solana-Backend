package mcp

import (
	"encoding/json"
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	utilities "solana/utilities"
	"time"

	"github.com/gin-gonic/gin"
)

func MarkItemSeen(c *gin.Context) {
	accountId := c.Param("accountId")
	profileId := c.Query("profileId")
	db := database.Get()

	var body struct {
		ItemIds []string `json:"itemIds"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utilities.MCP.InvalidPayload().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	if profileId == "" {
		utilities.Internal.ValidationFailed().
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")

	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", accountId).First(&user).Error; err != nil {
		utilities.MCP.ProfileNotFound().
			WithMessageVar([]string{accountId}).
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var profile accounts.Profiles
	var items []accounts.Items
	profileFound := true

	if err := db.Where("account_id = ? AND profile_id = ?", accountId, profileId).First(&profile).Error; err != nil {
		profileFound = false
	} else if profileId == "athena" || profileId == "common_core" {
		db.Where("account_id = ? AND profile_id = ?", accountId, profileId).Find(&items)
	}

	var response mcp.DefaultMCPResponse

	if !profileFound {
		response = mcp.DefaultMCPResponse{
			ProfileRevision:            0,
			ProfileId:                  profileId,
			ProfileChangesBaseRevision: 0,
			ProfileCommandRevision:     0,
			ServerTime:                 now,
			ResponseVersion:            1,
			ProfileChanges: []map[string]interface{}{
				{
					"changeType": "fullProfileUpdate",
					"profile": mcp.Profile{
						Created:    now,
						Updated:    now,
						Rvn:        0,
						WipeNumber: 1,
						AccountId:  accountId,
						ProfileId:  profileId,
						Version:    "no_version",
						Items:      make(map[string]interface{}),
						Stats: mcp.Stats{
							Attributes: make(map[string]interface{}),
						},
						CommandRevision: 0,
					},
				},
			},
		}
	} else {
		profile.Revision++
		db.Save(&profile)

		profileChanges := make([]map[string]interface{}, 0)

		for _, itemId := range body.ItemIds {
			var item accounts.Items
			if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
				itemId, accountId, profileId).First(&item).Error; err != nil {
				continue
			}

			var itemValue map[string]interface{}
			if err := json.Unmarshal([]byte(item.Value), &itemValue); err != nil {
				continue
			}

			itemValue["item_seen"] = true
			newValue, _ := json.Marshal(itemValue)
			item.Value = string(newValue)
			db.Save(&item)

			profileChanges = append(profileChanges, map[string]interface{}{
				"changeType":     "itemAttrChanged",
				"itemId":         itemId,
				"attributeName":  "item_seen",
				"attributeValue": true,
			})
		}

		response = mcp.DefaultMCPResponse{
			ProfileRevision:            profile.Revision,
			ProfileId:                  profileId,
			ProfileChangesBaseRevision: profile.Revision - 1,
			ProfileCommandRevision:     profile.Revision,
			ServerTime:                 now,
			ResponseVersion:            1,
			ProfileChanges:             profileChanges,
		}
	}

	c.JSON(200, response)
}
