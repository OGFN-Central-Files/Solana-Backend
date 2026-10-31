package generation

import (
	"fmt"
	"solana/modules/database"
	"solana/modules/database/entities/fortnite"
	"solana/modules/storefront/declarations"
	"solana/modules/storefront/models"
	"solana/modules/storefront/utils"
	"solana/utilities"
	"time"

	"github.com/fatih/color"
)

func FillDailyStorefront(dailySection *models.Storefront) error {
	utils.Mu.Lock()
	defer utils.Mu.Unlock()

	dailyItems, exists := utils.StorefrontData["Daily"]
	if !exists || len(dailyItems) == 0 {
		return fmt.Errorf("no daily items found")
	}

	utilities.LogWithTimestamp(color.GreenString, fmt.Sprintf("Generating up to 6 daily items for the daily storefront"))

	seenItems := make(map[string]bool)
	db := database.Get()

	for len(dailySection.CatalogEntries) < 6 {
		for _, item := range dailyItems {
			if len(dailySection.CatalogEntries) >= 6 {
				break
			}

			if seenItems[item.ID] {
				continue
			}

			currentVersion := utilities.GetConfig().CURRENT_VERSION
			entry, err := CreateShopEntry(item, "Daily", currentVersion >= 14)
			if err != nil {
				return fmt.Errorf("error creating shop entry: %w", err)
			}

			backendValue := declarations.DetermineItemBackendValue(item)

			catalogEntry := fortnite.Catalog{
				Created:    time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
				OfferId:    entry.OfferId,
				Storefront: "BRDailyStorefront",
				Name:       item.Name,
				Category:   "",
				Data:       string(utilities.MustMarshal(entry)),
				TemplateId: fmt.Sprintf("%s:%s", backendValue, item.ID),
			}

			if err := db.Create(&catalogEntry).Error; err != nil {
				return fmt.Errorf("error creating catalog entry: %w", err)
			}

			dailySection.CatalogEntries = append(dailySection.CatalogEntries, entry)
			seenItems[item.ID] = true
		}
	}

	return nil
}

func FillWeeklyStorefront(weeklySection *models.Storefront) error {
	utils.Mu.Lock()
	defer utils.Mu.Unlock()

	utilities.LogWithTimestamp(color.GreenString, "Generating items for weekly storefront")

	for section, items := range utils.StorefrontData {
		if section == "Daily" {
			continue
		}

		if len(items) == 0 {
			continue
		}

		utilities.LogWithTimestamp(color.GreenString, fmt.Sprintf("Generating %d items for section %s", len(items), section))

		for _, item := range items {
			currentVersion := utilities.GetConfig().CURRENT_VERSION
			entry, err := CreateShopEntry(item, section, currentVersion >= 14)
			if err != nil {
				fmt.Println("Skipping item due to error:", item.ID, "Error:", err)
				continue
			}
			db := database.Get()
			backendValue := declarations.DetermineItemBackendValue(item)

			var category string
			if len(entry.Categories) > 0 {
				category = entry.Categories[0]
			} else {
				category = ""
			}

			if err := db.Create(&fortnite.Catalog{
				Created:    time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
				OfferId:    entry.OfferId,
				Storefront: "BRWeeklyStorefront",
				Name:       item.Name,
				Category:   category,
				Data:       string(utilities.MustMarshal(entry)),
				TemplateId: backendValue + ":" + item.ID,
			}).Error; err != nil {
				return fmt.Errorf("error creating catalog entry: %s", err)
			}

			weeklySection.CatalogEntries = append(weeklySection.CatalogEntries, entry)
		}
	}

	return nil
}

func FillBRSeason10Storefront(BRSeason models.Storefront) error {
	utils.Mu.Lock()
	defer utils.Mu.Unlock()
	requirements := []models.Requirement{
		{
			RequirementType: "DenyOnFulfillment",
			RequiredId:      "2E43CCD24C3BE8F5ABBDF28E233B9350",
			MinQuantity:     1,
		},
	}

	if BRSeason.CatalogEntries == nil {
		BRSeason.CatalogEntries = []models.CatalogEntry{}
	}

	BRSeason.CatalogEntries = append(BRSeason.CatalogEntries,
		CreateCatalogBattlePassEntry(BRSeason, "2E43CCD24C3BE8F5ABBDF28E233B9350", "BR.Season10.BattlePass.01", "Battle Pass", "Claim your Season X Battle Pass", "Season X Now through October 6.\n\nInstantly get these items <Bold>valued at over 3,500 V-Bucks</>.\n  • <ItemName>X-Lord</> Outfit\n  • <ItemName>Catalyst</> Outfit\n  • <Bold>50% Bonus</> Season Match XP\n  • <Bold>60% Bonus</> Season Friend Match XP\n\nPlay to level up your Battle Pass, unlocking <Bold>over 100 rewards</> (typically takes 75 to 150 hours of play).\n  • <ItemName>Ultima Knight</> and <Bold>4 more Outfits</>\n  • <Bold>1,300 V-Bucks</>\n  • 7 Emotes\n  • 6 Wraps\n  • 6 Harvesting Tools\n  • 1 Pet\n  • 7 Gliders\n  • 7 Back Blings\n  • 5 Contrails\n  • 17 Sprays\n  • 3 Music Tracks\n  • 1 Toy\n  • 27 Loading Screens\n  • and so much more!\nWant it all faster? You can use V-Bucks to buy tiers any time!", 950, 0, "/Game/Catalog/DisplayAssets/DA_BR_Season10_BattlePass.DA_BR_Season10_BattlePass", requirements),
		CreateCatalogBattlePassEntry(BRSeason, "259920BC42F0AAC7C8672D856C9B622C", "BR.Season10.BattleBundle.01", "Battle Bundle", "Battle Pass + 25 tiers!", "Season X Now through October 6.\n\nInstantly get these items <Bold>valued at over 10,000 V-Bucks</>.\n  • <ItemName>X-Lord</> Outfit\n  • <ItemName>Catalyst</> Outfit\n  • <ItemName>Tilted Teknique</> Outfit\n  • <ItemName>Junk Bucket</> Glider\n  • <ItemName>Emote Camo </> Wrap\n  • <ItemName> Rift Lightning </> Contrails\n  • 300 V-Bucks\n  • <ItemName>The Final Showdown </> Music Track\n  • <Bold>70% Bonus</> Season Match XP\n  • <Bold>80% Bonus</> Season Friend Match XP\n  • and more!\n\nPlay to level up your Battle Pass, unlocking <Bold>over 75 rewards</> (typically takes 75 to 150 hours of play).\n  • <Bold>4 more Outfits</>\n  • <Bold>1,000 V-Bucks</>\n  • 6 Emotes\n  • 4 Wraps\n  • 5 Harvesting Tools\n  • 1 Pet\n  • 6 Gliders\n  • 7 Back Blings\n  • 3 Contrails\n  • 13 Sprays\n  • 2 Music Tracks\n  • 1 Toy\n  • 23 Loading Screens\n  • and so much more!\nWant it all faster? You can use V-Bucks to buy tiers any time!", 4700, 2800, "/Game/Catalog/DisplayAssets/DA_BR_Season10_BattlePassWithLevels.DA_BR_Season10_BattlePassWithLevels", requirements),
		CreateCatalogBattlePassEntry(BRSeason, "AF1B7AC14A5F6A9ED255B88902120757", "BR.Season10.SingleTier.01", "Battle Pass Tier", "", "Get great rewards now!", 150, 150, "/Game/Catalog/DisplayAssets/DA_BR_Season10_BattlePassWithLevels.DA_BR_Season10_BattlePassWithLevels", []models.Requirement{}),
	)
	db := database.Get()

	for _, entry := range BRSeason.CatalogEntries {
		catalogEntry := fortnite.Catalog{
			Created:    time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
			OfferId:    entry.OfferId,
			Storefront: "BRSeason10",
			Name:       entry.DevName,
			Category:   "",
			Data:       string(utilities.MustMarshal(entry)),
			TemplateId: "",
		}

		if err := db.Create(&catalogEntry).Error; err != nil {
			return fmt.Errorf("error creating catalog entry: %w", err)
		}
	}

	return nil
}

func FillAllSectionsInDatabase() error {
	db := database.Get()

	for section, items := range utils.StorefrontData {
		if len(items) == 0 {
			continue
		}

		utilities.LogWithTimestamp(color.GreenString, fmt.Sprintf("Generating section %s", section))

		if err := db.Create(&fortnite.ShopSections{
			Date:    time.Now().UTC(),
			Section: section,
		}).Error; err != nil {
			return fmt.Errorf("error creating catalog entry: %s", err)
		}
	}

	return nil
}
