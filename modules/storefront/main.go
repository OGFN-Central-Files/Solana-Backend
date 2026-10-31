package storefront

import (
	"fmt"
	"os"
	"strings"
	"time"

	"solana/modules/database"
	"solana/modules/database/entities/fortnite"
	"solana/modules/storefront/declarations"
	"solana/modules/storefront/generation"
	"solana/modules/storefront/scraper"
	"solana/modules/storefront/utils"
	"solana/utilities"

	"github.com/fatih/color"
)

var currentShopDate time.Time

func Init() {
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) > 0 && argsWithoutProg[0] == "now" {
		processStorefront()
		return
	}

	go waitUntilEndOfDay()
}

func waitUntilEndOfDay() {
	for {
		now := time.Now().UTC()
		currentDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := currentDay.AddDate(0, 0, 1)

		sleepDuration := endOfDay.Sub(now)
		if sleepDuration < 0 {
			continue
		}

		utilities.LogWithTimestamp(color.GreenString, fmt.Sprintf("Sleeping for %v", sleepDuration))

		time.Sleep(sleepDuration)
		processStorefront()

		utilities.LogInfo("Storefront processed, waiting for the next cycle.")
	}
}

func processStorefront() {
	db := database.Get()
	lastShop := fortnite.GetLastProcessedShop(db)

	if !lastShop.IsZero() {
		currentShopDate = lastShop.AddDate(0, 0, 1)
	} else {
		currentShopDate = time.Date(2019, 8, 1, 0, 0, 0, 0, time.UTC)
	}

	stopDate := time.Date(2019, 10, 10, 23, 59, 59, 0, time.UTC)
	now := time.Now()

	if currentShopDate.Before(stopDate) && currentShopDate.Before(now) {
		if !fortnite.IsShopProcessed(db, currentShopDate) {
			handleStorefront(currentShopDate)
			fortnite.SaveProcessedShop(db, currentShopDate)
			utilities.LogWithTimestamp(color.GreenString, "Generated storefront for %s",
				currentShopDate.Format("January-2-2006"))
		} else {
			utilities.LogInfo(fmt.Sprintf("Shop for %s already processed, skipping",
				currentShopDate.Format("January-2-2006")))
		}
	}
}

func handleStorefront(date time.Time) {
	db := database.Get()
	db.Exec("DELETE FROM catalog WHERE storefront = ?", "BRWeeklyStorefront")
	db.Exec("DELETE FROM catalog WHERE storefront = ?", "BRDailyStorefront")
	db.Exec("DELETE FROM catalog WHERE storefront = ?", "BRSeason10")
	db.Exec("DELETE FROM shopsections")

	formattedDate := strings.ToLower(date.Format("January-2-2006"))
	url := fmt.Sprintf("https://fnbr.co/shop/%s", formattedDate)

	utilities.LogWithTimestamp(color.CyanString, "Generating storefront for "+formattedDate)

	items := utils.LoadItems("https://fortnite-api.com/v2/cosmetics/br")
	if items == nil {
		utilities.LogWithTimestamp(color.RedString, "Failed to load items for "+formattedDate)
		return
	}

	result := scraper.GetStorefront(url, items)
	if result == nil {
		utilities.LogWithTimestamp(color.RedString, "Failed to scrape data for "+formattedDate)
		return
	}

	utils.Mu.Lock()
	utils.StorefrontData = result
	utils.Mu.Unlock()

	dailySection := declarations.CreateSection("BRDailyStorefront")
	weeklySection := declarations.CreateSection("BRWeeklyStorefront")
	BRSeason := declarations.CreateSection("BRSeason10")

	generation.FillDailyStorefront(&dailySection)
	generation.FillWeeklyStorefront(&weeklySection)
	generation.FillBRSeason10Storefront(BRSeason)
	generation.FillAllSectionsInDatabase()

	utilities.LogWithTimestamp(color.GreenString, "Finished generating storefront for "+formattedDate)
}
