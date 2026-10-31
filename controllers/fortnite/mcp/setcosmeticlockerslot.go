package mcp

import (
	"encoding/json"
	"net/http"
	"time"

	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SetCosmeticLockerSlot(c *gin.Context) {
	accountID := c.Param("accountId")
	if !IsCacheLoaded() {
		utilities.Internal.ServerError().WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	profileID := c.Query("profileId")
	if profileID == "" {
		utilities.MCP.ProfileNotFound().WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var body struct {
		LockerItem     string `json:"lockerItem"`
		Category       string `json:"category"`
		ItemToSlot     string `json:"itemToSlot"`
		SlotIndex      int    `json:"slotIndex"`
		VariantUpdates []struct {
			Channel string   `json:"channel"`
			Active  string   `json:"active"`
			Owned   []string `json:"owned,omitempty"`
		} `json:"variantUpdates,omitempty"`
		OptLockerUseCountOverride int `json:"optLockerUseCountOverride,omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utilities.MCP.InvalidPayload().WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	db := database.Get()

	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", accountID).First(&user).Error; err != nil {
		utilities.Account.AccountNotFound().WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var profile accounts.Profiles
	if err := db.Where("account_id = ? AND profile_id = ?", accountID, profileID).First(&profile).Error; err != nil {
		if profileID == "athena" {
			c.JSON(http.StatusOK, mcp.DefaultMCPResponse{
				ProfileChanges:             []map[string]interface{}{},
				ProfileId:                  profileID,
				ProfileRevision:            1,
				ProfileChangesBaseRevision: 0,
				ProfileCommandRevision:     1,
				ResponseVersion:            1,
				ServerTime:                 time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
			})
			return
		} else {
			utilities.MCP.ProfileNotFound().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}
	}

	var loadout accounts.Items
	if err := db.Where("template_id = ? AND account_id = ?", body.LockerItem, accountID).First(&loadout).Error; err != nil {
		utilities.MCP.ItemNotFound().WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	if body.ItemToSlot == "" || (len(body.ItemToSlot) >= 7 && body.ItemToSlot[len(body.ItemToSlot)-7:] == "_random") {
		var randomLoadoutData map[string]json.RawMessage
		if err := json.Unmarshal([]byte(loadout.Value), &randomLoadoutData); err != nil {
			utilities.Internal.JsonParsingFailed().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		var randomLockerSlotsData map[string]json.RawMessage
		if err := json.Unmarshal(randomLoadoutData["locker_slots_data"], &randomLockerSlotsData); err != nil {
			randomLockerSlotsData = make(map[string]json.RawMessage)
		}

		var randomSlots map[string]json.RawMessage
		if slotsRaw, ok := randomLockerSlotsData["slots"]; ok {
			if err := json.Unmarshal(slotsRaw, &randomSlots); err != nil {
				randomSlots = make(map[string]json.RawMessage)
			}
		} else {
			randomSlots = make(map[string]json.RawMessage)
		}

		if _, ok := randomSlots[body.Category]; !ok {
			emptyItemsJSON, _ := json.Marshal([]string{})
			categoryData := map[string]json.RawMessage{"items": emptyItemsJSON}
			categoryDataJSON, _ := json.Marshal(categoryData)
			randomSlots[body.Category] = categoryDataJSON
		}

		var categoryData map[string]json.RawMessage
		if err := json.Unmarshal(randomSlots[body.Category], &categoryData); err != nil {
			categoryData = make(map[string]json.RawMessage)
		}

		itemsJSON, _ := json.Marshal([]string{body.ItemToSlot})
		categoryData["items"] = itemsJSON

		categoryDataJSON, _ := json.Marshal(categoryData)
		randomSlots[body.Category] = categoryDataJSON
		slotsJSON, _ := json.Marshal(randomSlots)
		randomLockerSlotsData["slots"] = slotsJSON
		lockerSlotsDataJSON, _ := json.Marshal(randomLockerSlotsData)
		randomLoadoutData["locker_slots_data"] = lockerSlotsDataJSON

		randomLoadoutDataJSON, _ := json.Marshal(randomLoadoutData)
		loadout.Value = string(randomLoadoutDataJSON)
		loadout.AccountID = accountID
		loadout.ProfileID = "athena"
		loadout.TemplateID = body.LockerItem

		if err := db.Model(&loadout).Updates(loadout).Error; err != nil {
			utilities.Internal.DataBaseError().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		c.JSON(http.StatusOK, mcp.DefaultMCPResponse{
			ProfileChanges: []map[string]interface{}{
				{
					"changeType": "item_updated",
					"item":       body.ItemToSlot,
					"category":   body.Category,
					"slotIndex":  body.SlotIndex,
					"loadout":    randomLoadoutData,
				},
			},
			ProfileId:                  profileID,
			ProfileRevision:            1,
			ProfileChangesBaseRevision: 0,
			ProfileCommandRevision:     1,
			ResponseVersion:            1,
			ServerTime:                 time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
		})
		return
	}

	var item accounts.Items
	if err := db.Where("template_id = ? AND account_id = ?", body.ItemToSlot, accountID).First(&item).Error; err != nil {
		if !user.HasFL {
			utilities.Storefront.InvalidItem().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		if itemValue, ok := GetCachedItems().Load(body.ItemToSlot); !ok {
			utilities.MCP.ItemNotFound().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		} else if baseItem, ok := itemValue.(mcp.BaseItem); !ok {
			utilities.Internal.ServerError().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		} else if attributesJSON, err := json.Marshal(baseItem.Attributes); err != nil {
			utilities.Internal.JsonParsingFailed().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		} else {
			item = accounts.Items{
				AccountID:  accountID,
				TemplateID: body.ItemToSlot,
				UUID:       uuid.New().String(),
				Quantity:   baseItem.Quantity,
				Value:      string(attributesJSON),
				IsStat:     false,
				ProfileID:  "athena",
			}
		}
	}

	var mainLoadoutData map[string]json.RawMessage
	if err := json.Unmarshal([]byte(loadout.Value), &mainLoadoutData); err != nil {
		utilities.Internal.JsonParsingFailed().WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var lockerSlotsData map[string]json.RawMessage
	if err := json.Unmarshal(mainLoadoutData["locker_slots_data"], &lockerSlotsData); err != nil {
		utilities.Internal.JsonParsingFailed().WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var slots map[string]json.RawMessage
	if slotsRaw, ok := lockerSlotsData["slots"]; ok {
		if err := json.Unmarshal(slotsRaw, &slots); err != nil {
			utilities.Internal.JsonParsingFailed().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}
	} else {
		var nestedLockerData map[string]json.RawMessage
		if err := json.Unmarshal(lockerSlotsData["locker_slots_data"], &nestedLockerData); err != nil {
			utilities.Internal.JsonParsingFailed().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}
		if err := json.Unmarshal(nestedLockerData["slots"], &slots); err != nil {
			utilities.Internal.JsonParsingFailed().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}
	}

	var categoryData map[string]json.RawMessage
	if err := json.Unmarshal(slots[body.Category], &categoryData); err != nil {
		utilities.Internal.JsonParsingFailed().WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	switch body.Category {
	case "Dance":
		var items []string
		if err := json.Unmarshal(categoryData["items"], &items); err != nil {
			items = []string{}
		}

		for len(items) <= body.SlotIndex {
			items = append(items, "")
		}

		items[body.SlotIndex] = body.ItemToSlot
		itemsJSON, _ := json.Marshal(items)
		categoryData["items"] = itemsJSON

	case "ItemWrap":
		var items []string
		if err := json.Unmarshal(categoryData["items"], &items); err != nil {
			items = []string{}
		}

		for len(items) < 7 {
			items = append(items, "")
		}

		if body.SlotIndex == -1 {
			for i := range items {
				items[i] = body.ItemToSlot
			}
		} else {
			items[body.SlotIndex] = body.ItemToSlot
		}

		itemsJSON, _ := json.Marshal(items)
		categoryData["items"] = itemsJSON

	default:
		itemsJSON, _ := json.Marshal([]string{body.ItemToSlot})
		categoryData["items"] = itemsJSON
	}

	categoryDataJSON, _ := json.Marshal(categoryData)
	slots[body.Category] = categoryDataJSON
	slotsJSON, _ := json.Marshal(slots)
	lockerSlotsData["slots"] = slotsJSON
	lockerSlotsDataJSON, _ := json.Marshal(lockerSlotsData)
	mainLoadoutData["locker_slots_data"] = lockerSlotsDataJSON

	profileChanges := []map[string]interface{}{
		{
			"changeType":     "itemAttrChanged",
			"itemId":         body.LockerItem,
			"attributeName":  "locker_slots_data",
			"attributeValue": lockerSlotsData,
		},
	}

	if len(body.VariantUpdates) > 0 && item.Value != "" {
		var itemValue map[string]json.RawMessage
		if err := json.Unmarshal([]byte(item.Value), &itemValue); err != nil {
			utilities.Internal.JsonParsingFailed().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		var variants []map[string]interface{}
		if variantsRaw, ok := itemValue["variants"]; ok {
			if err := json.Unmarshal(variantsRaw, &variants); err != nil {
				variants = []map[string]interface{}{}
			}
		} else {
			variants = []map[string]interface{}{}
		}

		for _, update := range body.VariantUpdates {
			variantData := map[string]interface{}{
				"channel": update.Channel,
				"active":  update.Active,
			}

			if len(update.Owned) > 0 {
				variantData["owned"] = update.Owned
			}

			found := false
			for i, variant := range variants {
				if channel, ok := variant["channel"].(string); ok && channel == update.Channel {
					variants[i] = variantData
					found = true
					break
				}
			}

			if !found {
				variants = append(variants, variantData)
			}
		}

		variantsJSON, _ := json.Marshal(variants)
		itemValue["variants"] = variantsJSON
		itemValueJSON, _ := json.Marshal(itemValue)
		item.Value = string(itemValueJSON)

		profileChanges = append(profileChanges, map[string]interface{}{
			"changeType":     "itemAttrChanged",
			"itemId":         body.ItemToSlot,
			"attributeName":  "variants",
			"attributeValue": variants,
		})

		if err := db.Model(&item).Update("value", item.Value).Error; err != nil {
			utilities.Internal.DataBaseError().WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}
	}

	mainLoadoutDataJSON, _ := json.Marshal(mainLoadoutData)
	if err := db.Model(&loadout).Update("value", string(mainLoadoutDataJSON)).Error; err != nil {
		utilities.Internal.DataBaseError().WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	profile.Revision++
	if err := db.Model(&profile).Update("revision", profile.Revision).Error; err != nil {
		utilities.Internal.DataBaseError().WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	c.JSON(http.StatusOK, mcp.DefaultMCPResponse{
		ProfileChanges:             profileChanges,
		ProfileId:                  profileID,
		ProfileRevision:            profile.Revision,
		ProfileChangesBaseRevision: profile.Revision - 1,
		ProfileCommandRevision:     profile.Revision,
		ResponseVersion:            1,
		ServerTime:                 time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
	})
}
