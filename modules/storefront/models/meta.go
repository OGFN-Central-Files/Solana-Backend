package models

type MetaInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Meta struct {
	NewDisplayAssetPath  string
	LayoutId             string
	TileSize             string
	AnalyticOfferGroupId string
	SectionId            string
	TemplateId           string `json:"templateId"`
	DisplayAssetPath     string `json:"displayAssetPath"`
}
