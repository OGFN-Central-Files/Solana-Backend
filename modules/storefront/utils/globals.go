package utils

import (
	"solana/modules/storefront/scraper"
	"sync"
)

var (
	StorefrontData map[string][]scraper.Item
	Mu             sync.Mutex
)
