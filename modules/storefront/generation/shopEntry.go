package generation

import (
	"fmt"
	"solana/modules/storefront/declarations"
	"solana/modules/storefront/generation/displayassets"
	"solana/modules/storefront/models"
	"solana/modules/storefront/scraper"
	"solana/modules/storefront/utils"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var categoryTracker = struct {
	sync.Mutex
	data               map[string][]string
	lastPickaxeRemoval time.Time
}{data: make(map[string][]string)}

var setPriceCache = sync.Map{}

func CreateShopEntry(item scraper.Item, section string, s14above bool) (models.CatalogEntry, error) {
	if item.Set.Value != "" {
		if price, ok := setPriceCache.Load(item.Set.Value); ok {
			item.Price = price.(int)
		} else {
			for _, otherItem := range utils.StorefrontData[section] {
				if otherItem.Set.Value == item.Set.Value {
					if otherItem.Price > 0 {
						setPriceCache.Store(item.Set.Value, otherItem.Price)
						item.Price = otherItem.Price
					}
					break
				}
			}
		}
	}

	if (item.Price == 0 || item.Price < 0 || item.Name == "") && item.Name != "Heartspan" {
		return models.CatalogEntry{}, fmt.Errorf("Invalid item data")
	}

	entry := createBaseShopEntry(item)

	if section == "Featured" && !s14above {
		entry.DisplayAssetPath = displayassets.SetDisplayAsset("DA_Featured_" + item.ID)
		entry.NewDisplayAssetPath = displayassets.SetNewDisplayAssetPath(item.ID)

		entry.Meta.DisplayAssetPath = entry.DisplayAssetPath
		entry.Meta.NewDisplayAssetPath = entry.NewDisplayAssetPath
		entry.Meta.SectionId = "Featured"
		entry.Meta.TileSize = "Normal"
		entry.Meta.LayoutId = "Featured.99"
		entry.Meta.AnalyticOfferGroupId = "Featured"
		categoryOverrides := map[string]string{
			"Crystal": "Dino Guard",
		}

		if fortniteIsRetarded, exists := categoryOverrides[item.Name]; exists {
			entry.Categories = []string{fortniteIsRetarded}
		} else {
			entry.Categories = []string{item.Category}
		}
	} else {
		entry.DisplayAssetPath = displayassets.SetDisplayAsset("DA_Featured_" + item.ID)
		entry.NewDisplayAssetPath = displayassets.SetNewDisplayAssetPath(item.ID)

		entry.Meta.DisplayAssetPath = entry.DisplayAssetPath
		entry.Meta.NewDisplayAssetPath = entry.NewDisplayAssetPath
		entry.Meta.SectionId = section
		if strings.Contains(strings.ToLower(item.ID), "cid") {
			entry.Meta.TileSize = "Normal"
		} else {
			entry.Meta.TileSize = "Small"
		}
		entry.Meta.LayoutId = section + ".99"
		entry.Meta.AnalyticOfferGroupId = section
		categoryOverrides := map[string]string{
			"Crystal": "Dino Guard",
		}

		if fortniteIsRetarded, exists := categoryOverrides[item.Name]; exists {
			entry.Categories = []string{fortniteIsRetarded}
		} else {
			if section != "Daily" {
				entry.Categories = []string{item.Category}
			} else {
				entry.Categories = []string{}
			}
		}
	}

	isPickaxe := strings.Contains(strings.ToLower(item.ID), "pickaxe_")
	isCharacter := strings.Contains(strings.ToLower(item.ID), "cid_")

	categoryTracker.Lock()
	if isCharacter {
		categoryTracker.data[item.ID] = entry.Categories
	}
	categoryTracker.Unlock()

	lastRemovalTime := categoryTracker.lastPickaxeRemoval
	oneWeek := 7 * 24 * time.Hour
	now := time.Now()

	if isPickaxe {
		categoryTracker.Lock()
		if now.Sub(lastRemovalTime) >= oneWeek {
			categoryTracker.lastPickaxeRemoval = now

			if item.Set.Value == "Goalbound" {
				entry.Categories = []string{item.Set.Value}
			} else {
				for _, charCategories := range categoryTracker.data {
					if utils.StringSlicesEqual(charCategories, entry.Categories) {
						entry.Categories = []string{}
						break
					}
				}
			}
		}
		categoryTracker.Unlock()
	}

	tileSize := "Small"
	if section == "Featured" && !s14above {
		tileSize = "Normal"
	}
	if strings.Contains(strings.ToLower(item.ID), "cid") {
		tileSize = "Normal"
	} else {
		tileSize = "Small"
	}
	entry.MetaInfo = append(entry.MetaInfo, models.MetaInfo{
		Key:   "TileSize",
		Value: tileSize,
	})
	entry.MetaInfo = append(entry.MetaInfo, models.MetaInfo{
		Key:   "SectionId",
		Value: section,
	})

	entry.GiftInfo.PurchaseRequirements = append(entry.GiftInfo.PurchaseRequirements, entry.Requirements...)

	return entry, nil
}

func createBaseShopEntry(item scraper.Item) models.CatalogEntry {
	backendValue := declarations.DetermineItemBackendValue(item)

	return models.CatalogEntry{
		OfferId:   fmt.Sprintf("v2:/%s", uuid.New().String()),
		OfferType: "StaticPrice",
		DevName:   fmt.Sprintf("[VIRTUAL] 1x %s:%s for %d MtxCurrency", backendValue, item.ID, item.Price),
		ItemGrants: []models.ItemGrant{
			{
				TemplateId: fmt.Sprintf("%s:%s", backendValue, item.ID),
				Quantity:   1,
			},
		},
		Requirements: []models.Requirement{
			{
				RequirementType: "DenyOnItemOwnership",
				RequiredId:      fmt.Sprintf("%s:%s", backendValue, item.ID),
				MinQuantity:     1,
			},
		},
		Categories: []string{},
		MetaInfo: []models.MetaInfo{
			{
				Key:   "BannerOverride",
				Value: "",
			},
		},
		Prices: []models.Price{
			{
				CurrencyType:        "MtxCurrency",
				CurrencySubType:     "Currency",
				RegularPrice:        item.Price,
				DynamicRegularPrice: item.Price,
				FinalPrice:          item.Price,
				SaleExpiration:      "9999-12-31T23:59:59.999Z",
				BasePrice:           item.Price,
			},
		},
		GiftInfo: models.GiftInfo{
			BIsEnabled:              true,
			ForcedGiftBoxTemplateId: "",
			PurchaseRequirements:    []models.Requirement{},
			GiftRecordIds:           []string{},
		},
		Meta: models.Meta{
			NewDisplayAssetPath:  "",
			DisplayAssetPath:     "",
			SectionId:            "Daily",
			TileSize:             "Small",
			LayoutId:             "Daily.99",
			AnalyticOfferGroupId: "Daily",
		},
		DisplayAssetPath:     displayassets.SetDisplayAsset(item.ID),
		NewDisplayAssetPath:  "",
		Refundable:           true,
		Title:                "",
		Description:          "",
		ShortDescription:     "",
		AppStoreId:           []string{},
		FulfillmentIds:       []interface{}{},
		DailyLimit:           -1,
		WeeklyLimit:          -1,
		MonthlyLimit:         -1,
		SortPriority:         0,
		CatalogGroupPriority: 0,
		FilterWeight:         0,
		BannerOverride:       "",
		MatchFilter:          "",
		AdditionalGrants:     []models.ItemGrant{},
	}
}

func CreateCatalogBattlePassEntry(section models.Storefront, offerId string, devName string, title string, shortDescription string, description string, regularPrice int, finalPrice int, displayAssetPath string, requirements []models.Requirement) models.CatalogEntry {
	return models.CatalogEntry{
		OfferId:   offerId,
		DevName:   devName,
		OfferType: "StaticPrice",
		Prices: []models.Price{
			{
				CurrencyType:        "MtxCurrency",
				CurrencySubType:     "",
				RegularPrice:        regularPrice,
				DynamicRegularPrice: -1,
				FinalPrice:          finalPrice,
				SaleType:            "PercentOff",
				SaleExpiration:      "9999-12-31T23:59:59.999Z",
				BasePrice:           finalPrice,
			},
		},
		Categories:   []string{},
		DailyLimit:   -1,
		WeeklyLimit:  -1,
		MonthlyLimit: -1,
		Refundable:   false,
		AppStoreId: []string{
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
		},
		Requirements: requirements,
		MetaInfo: []models.MetaInfo{
			{
				Key:   "Preroll",
				Value: "False",
			},
		},
		CatalogGroupPriority: 0,
		SortPriority:         1,
		Description:          description,
		ShortDescription:     shortDescription,
		Title:                title,
		DisplayAssetPath:     displayAssetPath,
		ItemGrants:           []models.ItemGrant{},
	}
}
