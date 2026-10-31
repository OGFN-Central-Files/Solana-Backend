package dsmcp

import (
	"encoding/json"
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"
	"time"

	"github.com/gin-gonic/gin"
)

func SetCreativePlotMetadata(c *gin.Context) {
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
		Introduction   string `json:"introduction"`
		Locale         string `json:"locale"`
		PlotItemId     string `json:"plotItemId"`
		Tagline        string `json:"tagline"`
		Title          string `json:"title"`
		YoutubeVideoId string `json:"youtubeVideoId"`
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
	var items []accounts.Items
	profileFound := true

	if err := db.Where("account_id = ? AND profile_id = ?", accountId, profileId).First(&profile).Error; err != nil {
		profileFound = false
	} else if profileId == "athena" || profileId == "common_core" {
		db.Where("account_id = ? AND profile_id = ?", accountId, profileId).Find(&items)
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
	}

	profile.Revision++
	db.Save(&profile)

	db.Where("account_id = ? AND profile_id = ?", accountId, profileId).Find(&items)

	profileData := mcp.Profile{
		Created:         now,
		Updated:         now,
		Rvn:             profile.Revision,
		WipeNumber:      1,
		AccountId:       accountId,
		ProfileId:       profileId,
		Version:         "no_version",
		CommandRevision: profile.Revision,
		Items:           make(map[string]interface{}),
		Stats: mcp.Stats{
			Attributes: make(map[string]interface{}),
		},
	}

	for _, item := range items {
		if !item.IsStat {
			var valueMap interface{}
			err := json.Unmarshal([]byte(item.Value), &valueMap)
			if err != nil {
				continue
			}

			itemData := mcp.BaseItem{
				Attributes: valueMap,
				TemplateId: item.TemplateID,
				Quantity:   1,
			}

			profileData.Items[item.UUID] = itemData

			if item.UUID == body.PlotItemId {
				var value map[string]interface{}
				if err := json.Unmarshal([]byte(item.Value), &value); err != nil {
					continue
				}

				value["title"] = body.Title
				value["locale"] = body.Locale

				updatedValue, err := json.Marshal(value)
				if err != nil {
					continue
				}

				item.Value = string(updatedValue)
				db.Save(&item)
			}
		} else {
			var statValue interface{}
			if err := json.Unmarshal([]byte(item.Value), &statValue); err == nil {
				profileData.Stats.Attributes[item.TemplateID] = statValue
			}
		}
	}

	response = mcp.DefaultMCPResponse{
		ProfileRevision:            profile.Revision,
		ProfileId:                  profileId,
		ProfileChangesBaseRevision: profile.Revision - 1,
		ProfileCommandRevision:     profile.Revision,
		ServerTime:                 now,
		ResponseVersion:            1,
		ProfileChanges: []map[string]interface{}{
			{
				"changeType": "fullProfileUpdate",
				"profile":    profileData,
			},
		},
	}

	c.JSON(200, response)
}
