package scraper

import (
	"log"
	"net/http"
	"solana/modules/storefront/models"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Item struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	Price    int    `json:"price"`
	Category string `json:"category"`
	Set      struct {
		Value        string `json:"value"`
		Text         string `json:"text"`
		BackendValue string `json:"backendValue"`
	} `json:"set"`
}

func GetStorefront(url string, items map[string]models.Item) map[string][]Item {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	shopData := make(map[string][]Item)
	currentSection := ""
	seenItems := make(map[string]bool)

	doc.Find(".shop-section-title, .item-responsive").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("shop-section-title") {
			currentSection = strings.TrimSpace(strings.Replace(s.Text(), " Items", "", -1))
			shopData[currentSection] = []Item{}
			seenItems = make(map[string]bool)
		} else if currentSection != "" {
			itemName := strings.TrimSpace(s.Find(".item-name span").Text())
			priceSelection := s.Find(".item-price")
			if priceSelection.Length() == 0 {
				log.Printf("No price found for item: %s", itemName)
				return
			}

			itemPriceText := strings.Map(func(r rune) rune {
				if r >= '0' && r <= '9' {
					return r
				}
				return -1
			}, priceSelection.Text())

			if itemPriceText == "" {
				log.Printf("Price extraction failed for item: %s", itemName)
				return
			}

			itemPrice, err := strconv.Atoi(itemPriceText)
			if err != nil {
				log.Printf("Failed to convert price for %s (raw: %s): %v", itemName, itemPriceText, err)
				return
			}

			if itemName != "" {
				if item, exists := items[itemName]; exists {
					if seenItems[itemName] {
						log.Printf("Duplicate item found in section %s: %s", currentSection, itemName)
						return
					}
					seenItems[itemName] = true

					shopData[currentSection] = append(shopData[currentSection], Item{
						Name:     itemName,
						ID:       item.ID,
						Price:    itemPrice,
						Category: item.Set.Value,
						Set:      item.Set,
					})
				}
			}
		}
	})

	return shopData
}
