package fortnite

import (
	"net/http"
	"solana/modules/database"
	"solana/modules/database/entities/contentpages"
	"solana/modules/database/entities/fortnite"
	"solana/utilities"
	"strconv"

	"github.com/gin-gonic/gin"
)

func FortniteCPHandler(c *gin.Context) {
	ua := utilities.Parse(c.GetHeader("User-Agent"))

	if ua == nil {
		utilities.Basic.BadRequest().Apply(c.Writer)
		return
	}
	seasonStr := strconv.Itoa(ua.Season)
	backgroundStage := "season" + seasonStr
	backgroundImage := ""

	switch ua.Season {
	case 10:
		if ua.Build == "10.40" {
			backgroundStage = "blackmonday"
		} else {
			backgroundStage = "seasonx"
		}
	case 11:
		if ua.Build == "11.10" {
			backgroundStage = "fortnitemares"
		} else if ua.Build == "11.30" {
			backgroundStage = "Galileo"
		} else if ua.Build == "11.31" || ua.Build == "11.40" {
			backgroundStage = "Winter19"
		} else {
			backgroundStage = "season11"
		}
	case 12:
		backgroundStage = "season12"
	case 13:
		backgroundStage = "season13"
	case 14:
		if ua.Build == "14.40" {
			backgroundStage = "halloween2020"
		} else {
			backgroundStage = "season14"
		}
	case 15:
		if ua.Build == "15.10" {
			backgroundStage = "season15xmas"
		} else {
			backgroundStage = "season15"
		}
	case 16:
		backgroundStage = "season16"
	case 17:
		if ua.Build == "17.10" {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/t-bp-17-lobby-summer-2048x1024-709fa99e6be0.png"
		} else if ua.Build == "17.21" {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/t-bp17-21-lobby-2048x1024-f6027bf109de.png"
		} else if ua.Build == "17.40" {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/t-bp17-40-lobby-2048x1024-f742fc604aae.png"
		}
	case 18:
		backgroundStage = "season18"
	case 19:
		if ua.Build == "19.01" {
			backgroundStage = "winter2021"
			backgroundImage = "https://cdn2.unrealengine.com/t-bp19-lobby-xmas-2048x1024-f85d2684b4af.png"
		} else if ua.Build == "19.10" {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/fortnite-tilted-towers-1920x1080-ad94e5f0b016.jpg"
		} else {
			backgroundStage = "season19"
		}
	case 20:
		if ua.Build == "20.10" {
			backgroundStage = "season20"
			backgroundImage = "https://cdn2.unrealengine.com/t-bp20-lobby-2048x1024-d89eb522746c.png"
		} else if ua.Build == "20.40" {
			backgroundStage = "season20"
			backgroundImage = "https://cdn2.unrealengine.com/t-bp20-40-armadillo-glowup-lobby-2048x2048-2048x2048-3b83b887cc7f.jpg"
		} else {
			backgroundStage = "season20"
			backgroundImage = "https://cdn2.unrealengine.com/s20-landscapev4-2048x1024-2494a103ae6c.png"
		}
	case 21:
		if ua.Build == "21.30" {
			backgroundStage = "season2130"
			backgroundImage = "https://cdn2.unrealengine.com/nss-lobbybackground-2048x1024-f74a14565061.jpg"
		} else {
			backgroundStage = "season2100"
			backgroundImage = "https://cdn2.unrealengine.com/s21-lobby-background-2048x1024-2e7112b25dc3.jpg"
		}
	case 22:
		if ua.Build == "22.20" {
			backgroundStage = "season2220"
			backgroundImage = "https://cdn2.unrealengine.com/t-bp22-fortnitemares-lobby-square-2048x2048-2048x2048-3b7cda3aa517.jpg"
		} else {
			backgroundStage = "season2200"
			backgroundImage = "https://cdn2.unrealengine.com/t-bp22-lobby-square-2048x2048-2048x2048-e4e90c6e8018.jpg"
		}
	case 23:
		if ua.Build == "23.10" {
			backgroundStage = "season2310"
			backgroundImage = "https://cdn2.unrealengine.com/t-bp23-winterfest-lobby-square-2048x2048-2048x2048-277a476e5ca6.png"
		} else if ua.Build == "23.40" {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/mostwanted-final-v2-2048x2048-2048x2048-39f2b5041a40.jpg"
		} else {
			backgroundStage = "season2300"
			backgroundImage = "https://cdn2.unrealengine.com/t-bp23-lobby-2048x1024-2048x1024-26f2c1b27f63.png"
		}
	case 24:
		if ua.Build == "24.30" {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/ch4s2-lobbyupdate-4-20-2022-lifted-copy-3840x2160-d3a138f5f9e7.jpg"
		} else {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/t-ch4s2-bp-lobby-4096x2048-edde08d15f7e.jpg"
		}
	case 25:
		if ua.Build == "25.11" {
			backgroundStage = "season2500"
			backgroundImage = "https://cdn2.unrealengine.com/t-s25-14dos-lobby-4096x2048-2be24969eee3.jpg"
		} else {
			backgroundStage = "season2500"
			backgroundImage = "https://cdn2.unrealengine.com/s25-lobby-4k-4096x2048-4a832928e11f.jpg"
		}
	case 26:
		if ua.Build == "26.30" {
			backgroundStage = "season2630"
			backgroundImage = "https://cdn2.unrealengine.com/s26-lobby-timemachine-final-2560x1440-a3ce0018e3fa.jpg"
		} else {
			backgroundStage = "season2600"
			backgroundImage = "https://cdn2.unrealengine.com/0814-ch4s4-lobby-2048x1024-2048x1024-e3c2cf8d342d.png"
		}
	case 27:
		if ua.Build == "27.11" {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/durianlobby2-4096x2048-242a51b6a8ee.jpg"
		} else {
			backgroundStage = "rufus"
		}
	case 28:
		if ua.Build == "28.01" {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/winterfest2023-lobby-2048x1024-a8853c3a6f59.jpg"
		} else if ua.Build == "28.20" {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/s28-tmnt-lobby-4096x2048-e6c06a310c05.jpg"
		} else {
			backgroundImage = "https://cdn2.unrealengine.com/ch5s1-lobbybg-3640x2048-0974e0c3333c.jpg"
			backgroundStage = "defaultnotris"
		}
	case 29:
		if ua.Build == "29.20" {
			backgroundStage = "season2920"
			backgroundImage = "https://cdn2.unrealengine.com/iceberg-lobby-3840x2160-217bb6ea8af9.jpg"
		} else if ua.Build == "29.40" {
			backgroundStage = "defaultnotris"
			backgroundImage = "https://cdn2.unrealengine.com/mkart-2940-sw-fnbr-lobby-3840x2160-4f1f1486a54a.jpg"
		} else {
			backgroundStage = "defaultnotris"
		}
	}

	var tournamentInfo = []contentpages.TournamentInfo{}
	var playlistInfo = []contentpages.PlaylistInfo{}
	var shopSections = []fortnite.ShopSections{}

	db := database.Get()

	db.Find(&tournamentInfo)
	db.Find(&playlistInfo)
	db.Find(&shopSections)

	filteredTournaments := []interface{}{}
	for _, t := range tournamentInfo {
		tournament := map[string]interface{}{
			"_type":                  "Tournament Display Info",
			"title_color":            t.TitleColor,
			"loading_screen_image":   t.LoadingScreenImage,
			"background_text_color":  t.BackgroundTextColor,
			"background_right_color": t.BackgroundRightColor,
			"poster_back_image":      t.PosterBackImage,
			"type":                   t.Type,
			"pin_score_requirement":  t.PinScoreRequirement,
			"pin_earned_text":        t.PinEarnedText,
			"tournament_display_id":  t.TournamentDisplayId,
			"highlight_color":        t.HighlightColor,
			"schedule_info":          t.ScheduleInfo,
			"primary_color":          t.PrimaryColor,
			"flavor_description":     t.FlavorDescription,
			"poster_front_image":     t.PosterFrontImage,
			"short_format_title":     t.ShortFormatTitle,
			"title_line_2":           t.TitleLine2,
			"title_line_1":           t.TitleLine1,
			"shadow_color":           t.ShadowColor,
			"details_description":    t.DetailsDescription,
			"background_left_color":  t.BackgroundLeftColor,
			"long_format_title":      t.LongFormatTitle,
			"poster_fade_color":      t.PosterFadeColor,
			"secondary_color":        t.SecondaryColor,
			"playlist_tile_image":    t.PlaylistTileImage,
			"base_color":             t.BaseColor,
		}
		filteredTournaments = append(filteredTournaments, tournament)
	}

	filteredPlaylists := []interface{}{}
	for _, p := range playlistInfo {
		playlist := map[string]interface{}{
			"_type":          "FortPlaylistInfo",
			"image":          p.Image,
			"playlist_name":  p.PlaylistName,
			"special_border": p.SpecialBorder,
		}
		filteredPlaylists = append(filteredPlaylists, playlist)
	}

	filteredShopSections := []interface{}{}
	for _, s := range shopSections {
		section := map[string]interface{}{
			"_type":                           "ShopSection",
			"bEnableToastNotification":        true,
			"bHidden":                         false,
			"bShowIneligibleOffers":           true,
			"bShowIneligibleOffersIfGiftable": true,
			"bShowTimer":                      true,
			"bSortOffersByOwnership":          false,
			"background": map[string]interface{}{
				"_type": "DynamicBackground",
				"key":   "vault",
				"stage": "default",
			},
			"landingPriority":    2,
			"sectionDisplayName": s.Section,
			"sectionId":          s.Section,
		}
		filteredShopSections = append(filteredShopSections, section)
	}

	contentPages := map[string]interface{}{
		"_title":       "Fortnite Game",
		"_activeDate":  "0001-01-01T00:00:00",
		"lastModified": "0001-01-01T00:00:00",
		"_locale":      "en-US",
		"battleroyalenews": map[string]interface{}{
			"_activeDate":  "0001-01-01T00:00:00",
			"_locale":      "en-US",
			"_title":       "battleroyalenews",
			"lastModified": "0001-01-01T00:00:00",
			"news": map[string]interface{}{
				"_type": "Battle Royale News",
				"messages": []map[string]interface{}{
					{
						"image":     "https://cdn2.unrealengine.com/Fortnite%2Fblog%2Fv10-00-patch-notes%2F10BR_Mech_PatchNotes-1920x1080-30ada9e73d283d3dbdbf73bd86738e597aa939d4.jpg",
						"hidden":    false,
						"_type":     "CommonUI Simple Message MOTD",
						"title":     "Welcome to Solaris!",
						"body":      "We are currently in Development. If there are any issues, please report the issue to our discord.",
						"spotlight": false,
					},
					{
						"image":     "https://backend-overseas-project.xyz/SolarisDonatorNews.png",
						"hidden":    false,
						"_type":     "CommonUI Simple Message MOTD",
						"title":     "Donate to Solaris",
						"body":      "By donating to Solaris, you will recieve perks in-game and it will help us provide better servers! Donate at: https://solaris-fn.mysellix.io/",
						"spotlight": false,
					},
				},
				"motds": []map[string]interface{}{
					{
						"entryType":             "Text",
						"image":                 "https://cdn2.unrealengine.com/Fortnite%2Fblog%2Fv10-00-patch-notes%2F10BR_Mech_PatchNotes-1920x1080-30ada9e73d283d3dbdbf73bd86738e597aa939d4.jpg",
						"tileImage":             "https://cdn.solarisfn.org/Untitled_design_7.png",
						"hidden":                false,
						"videoMute":             false,
						"_type":                 "CommonUI Simple Message MOTD",
						"title":                 "Welcome to Solaris!",
						"body":                  "We are currently in Development. If there are any issues, please report the issue to our discord.",
						"videoLoop":             false,
						"videoStreamingEnabled": false,
						"id":                    "SolarisNewsPost0",
						"videoAutoplay":         false,
						"videoFullscreen":       false,
						"spotlight":             false,
					},
					{
						"entryType":             "Text",
						"image":                 "https://backend-overseas-project.xyz/SolarisDonatorNews.png",
						"tileImage":             "https://backend-overseas-project.xyz/SolarisDonatorNewsTile.png",
						"hidden":                false,
						"videoMute":             false,
						"_type":                 "CommonUI Simple Message MOTD",
						"title":                 "Donate to Solaris",
						"body":                  "By donating to Solaris, you will recieve perks in-game and it will help us provide better servers! Donate at: https://solaris-fn.mysellix.io/",
						"videoLoop":             false,
						"videoStreamingEnabled": false,
						"id":                    "SolarisNewsPost1",
						"videoAutoplay":         false,
						"videoFullscreen":       false,
						"spotlight":             false,
					},
				},
				"backgrounds": nil,
			},
		},
		"shopSections": map[string]interface{}{
			"_title": "shop-sections",
			"sectionList": map[string]interface{}{
				"_type":    "ShopSectionList",
				"sections": filteredShopSections,
			},
			"_noIndex":      false,
			"_activeDate":   "0001-12-01T21:00:00.000Z",
			"lastModified":  "0001-12-01T21:00:00.089Z",
			"_locale":       "en-US",
			"_templateName": "FortniteGameShopSections",
		},
		"subgameinfo": map[string]interface{}{
			"_activeDate": "0001-01-01T00:00:00",
			"_locale":     "en-US",
			"_title":      "SubgameInfo",
			"battleroyale": map[string]interface{}{
				"_type":       "Subgame Info",
				"color":       "5b2569",
				"description": "100 Player PvP",
				"image":       "",
				"subgame":     "battleroyale",
				"title":       "Battle Royale",
			},
			"creative": map[string]interface{}{
				"_type":       "Subgame Info",
				"color":       "0658b9",
				"description": "Your Islands. Your Friends. Your Rules.",
				"image":       "",
				"subgame":     "creative",
				"title":       "Creative",
			},
			"lastModified": "0001-01-01T00:00:00",
			"savetheworld": map[string]interface{}{
				"_type":       "Subgame Info",
				"color":       "7615E9FF",
				"description": "Cooperative PvE Adventure",
				"image":       "",
				"subgame":     "savetheworld",
				"title":       "Save The World",
			},
		},
		"mpItemShop": map[string]interface{}{
			"shopData": map[string]interface{}{
				"_type": "MP Item Shop - Data Root",
				"sections": []map[string]interface{}{
					{
						"metadata": map[string]interface{}{
							"offerGroups": []map[string]interface{}{
								{
									"bUseWidePreview": false,
									"_type":           "MP Item Shop - Row",
									"offerGroupId":    "99",
									"stackRanks": []map[string]interface{}{
										{
											"stackRankValue": 99,
											"_type":          "MP Item Shop - Stack Rank",
											"context":        "battleRoyale",
											"contentType":    "battleRoyale",
											"startDate":      "2023-01-01T00:00:00.000Z",
										},
									},
								},
							},
							"background": map[string]interface{}{
								"_type":         "MP Item Shop - Background",
								"customTexture": "",
							},
							"_type":                "MP Item Shop - Section Metadata",
							"showIneligibleOffers": "Always",
							"stackRanks": []map[string]interface{}{
								{
									"stackRankValue": 79,
									"_type":          "MP Item Shop - Stack Rank",
									"context":        "battleRoyale",
									"contentType":    "battleRoyale",
									"startDate":      "2023-01-01T00:00:00.000Z",
								},
							},
						},
						"displayName": "Item Shop",
						"_type":       "MP Item Shop - Section",
						"sectionID":   "Featured",
						"category":    "Item Shop",
					},
				},
			},
			"_title":       "mpItemShop",
			"_noIndex":     false,
			"_activeDate":  "2024-12-26T00:00:00.000Z",
			"lastModified": "2024-12-26T00:00:00.000Z",
			"_locale":      "en-US",
		},
		"dynamicbackgrounds": map[string]interface{}{
			"backgrounds": map[string]interface{}{
				"backgrounds": []map[string]interface{}{
					{
						"stage":           backgroundStage,
						"backgroundimage": backgroundImage,
						"_type":           "DynamicBackground",
						"key":             "lobby",
					},
					{
						"stage": backgroundStage,
						"_type": "DynamicBackground",
						"key":   "vault",
					},
				},
				"_type": "DynamicBackgroundList",
			},
			"_title":       "dynamicbackgrounds",
			"_noIndex":     false,
			"_activeDate":  "2019-08-21T15:59:59.342Z",
			"lastModified": "0001-01-01T00:00:00",
			"_locale":      "en-US",
		},
		"lobby": map[string]interface{}{
			"backgroundimage": backgroundImage,
			"stage":           "seasonx",
			"_title":          "lobby",
			"_activeDate":     "2019-08-21T15:59:59.342Z",
			"lastModified":    "0001-01-01T00:00:00",
			"_locale":         "en-US",
		},
		"playlistinformation": map[string]interface{}{
			"is_tile_hidden":                    true,
			"frontend_matchmaking_header_style": "Basic",
			"conversion_config": map[string]interface{}{
				"containerName":    "playlist_info",
				"_type":            "Conversion Config",
				"enableReferences": true,
				"contentName":      "playlists",
			},
			"show_ad_violator": false,
			"_title":           "playlistinformation",
			"playlist_info": map[string]interface{}{
				"_type":     "Playlist Information",
				"playlists": filteredPlaylists,
			},
			"_noIndex":     false,
			"_activeDate":  "0001-01-01T00:00:00",
			"lastModified": "0001-01-01T00:00:00",
			"_locale":      "en-US",
		},
		"tournamentinformation": map[string]interface{}{
			"conversion_config": map[string]interface{}{
				"containerName":    "tournament_info",
				"_type":            "Conversion Config",
				"enableReferences": true,
				"contentName":      "tournaments",
			},
			"tournament_info": map[string]interface{}{
				"tournaments": filteredTournaments,
				"_type":       "Tournaments Info",
			},
			"_title":       "tournamentinformation",
			"_noIndex":     false,
			"_activeDate":  "0001-01-01T00:00:00",
			"lastModified": "0001-01-01T00:00:00",
			"_locale":      "en-US",
		},
		"emergencynotice": map[string]interface{}{
			"news": map[string]interface{}{
				"_type": "Battle Royale News",
				"messages": []map[string]interface{}{
					{
						"hidden":    false,
						"_type":     "CommonUI Simple Message Base",
						"title":     "Solaris",
						"body":      "Relive OG Fortnite Season X with Solaris\ndiscord.gg/solarisfn",
						"spotlight": true,
					},
				},
			},
			"_title":       "emergencynotice",
			"_noIndex":     false,
			"alwaysShow":   false,
			"_activeDate":  "2018-08-06T19:00:26.217Z",
			"lastModified": "2019-10-29T22:32:52.686Z",
			"_locale":      "en-US",
		},
		"emergencynoticev2": map[string]interface{}{
			"jcr_isCheckedOut": true,
			"_title":           "emergencynoticev2",
			"_noIndex":         false,
			"jcr_baseVersion":  "a7ca237317f1e7da533b38-74ee-468b-8c63-a7c3c256b313",
			"emergencynotices": map[string]interface{}{
				"_type": "Emergency Notices",
				"emergencynotices": []map[string]interface{}{
					{
						"hidden": false,
						"_type":  "CommonUI Emergency Notice Base",
						"title":  "",
						"body":   "",
					},
				},
			},
			"_activeDate":  "2018-08-06T19:00:26.217Z",
			"lastModified": "2021-06-22T08:27:47.969Z",
			"_locale":      "en-US",
		},
	}

	c.JSON(http.StatusOK, contentPages)
}
