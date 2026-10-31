package managers

import (
	"fmt"
	"solana/classes/mcp"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"time"
)

// credit to skies for the original c# code

func CreateProfile(profileType, accountId string, displayName string) (*accounts.Profiles, error) {
	db := database.Get()

	newProfile := accounts.Profiles{
		AccountID: accountId,
		ProfileID: profileType,
		Revision:  0,
	}

	if err := db.Create(&newProfile).Error; err != nil {
		return nil, fmt.Errorf("failed to create profile: %v", err)
	}

	var items []accounts.Items
	switch profileType {
	case mcp.Athena:
		items = createAthenaItems(accountId)
	case mcp.CommonCore:
		items = createCommonCoreItems(accountId)
	case mcp.Creative:
		items = createCreativeItems(accountId, displayName)
	}

	for _, item := range items {
		if err := db.Create(&item).Error; err != nil {
			return nil, fmt.Errorf("failed to create item: %v", err)
		}
	}

	return &newProfile, nil
}

func createAthenaItems(accountId string) []accounts.Items {
	items := []accounts.Items{
		CreateItem(mcp.Athena, accountId, "AthenaPickaxe:DefaultPickaxe"),
		CreateItem(mcp.Athena, accountId, "AthenaGlider:DefaultGlider"),
		CreateItem(mcp.Athena, accountId, "AthenaDance:EID_DanceMoves"),
		CreateLoadoutItem(mcp.Athena, accountId, "loadout1"),
	}

	statItems := map[string]interface{}{
		"use_random_loadout":   false,
		"past_seasons":         []interface{}{},
		"season_match_boost":   0,
		"loadouts":             []string{"loadout1"},
		"mfa_reward_claimed":   false,
		"rested_xp_overflow":   0,
		"current_mtx_platform": "Epic",
		"last_xp_interaction":  time.Now().UTC().Format(time.RFC3339),
		"quest_manager": map[string]interface{}{
			"dailyLoginInterval": time.Time{}.Format(time.RFC3339),
			"dailyQuestRerolls":  1,
		},
		"book_level":               1,
		"season_num":               1,
		"book_purchased":           false,
		"party_assist_quest":       "",
		"book_xp":                  0,
		"level":                    1,
		"xp":                       0,
		"battlestars":              0,
		"battlestars_season_total": 0,
		"rested_xp":                2500,
		"rested_xp_mult":           4.0,
		"accountLevel":             1,
		"style_points":             0,
		"alien_style_points":       0,
		"active_loadout_index":     0,
		"favorite_character":       "",
		"favorite_backpack":        "",
		"favorite_pickaxe":         "AthenaPickaxe:DefaultPickaxe",
		"favorite_glider":          "AthenaGlider:DefaultGlider",
		"favorite_skydivecontrail": "",
		"favorite_dance":           []string{"", "", "", "", "", ""},
		"favorite_itemwraps":       []string{"", "", "", "", "", "", ""},
		"favorite_loadingscreen":   "",
		"favorite_musicpack":       "",
		"banner_icon":              "StandardBanner31",
		"banner_color":             "DefaultColor1",
	}

	for key, value := range statItems {
		items = append(items, CreateStatItem(mcp.Athena, accountId, key, value))
	}

	return items
}

func createCreativeItems(accountId string, displayName string) []accounts.Items {
	items := []accounts.Items{}

	statItems := map[string]interface{}{
		"max_creative_plots": 30,
		"support_code":       displayName,
		"creator_name":       displayName,
	}

	for key, value := range statItems {
		items = append(items, CreateStatItem(mcp.CommonCore, accountId, key, value))
	}

	return items
}

func createCommonCoreItems(accountId string) []accounts.Items {
	items := []accounts.Items{
		CreateCurrencyItem(mcp.CommonCore, accountId, "Currency:MtxPurchased"),
	}

	statItems := map[string]interface{}{
		"survey_data":            map[string]interface{}{},
		"personal_offers":        map[string]interface{}{},
		"intro_game_played":      true,
		"import_friends_claimed": map[string]interface{}{},
		"mtx_purchase_history": map[string]interface{}{
			"purchases":     []interface{}{},
			"refundCredits": 3,
			"refundsUsed":   0,
		},
		"undo_cooldowns":           []interface{}{},
		"mtx_affiliate_set_time":   "None",
		"inventory_limit_bonus":    0,
		"current_mtx_platform":     "EpicPC",
		"mtx_affiliate":            "None",
		"weekly_purchases":         map[string]interface{}{},
		"daily_purchases":          map[string]interface{}{},
		"ban_history":              map[string]interface{}{},
		"in_app_purchases":         map[string]interface{}{},
		"permissions":              []interface{}{},
		"mfa_enabled":              true,
		"allowed_to_send_gifts":    true,
		"allowed_to_receive_gifts": true,
		"gift_history":             map[string]interface{}{},
		"banner_icon":              "StandardBanner31",
		"banner_color":             "DefaultColor1",
		"homebase_name":            "",
	}

	for key, value := range statItems {
		items = append(items, CreateStatItem(mcp.CommonCore, accountId, key, value))
	}

	for i := 1; i <= 21; i++ {
		templateId := fmt.Sprintf("HomebaseBannerColor:DefaultColor%d", i)
		items = append(items, CreateItemBase(mcp.CommonCore, accountId, templateId,
			map[string]interface{}{"item_seen": false}, 1, false))
	}

	for i := 1; i <= 31; i++ {
		templateId := fmt.Sprintf("HomebaseBannerIcon:StandardBanner%d", i)
		items = append(items, CreateItemBase(mcp.CommonCore, accountId, templateId,
			map[string]interface{}{"item_seen": false}, 1, false))
	}

	return items
}
