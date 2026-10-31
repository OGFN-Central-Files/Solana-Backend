package storefront

type Catalog struct {
	RefreshIntervalHrs int                 `json:"refreshIntervalHrs"`
	DailyPurchaseHrs   int                 `json:"dailyPurchaseHrs"`
	Expiration         string              `json:"expiration"`
	Storefronts        []DefaultStorefront `json:"storefronts"`
}

type CatalogEntry struct {
	OfferId              string        `json:"offerId"`
	OfferType            string        `json:"offerType"`
	DevName              string        `json:"devName"`
	ItemGrants           []ItemGrant   `json:"itemGrants"`
	Requirements         []Requirement `json:"requirements"`
	Categories           []string      `json:"categories"`
	MetaInfo             []MetaInfo    `json:"metaInfo"`
	Meta                 Meta          `json:"meta"`
	GiftInfo             GiftInfo      `json:"giftInfo"`
	Prices               []Price       `json:"prices"`
	DisplayAssetPath     string        `json:"displayAssetPath"`
	NewDisplayAssetPath  string        `json:"newDisplayAssetPath"`
	Refundable           bool          `json:"refundable"`
	Title                string        `json:"title"`
	Description          string        `json:"description"`
	ShortDescription     string        `json:"shortDescription"`
	AppStoreId           []string      `json:"appStoreId"`
	FulfillmentIds       []string      `json:"fulfillmentIds"`
	DailyLimit           int           `json:"dailyLimit"`
	WeeklyLimit          int           `json:"weeklyLimit"`
	MonthlyLimit         int           `json:"monthlyLimit"`
	SortPriority         int           `json:"sortPriority"`
	CatalogGroupPriority int           `json:"catalogGroupPriority"`
	FilterWeight         int           `json:"filterWeight"`
	BannerOverride       string        `json:"bannerOverride"`
	MatchFilter          string        `json:"matchFilter"`
	AdditionalGrants     []interface{} `json:"additionalGrants"`
}

type DefaultStorefront struct {
	Name           string         `json:"name"`
	CatalogEntries []CatalogEntry `json:"catalogEntries"`
}

type GiftInfo struct {
	BIsEnabled              bool          `json:"bIsEnabled"`
	ForcedGiftBoxTemplateId string        `json:"forcedGiftBoxTemplateId"`
	PurchaseRequirements    []Requirement `json:"purchaseRequirements"`
	GiftRecordIds           []string      `json:"giftRecordIds"`
}

type ItemGrant struct {
	TemplateId string `json:"templateId"`
	Quantity   int    `json:"quantity"`
}

type Meta struct {
	NewDisplayAssetPath  string `json:"newDisplayAssetPath"`
	DisplayAssetPath     string `json:"displayAssetPath"`
	SectionId            string `json:"sectionId"`
	TileSize             string `json:"tileSize"`
	LayoutId             string `json:"layoutId"`
	AnalyticOfferGroupId string `json:"analyticOfferGroupId"`
}

type MetaInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Price struct {
	CurrencyType        string `json:"currencyType"`
	CurrencySubType     string `json:"currencySubType"`
	RegularPrice        int    `json:"regularPrice"`
	FinalPrice          int    `json:"finalPrice"`
	BasePrice           int    `json:"basePrice"`
	SaleExpiration      string `json:"saleExpiration"`
	DynamicRegularPrice int    `json:"dynamicRegularPrice"`
}

type Requirement struct {
	RequirementType string `json:"requirementType"`
	RequiredId      string `json:"requiredId"`
	MinQuantity     int    `json:"minQuantity"`
}
