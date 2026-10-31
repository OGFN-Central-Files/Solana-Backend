package mcp

import (
	"encoding/json"
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	utilities "solana/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func SetMtxPlatform(c *gin.Context) {
	accountId := c.Param("accountId")
	profileId := c.DefaultQuery("profileId", "")
	if profileId == "" {
		utilities.Internal.ValidationFailed().
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var body map[string]interface{}
	if err := c.BindJSON(&body); err != nil {
		utilities.Internal.ValidationFailed().
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	newPlatform, exists := body["newPlatform"].(string)
	if !exists || newPlatform == "" {
		utilities.Internal.ValidationFailed().
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	db := database.Get()

	now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	var response mcp.DefaultMCPResponse

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
	} else if profileFound {
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
	} else {
		profileData := mcp.Profile{
			Created:         now,
			Updated:         now,
			Rvn:             profile.Revision,
			WipeNumber:      1,
			AccountId:       accountId,
			ProfileId:       profile.ProfileID,
			Version:         "no_version",
			CommandRevision: profile.Revision,
			Items:           make(map[string]interface{}),
			Stats: mcp.Stats{
				Attributes: make(map[string]interface{}),
			},
		}

		for _, item := range items {
			if !item.IsStat {
				var valueMap map[string]interface{}
				if err := json.Unmarshal([]byte(item.Value), &valueMap); err != nil {
					continue
				}

				valueMap["platform"] = newPlatform

				updatedValue, err := json.Marshal(valueMap)
				if err != nil {
					continue
				}

				item.Value = string(updatedValue)
				db.Save(&item)

				itemData := mcp.BaseItem{
					Attributes: valueMap,
					TemplateId: item.TemplateID,
					Quantity:   item.Quantity,
				}
				if strings.Contains(item.TemplateID, "loadout") {
					itemData.TemplateId = "CosmeticLocker:cosmeticlocker_athena"
				}
				if profileId == "creative" {
					profileData.Items[item.UUID] = itemData
				} else {
					profileData.Items[item.TemplateID] = itemData
				}
			} else {
				var statValue interface{}
				json.Unmarshal([]byte(item.Value), &statValue)
				profileData.Stats.Attributes[item.TemplateID] = statValue
			}
		}


		profile.Revision++
		db.Save(&profile)

		response = mcp.DefaultMCPResponse{
			ProfileRevision:            profile.Revision,
			ProfileId:                  profileId,
			ProfileChangesBaseRevision: profile.Revision,
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
	}

	c.JSON(200, response)
}
