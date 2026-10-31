package mcp

import (
	"net/http"
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	utilities "solana/utilities"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SetPartyAssistQuest(c *gin.Context) {
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
		QuestToPinAsPartyAssist string `json:"questToPinAsPartyAssist"`
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
		var value accounts.Items
		valueFound := false
		if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ?", accountId, profileId, "party_assist_quest").First(&value).Error; err != nil {
			valueFound = false
		} else {
			valueFound = true
		}

		profileChanges := []map[string]interface{}{}

		if !valueFound {
			value = accounts.Items{
				AccountID:  accountId,
				TemplateID: "party_assist_quest",
				ProfileID:  "athena",
				UUID:       uuid.New().String(),
				IsStat:     true,
				Value:      "\"" + body.QuestToPinAsPartyAssist + "\"",
				Quantity:   1,
			}
			db.Create(&value)
			profileChanges = append(profileChanges, map[string]interface{}{
				"changeType": "statModified",
				"name":       "party_assist_quest",
				"value":      body.QuestToPinAsPartyAssist,
			})
		} else {
			value.Value = "\"" + body.QuestToPinAsPartyAssist + "\""
			db.Save(&value)
			profileChanges = append(profileChanges, map[string]interface{}{
				"changeType": "statModified",
				"name":       "party_assist_quest",
				"value":      body.QuestToPinAsPartyAssist,
			})
		}

		profile.Revision++
		db.Save(&profile)

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
