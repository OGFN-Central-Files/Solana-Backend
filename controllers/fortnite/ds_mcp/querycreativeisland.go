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

func QueryCreativeIsland(c *gin.Context) {
	accountId := c.Param("accountId")
	UUID := c.Param("uuid")
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

	if err := db.Where("account_id = ? AND profile_id = ?", accountId, "creative").First(&profile).Error; err != nil {
		profileFound = false
	} else if profileFound {
		db.Where("account_id = ? AND profile_id = ?", accountId, "creative").Find(&items)
	}

	if !profileFound {
		response = mcp.DefaultMCPResponse{
			ProfileRevision:            0,
			ProfileId:                  "creative",
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
						ProfileId:  "creative",
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
		for _, item := range items {
			if !item.IsStat {
				var valueMap interface{}
				json.Unmarshal([]byte(item.Value), &valueMap)
				itemData := mcp.BaseItem{
					Attributes: valueMap,
					TemplateId: item.TemplateID,
					Quantity:   item.Quantity,
				}

				if item.UUID == UUID {
					c.JSON(200, itemData)
					return
				}
			}
		}
	}

	c.JSON(200, response)
}
