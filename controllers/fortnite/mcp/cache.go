package mcp

import (
	"encoding/json"
	"os"
	"solana/classes/mcp"
	"sync"
	"sync/atomic"
)

type Item struct {
	AccountID  string                 `json:"accountId"`
	ProfileID  string                 `json:"profileid"`
	TemplateID string                 `json:"templateid"`
	Value      map[string]interface{} `json:"value"`
	Quantity   int                    `json:"quantity"`
	IsStat     bool                   `json:"isStat"`
}

var (
	itemsCache sync.Map
	isLoaded   atomic.Bool
)

func LoadCache() {
	data, _ := os.ReadFile("static/athena/allitems.json")

	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return
	}

	for _, item := range items {
		baseItem := mcp.BaseItem{
			Attributes: item.Value,
			TemplateId: item.TemplateID,
			Quantity:   item.Quantity,
		}
		itemsCache.Store(item.TemplateID, baseItem)
	}
	isLoaded.Store(true)
}

func GetCachedItems() *sync.Map {
	return &itemsCache
}

func IsCacheLoaded() bool {
	return isLoaded.Load()
}
