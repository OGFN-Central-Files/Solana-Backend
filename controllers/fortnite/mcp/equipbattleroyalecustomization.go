package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func EquipBattleRoyaleCustomization(c *gin.Context) {
	accountId := c.Param("accountId")

	if !IsCacheLoaded() {
		utilities.Internal.ServerError().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	profileId := c.Query("profileId")
	if profileId == "" {
		utilities.MCP.ProfileNotFound().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	var body struct {
		ItemToSlot      string `json:"itemToSlot"`
		SlotName        string `json:"slotName"`
		IndexWithinSlot int    `json:"indexWithinSlot"`
		VariantUpdates  []struct {
			Channel string `json:"channel"`
			Active  string `json:"active"`
		} `json:"variantUpdates"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utilities.MCP.InvalidPayload().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	db := database.Get()

	var user accounts.Users
	var profile accounts.Profiles

	if err := db.Where("\"accountId\" = ?", accountId).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	if err := db.Where("account_id = ? AND profile_id = ?", accountId, profileId).First(&profile).Error; err != nil {
		utilities.Account.AccountNotFound().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	profileChanges := make([]map[string]interface{}, 0)

	if strings.Contains(body.ItemToSlot, "_random") || body.ItemToSlot == "" {
		statModification := map[string]interface{}{
			"changeType": "statModified",
			"name":       fmt.Sprintf("favorite_%s", strings.ToLower(body.SlotName)),
			"value":      body.ItemToSlot,
		}
		profileChanges = append(profileChanges, statModification)

		favoriteKey := fmt.Sprintf("favorite_%s", strings.ToLower(body.SlotName))
		var favoriteItem accounts.Items
		if err := db.Where("template_id = ? AND account_id = ?", favoriteKey, accountId).First(&favoriteItem).Error; err == nil {
			favoriteItem.Value = body.ItemToSlot
			db.Save(&favoriteItem)
			profile.Revision++
			db.Save(&profile)
			c.JSON(200, mcp.DefaultMCPResponse{
				ProfileId:                  profileId,
				ProfileRevision:            profile.Revision,
				ProfileChangesBaseRevision: profile.Revision - 1,
				ProfileChanges:             profileChanges,
				ProfileCommandRevision:     profile.Revision,
				ServerTime:                 time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
				ResponseVersion:            1,
			})
			return
		}
	} else {
		var item accounts.Items
		itemErr := db.Where("template_id = ? AND account_id = ?", body.ItemToSlot, accountId).First(&item).Error

		if itemErr != nil {
			if !user.HasFL {
				utilities.MCP.ItemNotFound().
					WithIntent(utilities.Prod).
					Apply(c.Writer)
				return
			}

			cache := GetCachedItems()
			var newItemValue string
			found := false

			cache.Range(func(templateID, baseItem interface{}) bool {
				if templateID.(string) == body.ItemToSlot {
					itemValue, _ := json.Marshal(baseItem)
					newItemValue = string(itemValue)
					found = true
					return false
				}
				return true
			})

			if !found {
				utilities.MCP.ItemNotFound().
					WithIntent(utilities.Prod).
					Apply(c.Writer)
				return
			}

			item = accounts.Items{
				AccountID:  accountId,
				ProfileID:  profileId,
				TemplateID: body.ItemToSlot,
				Value:      newItemValue,
				IsStat:     false,
				Quantity:   1,
				UUID:       uuid.New().String(),
			}
			db.Create(&item)
		}

		var statKey string
		switch {
		case body.SlotName == "Dance" && body.IndexWithinSlot >= 0 && body.IndexWithinSlot <= 5:
			statKey = "favorite_dance"
		case body.SlotName == "ItemWrap":
			statKey = "favorite_itemwraps"
		case body.SlotName == "Character" || body.SlotName == "Backpack" ||
			body.SlotName == "Pickaxe" || body.SlotName == "Glider" ||
			body.SlotName == "SkyDiveContrail" || body.SlotName == "MusicPack" ||
			body.SlotName == "LoadingScreen":
			statKey = fmt.Sprintf("favorite_%s", strings.ToLower(body.SlotName))
		}

		if statKey != "" {
			var statItem accounts.Items
			if err := db.Where("template_id = ? AND account_id = ?", statKey, accountId).First(&statItem).Error; err == nil {
				var value interface{}

				if body.SlotName == "ItemWrap" && (body.IndexWithinSlot == -1 || body.IndexWithinSlot <= 7) {
					wraps := make([]string, 7)
					for i := range wraps {
						wraps[i] = body.ItemToSlot
					}
					value = wraps
				} else if body.SlotName == "Dance" && body.IndexWithinSlot >= 0 && body.IndexWithinSlot <= 5 {
					var dances []string
					json.Unmarshal([]byte(statItem.Value), &dances)
					if dances == nil {
						dances = make([]string, 6)
					}
					dances[body.IndexWithinSlot] = body.ItemToSlot
					value = dances
				} else {
					value = body.ItemToSlot
				}

				newValue, _ := json.Marshal(value)
				statItem.Value = string(newValue)
				db.Save(&statItem)

				profileChanges = append(profileChanges, map[string]interface{}{
					"changeType": "statModified",
					"name":       statKey,
					"value":      value,
				})
			}
		}

		if len(body.VariantUpdates) > 0 {
			var itemValue map[string]interface{}
			err := json.Unmarshal([]byte(item.Value), &itemValue)
			if err != nil {
				log.Printf("Error unmarshaling item value: %v", err)
				return
			}

			if itemValue["attributes"] == nil {
				itemValue["attributes"] = make(map[string]interface{})
			}

			attributes := itemValue["attributes"].(map[string]interface{})

			var existingVariants []map[string]interface{}
			if variants, ok := attributes["variants"].([]interface{}); ok {
				for _, v := range variants {
					if variant, ok := v.(map[string]interface{}); ok {
						existingVariants = append(existingVariants, variant)
					}
				}
			}

			for _, update := range body.VariantUpdates {
				variantFound := false
				for i, variant := range existingVariants {
					if channel, ok := variant["channel"].(string); ok && channel == update.Channel {
						if owned, exists := variant["owned"].([]interface{}); exists {
							variant["owned"] = owned
						} else {
							variant["owned"] = []interface{}{}
						}
						existingVariants[i]["active"] = update.Active
						variantFound = true
						break
					}
				}

				if !variantFound {
					existingVariants = append(existingVariants, map[string]interface{}{
						"channel": update.Channel,
						"active":  update.Active,
						"owned":   []interface{}{},
					})
				}
			}

			attributes["variants"] = existingVariants

			newAttributesValue, err := json.Marshal(attributes)
			if err != nil {
				log.Printf("Error marshaling updated attributes: %v", err)
				return
			}

			item.Value = string(newAttributesValue)
			if err := db.Model(&item).Update("Value", item.Value).Error; err != nil {
				log.Printf("Error updating item attributes in database: %v", err)
				return
			}

			profileChanges = append(profileChanges, map[string]interface{}{
				"changeType":     "itemAttrChanged",
				"itemId":         body.ItemToSlot,
				"attributeName":  "variants",
				"attributeValue": existingVariants,
			})
		}

	}

	profile.Revision++
	db.Save(&profile)

	c.JSON(200, mcp.DefaultMCPResponse{
		ProfileId:                  profileId,
		ProfileRevision:            profile.Revision,
		ProfileChangesBaseRevision: profile.Revision - 1,
		ProfileChanges:             profileChanges,
		ProfileCommandRevision:     profile.Revision,
		ServerTime:                 time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
		ResponseVersion:            1,
	})
}
