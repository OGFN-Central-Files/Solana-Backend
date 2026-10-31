package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"solana/modules/storefront/models"
	"solana/utilities"
	"time"

	"github.com/fatih/color"
)

type ApiResponse struct {
	Status int           `json:"status"`
	Data   []models.Item `json:"data"`
}

func LoadItems(apiURL string) map[string]models.Item {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(apiURL)
	if err != nil {
		log.Panicf("Failed to fetch items from API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Panicf("API request failed with status: %d", resp.StatusCode)
	}

	var apiResponse ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		log.Panicf("Failed to parse API response: %v", err)
	}

	itemCount := len(apiResponse.Data)
	itemMap := make(map[string]models.Item, itemCount)

	for i := range apiResponse.Data {
		item := &apiResponse.Data[i]
		if item.Name != "" && item.ID != "" {
			itemMap[item.Name] = *item
		}
	}

	utilities.LogWithTimestamp(color.GreenString, "Loaded %d items from API", len(itemMap))
	return itemMap
}

func SaveJSON(filepath string, data interface{}) {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	err = os.WriteFile(filepath, file, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}
}
