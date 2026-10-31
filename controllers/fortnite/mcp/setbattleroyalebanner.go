package mcp

import (
	"net/http"
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"
	"time"

	"github.com/gin-gonic/gin"
)

func SetBattleRoyaleBanner(c *gin.Context) {
	accountId := c.Param("accountId")
	profileId := c.Query("profileId")
	db := database.Get()

	if profileId == "" {
		utilities.Internal.ValidationFailed().
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", accountId).First(&user).Error; err != nil {
		utilities.MCP.ProfileNotFound().
			WithMessageVar([]string{accountId}).
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var body struct {
		HomebaseBannerColorId string `json:"homebaseBannerColorId"`
		HomebaseBannerIconId  string `json:"homebaseBannerIconId"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utilities.MCP.InvalidPayload().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	var response mcp.DefaultMCPResponse

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
		c.JSON(200, response)
		return
	} else {
		var bannerIcon accounts.Items
		var bannerColor accounts.Items

		db.Where("account_id = ? AND template_id = ?", accountId, "banner_icon").First(&bannerIcon)
		db.Where("account_id = ? AND template_id = ?", accountId, "banner_color").First(&bannerColor)

		if bannerIcon.ID == 0 || bannerColor.ID == 0 {
			utilities.MCP.ItemNotFound().
				WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		bannerIcon.Value = "\"" + body.HomebaseBannerIconId + "\""
		bannerColor.Value = "\"" + body.HomebaseBannerColorId + "\""

		db.Save(&bannerIcon)
		db.Save(&bannerColor)

		profile.Revision++
		db.Save(&profile)

		profileChanges := []map[string]interface{}{
			{
				"changeType": "statModified",
				"name":       "banner_color",
				"value":      body.HomebaseBannerColorId,
			},
			{
				"changeType": "statModified",
				"name":       "banner_icon",
				"value":      body.HomebaseBannerIconId,
			},
		}

		response = mcp.DefaultMCPResponse{
			ProfileId:                  profileId,
			ProfileRevision:            profile.Revision,
			ProfileChangesBaseRevision: profile.Revision - 1,
			ProfileCommandRevision:     profile.Revision,
			ResponseVersion:            1,
			ServerTime:                 now,
			ProfileChanges:             profileChanges,
		}
	}

	c.JSON(http.StatusOK, response)
}
