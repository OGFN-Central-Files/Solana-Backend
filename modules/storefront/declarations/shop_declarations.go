package declarations

import (
	"solana/modules/storefront/models"
	"solana/modules/storefront/scraper"
	"strings"
)

func CreateSection(section string) models.Storefront {
	return models.Storefront{
		Name:           section,
		CatalogEntries: []models.CatalogEntry{},
	}
}

func DetermineItemBackendValue(item scraper.Item) string {
	itemTypes := map[string]string{
		"EID":                        "AthenaDance",
		"Emoji":                      "AthenaDance",
		"SPID":                       "AthenaDance",
		"TOY":                        "AthenaDance",
		"LSID":                       "AthenaLoadingScreen",
		"Trails":                     "AthenaSkyDiveContrail",
		"MtxGiveaway":                "Currency",
		"Glider":                     "AthenaGlider",
		"BID":                        "AthenaBackpack",
		"PetCarrier":                 "AthenaBackpack",
		"MusicPack":                  "AthenaMusicPack",
		"Pickaxe":                    "AthenaPickaxe",
		"AthenaSeasonXpBoost":        "Token",
		"AthenaSeasonFriendXpBoost":  "Token",
		"CID":                        "AthenaCharacter",
		"AthenaSeasonMergedXpBoosts": "Token",
		"Wrap":                       "AthenaItemWrap",
		"AthenaSeasonalXP":           "Token",
		"AthenaNextSeasonTierBoost":  "Token",
		"AthenaNextSeasonXPBoost":    "Token",
		"VTID":                       "CosmeticVariantToken",
	}

	for prefix, backendValue := range itemTypes {
		if strings.Contains(item.ID, prefix) {
			return backendValue
		}
	}

	return ""
}
