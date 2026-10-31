package dsmcp

import (
	"encoding/json"
	"fmt"
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateNewIsland(c *gin.Context) {
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
		Locale     string `json:"locale"`
		TemplateId string `json:"templateId"`
		Title      string `json:"title"`
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

	var existingIslands []accounts.Items
	db.Where("account_id = ? AND profile_id = ? AND template_id = ?", accountId, profileId, body.TemplateId).Find(&existingIslands)

	islandIndex := len(existingIslands) + 1

	var newItem accounts.Items
	newItemUUID := uuid.New().String()

	newItemValue := fmt.Sprintf(`{
		"last_used_date": "%s",
		"vk_project_id": "%s",
		"level": 1,
		"permissions": {
            "permission": "Private",
            "whitelist": ["%s"]
        },
		"islandIndex": %d,
		"locale": "%s",
		"title": "%s"
	}`, now, uuid.New().String(), accountId, islandIndex, body.Locale, body.Title)

	newItem = accounts.Items{
		AccountID:  accountId,
		ProfileID:  profileId,
		TemplateID: body.TemplateId,
		Value:      newItemValue,
		IsStat:     false,
		Quantity:   1,
		UUID:       newItemUUID,
	}

	if err := db.Create(&newItem).Error; err != nil {
		utilities.Internal.DataBaseError().
			WithIntent(utilities.Prod).Apply(c.Writer)
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
		} else {
			var statValue interface{}
			if err := json.Unmarshal([]byte(item.Value), &statValue); err == nil {
				profileData.Stats.Attributes[item.TemplateID] = statValue
			}
		}
	}

	var newItemAttributes map[string]interface{}
	json.Unmarshal([]byte(newItem.Value), &newItemAttributes)

	response = mcp.DefaultMCPResponse{
		ProfileRevision:            profile.Revision,
		ProfileId:                  profileId,
		ProfileChangesBaseRevision: profile.Revision - 1,
		ProfileCommandRevision:     profile.Revision,
		ServerTime:                 now,
		ResponseVersion:            1,
		ProfileChanges: []map[string]interface{}{
			// {
			// 	"changeType": "itemAdded",
			// 	"itemId":     newItem.UUID,
			// 	"item": map[string]interface{}{
			// 		"templateId": newItem.TemplateID,
			// 		"attributes": newItemAttributes,
			// 		"quantity":   1,
			// 	},
			// }, ts is proper way but like i cba to do it
			{
				"changeType": "fullProfileUpdate",
				"profile":    profileData,
			},
		},
	}

	c.JSON(200, response)
}
