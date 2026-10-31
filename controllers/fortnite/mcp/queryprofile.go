package mcp

import (
	"encoding/json"
	"log"
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	utilities "solana/utilities"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func QueryProfile(c *gin.Context) {
	accountId := c.Param("accountId")
	profileId := c.Query("profileId")
	db := database.Get()

	if profileId == "" {
		utilities.Internal.ValidationFailed().
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

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
		// profile.Revision++
		// db.Save(&profile)

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
				var valueMap interface{}
				json.Unmarshal([]byte(item.Value), &valueMap)
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
				if profileId == "common_core" && user.MatchmakingBannedReason != "" {
					banStartTime, _ := time.Parse(time.RFC3339, user.MatchmakingBannedSince)
					banEndTime, _ := time.Parse(time.RFC3339, user.MatchmakingBannedUntil)
					banDurationDays := int(banEndTime.Sub(banStartTime).Hours() / 24)

					profileData.Stats.Attributes["ban_status"] = gin.H{
						"bRequiresUserAck":     true,
						"bBanHasStarted":       true,
						"banStartTimeUtc":      user.MatchmakingBannedSince,
						"banDurationDays":      banDurationDays,
						"banReasons":           []string{user.MatchmakingBannedReason},
						"additionalInfo":       "",
						"exploitProgramName":   "",
						"competitiveBanReason": "None",
					}
				}
				profileData.Stats.Attributes[item.TemplateID] = statValue
			}
		}

		if profileId == "athena" {
			profileData.Stats.Attributes["season_num"] = utilities.Parse(c.GetHeader("User-Agent")).Season

			for _, item := range items {
				if item.TemplateID == "season_num" {
					seasonNum := strconv.Itoa(utilities.Parse(c.GetHeader("User-Agent")).Season)
					if err := db.Model(&item).Update("value", seasonNum).Error; err != nil {
						log.Fatalf("failed to update season_num: %v", err)
						utilities.Internal.DataBaseError().WithIntent(utilities.Prod).Apply(c.Writer)
						return
					}
				}
			}
		}

		if profileId == "athena" {
			var quests []accounts.Quests
			db.Where("account_id = ? and profile_id = ? ", accountId, "athena").Find(&quests)

			for _, quest := range quests {
				var questAttributes map[string]interface{}
				json.Unmarshal([]byte(quest.Value), &questAttributes)
				profileData.Items[quest.TemplateID] = mcp.BaseItem{
					TemplateId: quest.TemplateID,
					Attributes: questAttributes,
					Quantity:   1,
				}
			}
		}

		if user.HasFL && profileId == "athena" && IsCacheLoaded() {
			cache := GetCachedItems()
			cache.Range(func(templateID, baseItem interface{}) bool {
				if _, exists := profileData.Items[templateID.(string)]; !exists {
					profileData.Items[templateID.(string)] = baseItem.(mcp.BaseItem)
				}
				return true
			})
		}

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
