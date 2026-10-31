package managers

import (
	"encoding/json"
	mcp "solana/classes/mcp"
	"solana/modules/database/entities/accounts"

	"github.com/google/uuid"
)

func CreateItemBase(profileId, accountId, templateId string, value interface{}, quantity int, isStat bool) accounts.Items {
	var valueJSON string
	if value == nil {
		baseValue := mcp.BaseItemAttributes{
			XP:       0,
			Level:    1,
			Variants: []mcp.Variants{},
			ItemSeen: false,
		}
		valueBytes, _ := json.Marshal(baseValue)
		valueJSON = string(valueBytes)
	} else {
		valueBytes, _ := json.Marshal(value)
		valueJSON = string(valueBytes)
	}

	if templateId == "Currency:MtxPurchased" {
		quantity = 0
	}

	return accounts.Items{
		AccountID:  accountId,
		UUID:       uuid.New().String(),
		ProfileID:  profileId,
		TemplateID: templateId,
		Value:      valueJSON,
		Quantity:   quantity,
		IsStat:     isStat,
	}
}

func CreateItem(profileId, accountId, templateId string) accounts.Items {
	return CreateItemBase(profileId, accountId, templateId, nil, 1, false)
}

func CreateLoadoutItem(profileId, accountId, templateId string) accounts.Items {
	loadoutData := map[string]interface{}{
		"favorite":    false,
		"item_seen":   false,
		"use_count":   0,
		"locker_name": "",
		"locker_slots_data": map[string]interface{}{
			"slots": map[string]interface{}{
				"Dance": map[string]interface{}{
					"items": []string{"AthenaDance:EID_DanceMoves", "", "", "", "", ""},
				},
				"Glider": map[string]interface{}{
					"items": []string{"AthenaGlider:DefaultGlider"},
				},
				"Pickaxe": map[string]interface{}{
					"items":          []string{"AthenaPickaxe:DefaultPickaxe"},
					"activeVariants": []interface{}{},
				},
				"Character": map[string]interface{}{
					"items": []string{"AthenaCharacter:CID_001_Athena_Commando_F_Default"},
					"activeVariants": []map[string]interface{}{
						{"variants": []interface{}{}},
					},
				},
				"Backpack": map[string]interface{}{
					"items": []string{""},
					"activeVariants": []map[string]interface{}{
						{"variants": []interface{}{}},
					},
				},
				"ItemWrap": map[string]interface{}{
					"items":          []string{"", "", "", "", "", "", ""},
					"activeVariants": []interface{}{nil, nil, nil, nil, nil, nil, nil},
				},
				"MusicPack": map[string]interface{}{
					"items":          []string{""},
					"activeVariants": []interface{}{nil},
				},
				"LoadingScreen": map[string]interface{}{
					"items":          []string{""},
					"activeVariants": []interface{}{nil},
				},
				"SkyDiveContrail": map[string]interface{}{
					"items":          []string{""},
					"activeVariants": []interface{}{nil},
				},
			},
		},
		"banner_icon_template":  "",
		"banner_color_template": "",
	}

	return CreateItemBase(profileId, accountId, templateId, loadoutData, 1, false)
}

func CreateCurrencyItem(profileId, accountId, templateId string) accounts.Items {
	value := map[string]interface{}{
		"platform": "EpicPC",
		"level":    1,
	}

	return CreateItemBase(profileId, accountId, templateId, value, 0, false)
}

func CreateStatItem(profileId, accountId, templateId string, value interface{}) accounts.Items {
	return CreateItemBase(profileId, accountId, templateId, value, 1, true)
}
