package creative

import "github.com/gin-gonic/gin"

func GetLinkViaMnemonic(c *gin.Context) {
	response := map[string]interface{}{
		"namespace":   "namespace",
		"accountId":   "1a1a05b96bd94042b29a3407716f90af",
		"creatorName": "itztiva",
		"menmonic":    "1234-5678-9101",
		"linkType":    "linkType",
		"metadata": map[string]interface{}{
			"quicksilver_id": "quicksilver_id",
			"image_url":      "https://cdn2.unrealengine.com/fneco-28-00-creative-discovertiles-keyart-1920x1080-1920x1080-51c98fe78d7a.jpg",
			"image_urls": map[string]string{
				"url_s": "https://cdn2.unrealengine.com/fneco-28-00-creative-discovertiles-keyart-1920x1080-1920x1080-51c98fe78d7a.jpg",
				"url_m": "https://cdn2.unrealengine.com/fneco-28-00-creative-discovertiles-keyart-1920x1080-1920x1080-51c98fe78d7a.jpg",
				"url":   "https://cdn2.unrealengine.com/fneco-28-00-creative-discovertiles-keyart-1920x1080-1920x1080-51c98fe78d7a.jpg",
			},
			"title":  "Tiva is the best",
			"locale": "locale",
			"matchmaking": map[string]interface{}{
				"selectedJoinInProgressType": 980,
				"playersPerTeam":             -1,
				"maximumNumberOfPlayers":     32,
				"override_playlist":          "",
				"playerCount":                32,
				"mmsType":                    "keep_full",
				"mmsPrivacy":                 "Public",
				"numberOfTeams":              16,
				"bAllowJoinInProgress":       true,
				"minimumNumberOfPlayers":     32,
				"joinInProgressTeam":         255,
			},
			"generated_image_urls": map[string]interface{}{
				"url_s": "https://cdn2.unrealengine.com/fneco-28-00-creative-discovertiles-keyart-1920x1080-1920x1080-51c98fe78d7a.jpg",
				"url_m": "https://cdn2.unrealengine.com/fneco-28-00-creative-discovertiles-keyart-1920x1080-1920x1080-51c98fe78d7a.jpg",
				"compressed": map[string]string{
					"url_s": "https://cdn2.unrealengine.com/fneco-28-00-creative-discovertiles-keyart-1920x1080-1920x1080-51c98fe78d7a.jpg",
					"url_m": "https://cdn2.unrealengine.com/fneco-28-00-creative-discovertiles-keyart-1920x1080-1920x1080-51c98fe78d7a.jpg",
					"url":   "https://cdn2.unrealengine.com/fneco-28-00-creative-discovertiles-keyart-1920x1080-1920x1080-51c98fe78d7a.jpg",
				},
				"url": "url",
			},
			"mode":       "live",
			"islandType": "CreativePlot:flatgrid_large",
			"dynamicXp": map[string]interface{}{
				"uniqueGameVersion": 138,
				"calibrationPhase":  "LiveXP",
			},
			"tagline":      "Tiva is the best",
			"supportCode":  "bio",
			"projectId":    "9b20452c-1659-4583-852e-bd84f973c967",
			"introduction": "Tiva is so cool!",
		},
		"version":          138,
		"active":           true,
		"disabled":         false,
		"created":          "2023-05-06T02:13:43.780Z",
		"published":        "2023-05-06T02:13:43.780Z",
		"descriptionTags":  []string{"4v4", "practice", "zonewars"},
		"moderationStatus": "Approved",
		"lastActiveDate":   "2023-05-06T02:17:34.231Z",
		"discoveryIntent":  "PUBLIC",
	}

	c.JSON(200, response)
}
