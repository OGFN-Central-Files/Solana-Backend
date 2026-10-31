package launcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"solana/modules/database"
	"solana/modules/database/entities/fortnite"
	"solana/modules/database/entities/solaris"
	"solana/utilities"

	"github.com/gin-gonic/gin"
)

func GetLauncher(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"playerCount": 0,
	})
}

func GetLauncherPostsV2(c *gin.Context) {
	db := database.Get()
	var posts []solaris.Posts

	result := db.Raw(`
		SELECT id, title, date, content, features 
		FROM launcher_posts 
		ORDER BY date DESC`).Scan(&posts)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	var transformedPosts []gin.H
	for _, post := range posts {
		var features []string
		if err := json.Unmarshal([]byte(post.Features), &features); err != nil {
			features = []string{}
		}
		transformedPosts = append(transformedPosts, gin.H{
			"id":       post.ID,
			"title":    post.Title,
			"date":     post.Date,
			"content":  post.Content,
			"features": features,
		})
	}

	c.JSON(http.StatusOK, transformedPosts)
}

func GetLauncherUpdater(c *gin.Context) {
	db := database.Get()
	var updater solaris.Updater

	result := db.Raw(`
		SELECT version, pub_date, url, signature, notes 
		FROM launcher_updater 
		WHERE version = (SELECT MAX(version) FROM launcher_updater)`).Scan(&updater)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updater"})
		return
	}

	if updater.Version == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Launcher updater  not found"})
		return
	}

	c.JSON(http.StatusOK, updater)
}

func GetLauncherNews(c *gin.Context) {
	db := database.Get()
	var news []solaris.Banners

	result := db.Raw(`
		SELECT banner_id, title, subtitle, description, background_type, background_url, 
			   overlay_color, overlay_opacity, theme_primary, theme_secondary, theme_text, 
			   theme_accent, cta_text, cta_url, cta_type, is_premium, priority, expire_at, 
			   tags, enabled, created_at, updated_at 
		FROM news_banners 
		WHERE enabled = true 
		ORDER BY priority DESC`).Scan(&news)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch news"})
		return
	}

	var transformedNews []gin.H
	for _, item := range news {
		transformedNews = append(transformedNews, gin.H{
			"bannerId":       item.BannerId,
			"title":          item.Title,
			"subtitle":       item.Subtitle,
			"description":    item.Description,
			"backgroundType": item.BackgroundType,
			"backgroundUrl":  item.BackgroundUrl,
			"overlayColor":   item.OverlayColor,
			"overlayOpacity": item.OverlayOpacity,
			"themePrimary":   item.ThemePrimary,
			"themeSecondary": item.ThemeSecondary,
			"themeText":      item.ThemeText,
			"themeAccent":    item.ThemeAccent,
			"ctaText":        item.CtaText,
			"ctaUrl":         item.CtaUrl,
			"ctaType":        item.CtaType,
			"isPremium":      item.IsPremium,
			"priority":       item.Priority,
			"expireAt":       item.ExpireAt,
			"tags":           item.Tags,
			"enabled":        item.Enabled,
			"createdAt":      item.CreatedAt,
			"updatedAt":      item.UpdatedAt,
			"id":             item.ID,
		})
	}

	c.JSON(http.StatusOK, transformedNews)
}

func GetStorefront(c *gin.Context) {
	db := database.Get()

	var weeklyItems []fortnite.Catalog
	var dailyItems []fortnite.Catalog

	weeklyDone := make(chan bool)
	dailyDone := make(chan bool)
	var weeklyErr, dailyErr error

	go func() {
		weeklyErr = db.Where("storefront = ?", "BRWeeklyStorefront").Find(&weeklyItems).Error
		weeklyDone <- true
	}()

	go func() {
		dailyErr = db.Where("storefront = ?", "BRDailyStorefront").Find(&dailyItems).Error
		dailyDone <- true
	}()

	<-weeklyDone
	<-dailyDone

	if weeklyErr != nil || dailyErr != nil {
		utilities.Storefront.InvalidItem().Apply(c.Writer)
		return
	}

	itemsData, err := os.ReadFile("static/storefront/items.json")
	if err != nil {
		utilities.Storefront.InvalidItem().Apply(c.Writer)
		return
	}

	if len(itemsData) >= 3 && itemsData[0] == 0xEF && itemsData[1] == 0xBB && itemsData[2] == 0xBF {
		itemsData = itemsData[3:]
	}

	var allItems []map[string]interface{}
	if err := json.Unmarshal(itemsData, &allItems); err != nil {
		utilities.Storefront.InvalidItem().Apply(c.Writer)
		return
	}

	findRarity := func(name string) string {
		if name == "" {
			return "none"
		}

		for _, item := range allItems {
			itemName, ok := item["name"].(string)
			if !ok {
				continue
			}

			if itemName == name {
				if rarity, ok := item["rarity"]; ok {
					switch r := rarity.(type) {
					case string:
						return r
					case map[string]interface{}:
						if value, ok := r["value"].(string); ok {
							return value
						}
					}
				}
				return "none"
			}
		}
		return "none"
	}

	formatEntries := func(entries []fortnite.Catalog) []map[string]interface{} {
		result := make([]map[string]interface{}, 0, len(entries))

		for _, entry := range entries {
			var priceData map[string]interface{}
			regularPrice := 0

			if err := json.Unmarshal([]byte(entry.Data), &priceData); err == nil {
				if prices, ok := priceData["Prices"].([]interface{}); ok && len(prices) > 0 {
					if price, ok := prices[0].(map[string]interface{}); ok {
						if rp, ok := price["RegularPrice"].(float64); ok {
							regularPrice = int(rp)
						}
					}
				}
			}

			id := strings.ToLower(entry.TemplateId)
			id = strings.Replace(id, "athenacharacter:", "", -1)
			id = strings.Replace(id, "athenabackpack:", "", -1)
			id = strings.Replace(id, "athenapickaxe:", "", -1)
			id = strings.Replace(id, "athenaglider:", "", -1)
			id = strings.Replace(id, "athenadance:", "", -1)
			id = strings.Replace(id, "athenaitemwrap:", "", -1)
			id = strings.Replace(id, "athenamusicpack:", "", -1)

			rarity := findRarity(entry.Name)

			imageURL := fmt.Sprintf("https://fortnite-api.com/images/cosmetics/br/%s/featured.png", id)

			item := map[string]interface{}{
				"offerId":    entry.OfferId,
				"name":       entry.Name,
				"templateId": id,
				"category":   entry.Category,
				"rarity":     rarity,
				"price":      regularPrice,
				"image":      imageURL,
			}

			result = append(result, item)
		}

		return result
	}

	tomorrowAtMidnight := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 2)

	response := map[string]interface{}{
		"Refresh": tomorrowAtMidnight.Format(time.RFC3339),
		"Storefront": []map[string]interface{}{
			{
				"name":    "Featured",
				"Entries": formatEntries(weeklyItems),
			},
			{
				"name":    "Daily",
				"Entries": formatEntries(dailyItems),
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

func GetLauncherFiles(c *gin.Context) {
	db := database.Get()
	var files []solaris.Files

	result := db.Find(&files)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}

	c.JSON(http.StatusOK, files)
}

func GetLauncherTrailer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"url": "https://www.youtube.com/embed/Oy_sdUZmJGA?si=qSpBOjO2sSwcbp1J",
	})
}

func GetLauncherUpdate(c *gin.Context) {
	db := database.Get()

	var updater solaris.Updater

	result := db.Raw(`
		SELECT version, pub_date as PublishDate, url, signature, notes
		FROM launcher_updater
		WHERE version = (SELECT MAX(version) FROM launcher_updater)
	`).Scan(&updater)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Download"})
		return
	}

	if updater.Version == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Launcher Download not found"})
		return
	}

	c.Redirect(http.StatusFound, updater.Url)
}
