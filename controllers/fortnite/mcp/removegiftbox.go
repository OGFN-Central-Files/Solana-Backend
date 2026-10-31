package mcp

import (
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	utilities "solana/utilities"

	"time"

	"github.com/gin-gonic/gin"
)

func RemoveGiftBox(c *gin.Context) {
	accountId := c.Param("accountId")
	profileId := c.DefaultQuery("profileId", "")
	now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	var response mcp.DefaultMCPResponse

	if profileId == "" {
		utilities.MCP.ProfileNotFound().
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	db := database.Get()

	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", accountId).First(&user).Error; err != nil {
		utilities.MCP.ProfileNotFound().
			WithMessageVar([]string{accountId}).
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var profile accounts.Profiles
	profileFound := true

	if err := db.Where("account_id = ? AND profile_id = ?", accountId, profileId).First(&profile).Error; err != nil {
		profileFound = false
	}

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
		body := struct {
			GiftBoxItemId string `json:"giftBoxItemId"`
		}{}
		if err := c.BindJSON(&body); err != nil || body.GiftBoxItemId == "" {
			utilities.Internal.ValidationFailed().Apply(c.Writer)
			return
		}

		if err := db.Where("account_id = ? AND template_id = ?", accountId, body.GiftBoxItemId).Delete(&accounts.Items{}).Error; err != nil {
			utilities.Internal.ServerError().Apply(c.Writer)
			return
		}

		db.Exec("DELETE FROM items WHERE account_id = ? AND template_id = ? AND profile_id = ?", user.AccountID, body.GiftBoxItemId, "common_core")

		profileChanges := []map[string]interface{}{
			{
				"changeType": "itemRemoved",
				"itemId":     body.GiftBoxItemId,
			},
		}

		profile.Revision++
		db.Save(&profile)

		response = mcp.DefaultMCPResponse{
			ProfileChanges:             profileChanges,
			ProfileId:                  profileId,
			ProfileRevision:            profile.Revision,
			ProfileChangesBaseRevision: profile.Revision - 1,
			ProfileCommandRevision:     profile.Revision,
			ServerTime:                 now,
			ResponseVersion:            1,
		}
	}

	c.JSON(200, response)
}
