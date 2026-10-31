package dsmcp

import (
	"encoding/json"
	"log"
	"net/http"
	"solana/classes/battlepass"
	"solana/classes/mcp"
	"solana/modules/caching"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/modules/database/entities/fortnite"
	"solana/utilities"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SendPlayerMatchStats(c *gin.Context) {
	var body struct {
		Player struct {
			DisplayName string   `json:"DisplayName"`
			Team        []string `json:"team"`
		} `json:"player"`
		DedicatedServer struct {
			Region              string `json:"Region"`
			Playlist            string `json:"Playlist"`
			SessionId           string `json:"SessionId"`
			StartingPlayerCount int    `json:"StartingPlayerCount"`
		} `json:"dedicatedServer"`
		AthenaProfile struct {
			ProfileRevision            int    `json:"ProfileRevision"`
			ProfileChangesBaseRevision int    `json:"ProfileChangesBaseRevision"`
			ProfileCommandRevision     int    `json:"ProfileCommandRevision"`
			ServerTime                 string `json:"ServerTime"`
			ResponseVersion            int    `json:"ResponseVersion"`
			ProfileChanges             []struct {
				UpdateType string `json:"updateType"`
				Stats      struct {
					XP int `json:"xp"`
				} `json:"stats"`
			} `json:"ProfileChanges"`
		} `json:"athenaProfile"`
		MatchStats struct {
			Eliminations int `json:"Eliminations"`
			Position     int `json:"Position"`
			Tournament   struct {
				Event  string `json:"EventId"`
				Window string `json:"Window"`
			} `json:"Tournament"`
		} `json:"matchStats"`
	}

	if err := c.BindJSON(&body); err != nil {
		utilities.MCP.InvalidPayload().
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	accountId := c.Param("accountId")

	if accountId == "" {
		utilities.MCP.ProfileNotFound().
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var response mcp.DefaultMCPResponse
	profileChanges := []map[string]interface{}{}

	db := database.Get()

	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", accountId).First(&user).Error; err != nil {
		utilities.MCP.ProfileNotFound().
			WithMessageVar([]string{accountId}).
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")

	var profile accounts.Profiles
	var items []accounts.Items
	var quests []accounts.Quests

	profileFound := true

	if err := db.Where("account_id = ? AND profile_id = ?", accountId, "athena").First(&profile).Error; err != nil {
		profileFound = false
	} else if profileFound {
		db.Where("account_id = ? AND profile_id = ?", accountId, "athena").Find(&quests)
		db.Where("account_id = ? AND profile_id = ?", accountId, "athena").Find(&items)
	}

	if !profileFound {
		response = mcp.DefaultMCPResponse{
			ProfileRevision:            0,
			ProfileId:                  "athena",
			ProfileChangesBaseRevision: 0,
			ProfileCommandRevision:     0,
			ServerTime:                 now,
			ResponseVersion:            1,
			ProfileChanges: []map[string]interface{}{
				{
					"changeType": "fullProfileUpdate",
					"profile": mcp.Profile{
						Created:    now,
						Updated:    now,
						Rvn:        0,
						WipeNumber: 1,
						AccountId:  accountId,
						ProfileId:  "athena",
						Version:    "no_version",
						Items:      make(map[string]interface{}),
						Stats: mcp.Stats{
							Attributes: make(map[string]interface{}),
						},
						CommandRevision: 0,
					},
				},
			},
		}
	} else {
		var XP int
		for _, change := range body.AthenaProfile.ProfileChanges {
			if change.UpdateType == "statsUpdate" {
				XP = change.Stats.XP
				break
			}
		}

		matchStats := fortnite.MatchStats{
			AccountId:        accountId,
			Kills:            body.MatchStats.Eliminations,
			Team:             body.Player.Team[0],
			Xp:               XP,
			ChallengeUpdates: "",
			Position:         body.MatchStats.Position,
			Playlist:         body.DedicatedServer.Playlist,
		}

		if err := db.Create(&matchStats).Error; err != nil {
			log.Printf("Error creating match stats: %v", err)
			utilities.MCP.InternalServerError().
				WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		var level accounts.Items
		var bookLevel accounts.Items
		var bookXp accounts.Items
		if err := db.Where("account_id = ? AND template_id = ?", accountId, "book_xp").First(&bookXp).Error; err != nil {
			log.Printf("Error fetching book_xp: %v", err)
			bookXp = accounts.Items{
				AccountID:  accountId,
				ProfileID:  "athena",
				TemplateID: "book_xp",
				Value:      "0",
				IsStat:     false,
				Quantity:   1,
				UUID:       uuid.New().String(),
			}
			if err := db.Create(&bookXp).Error; err != nil {
				log.Printf("Error creating book_xp: %v", err)
				utilities.MCP.InternalServerError().
					WithIntent(utilities.Prod).Apply(c.Writer)
				return
			}
		}

		var xp accounts.Items
		for _, item := range items {
			if item.TemplateID == "level" {
				level = item
			} else if item.TemplateID == "book_level" {
				bookLevel = item
			} else if item.TemplateID == "book_xp" {
				bookXp = item
			} else if item.TemplateID == "xp" {
				xp = item
			}
		}

		var currency accounts.Items
		if err := db.Where("account_id = ? AND template_id = ? AND profile_id = ?", accountId, "Currency:MtxPurchased", "common_core").First(&currency).Error; err != nil {
			utilities.MCP.ProfileNotFound().
				WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		currency.Quantity += body.MatchStats.Eliminations * 25

		if body.MatchStats.Position == 1 {
			currency.Quantity += 150

			reward := "AthenaGlider:Umbrella_Season_" + strconv.Itoa(utilities.GetConfig().CURRENT_VERSION)
			umbrella := "AthenaGlider:Solo_Umbrella"

			rewardExists := false
			umbrellaExists := false
			for _, item := range items {
				if item.TemplateID == reward {
					rewardExists = true
				}
				if item.TemplateID == umbrella {
					umbrellaExists = true
				}
			}

			if !rewardExists {
				rewardItem := accounts.Items{
					AccountID:  accountId,
					ProfileID:  "athena",
					TemplateID: reward,
					Value: func() string {
						value, err := json.Marshal(gin.H{
							"xp":        0,
							"level":     1,
							"variants":  []gin.H{},
							"item_seen": false,
						})
						if err != nil {
							return ""
						}
						return string(value)
					}(),
					IsStat:   false,
					Quantity: 1,
					UUID:     uuid.New().String(),
				}
				if err := db.Create(&rewardItem).Error; err != nil {
					log.Printf("Error creating reward item: %v", err)
					utilities.MCP.InternalServerError().
						WithIntent(utilities.Prod).Apply(c.Writer)
					return
				}

				var umbrellaItem accounts.Items

				if !umbrellaExists {
					umbrellaItem = accounts.Items{
						AccountID:  accountId,
						ProfileID:  "athena",
						TemplateID: umbrella,
						Value: func() string {
							value, err := json.Marshal(gin.H{
								"xp":        0,
								"level":     1,
								"variants":  []gin.H{},
								"item_seen": false,
							})
							if err != nil {
								return ""
							}
							return string(value)
						}(),
						IsStat:   false,
						Quantity: 1,
						UUID:     uuid.New().String(),
					}
					if err := db.Create(&umbrellaItem).Error; err != nil {
						log.Printf("Error creating umbrella item: %v", err)
						utilities.MCP.InternalServerError().
							WithIntent(utilities.Prod).Apply(c.Writer)
						return
					}
				}

				lootList := []interface{}{
					map[string]interface{}{
						"itemType": reward,
						"itemGuid": rewardItem.UUID,
						"quantity": 1,
					},
					map[string]interface{}{
						"itemType": umbrella,
						"itemGuid": umbrellaItem.UUID,
						"quantity": 1,
					},
				}
				giftbox := accounts.Items{
					AccountID:  accountId,
					ProfileID:  "common_core",
					TemplateID: "GiftBox:GB_SeasonFirstWin",
					Value: func() string {
						value, err := json.Marshal(map[string]interface{}{
							"max_level_bonus": 0,
							"fromAccountId":   "",
							"lootList":        lootList,
						})
						if err != nil {
							return ""
						}
						return string(value)
					}(),
					IsStat:   false,
					Quantity: 1,
					UUID:     uuid.New().String(),
				}
				if err := db.Create(&giftbox).Error; err != nil {
					log.Printf("Error creating giftbox: %v", err)
					utilities.MCP.InternalServerError().
						WithIntent(utilities.Prod).Apply(c.Writer)
					return
				}
			}
		}

		if err := db.Save(&currency).Error; err != nil {
			log.Printf("Error saving currency: %v", err)
			utilities.MCP.InternalServerError().
				WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		xcacheManager := caching.GetBattlePassCacheManager()
		bcacheManager := caching.GetChallengeBundleCacheManager()

		XPData, found := xcacheManager.Get("AthenaSeason" + strconv.Itoa(utilities.GetConfig().CURRENT_VERSION) + "XPCurve")
		if !found {
			log.Printf("Error getting xp curve")
			utilities.MCP.ProfileNotFound().
				WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		Stars, found := xcacheManager.Get("BRSeason" + strconv.Itoa(utilities.GetConfig().CURRENT_VERSION) + "Stars")
		if !found {
			log.Printf("Error getting stars")
			utilities.MCP.ProfileNotFound().
				WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		xpInt, err := strconv.Atoi(xp.Value)
		if err != nil {
			log.Printf("Error parsing book_xp value: %v", err)
			return
		}

		for _, xpCurve := range XPData.([]battlepass.AthenaSeasonXPCurve) {
			if xpCurve.Rows[level.Value].Level > 0 {
				xp.Value = strconv.Itoa(xpInt + XP)
				xpInt, _ = strconv.Atoi(xp.Value)
				log.Printf("XP: %v", xp.Value)
				for xpInt >= xpCurve.Rows[level.Value].XpToNextLevel {
					xpInt -= xpCurve.Rows[level.Value].XpToNextLevel
					level.Value = strconv.Itoa(xpCurve.Rows[level.Value].Level + 1)
					bookLevelInt, err := strconv.Atoi(bookLevel.Value)

					for _, star := range Stars.([]battlepass.BattlePassStars) {
						if err != nil {
							log.Printf("Error getting book level: %v", err)
							utilities.MCP.ProfileNotFound().
								WithIntent(utilities.Prod).Apply(c.Writer)
							return
						}
						levelInt, err := strconv.Atoi(level.Value)
						if err != nil {
							log.Printf("Error converting level value: %v", err)
							utilities.MCP.ProfileNotFound().
								WithIntent(utilities.Prod).Apply(c.Writer)
							return
						}
						if star.Level == levelInt {
							if bookXp.Value == "" {
								bookXp.Value = "0"
							}
							bookXpInt, err := strconv.Atoi(bookXp.Value)
							if err != nil {
								log.Printf("Error getting book xp: %v", err)
								utilities.MCP.ProfileNotFound().
									WithIntent(utilities.Prod).Apply(c.Writer)
								return
							}
							if star.Quantity+bookXpInt >= 10 {
								bookXp.Value = strconv.Itoa(star.Quantity + bookXpInt - 10)
								bookLevel.Value = strconv.Itoa(bookLevelInt + 1)
								if bpInterface, found := xcacheManager.Get("BRSeason" + strconv.Itoa(utilities.GetConfig().CURRENT_VERSION)); found {
									bp, ok := bpInterface.(battlepass.BattlePass)
									if !ok {
										utilities.Storefront.InvalidItem().Apply(c.Writer)
										return
									} else {
										lootItems := []map[string]interface{}{}
										newItems := []accounts.Items{}

										var bpOwned accounts.Items
										if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
											"book_purchased", accountId, "athena").First(&bpOwned).Error; err == nil {
											var bpOwnedValue string
											if err := json.Unmarshal([]byte(bpOwned.Value), &bpOwnedValue); err == nil {
												if bpOwnedValue == "false" {
													utilities.Storefront.InvalidItem().Apply(c.Writer)
													return
												}
											}
										}

										bookLevelInt -= 1
										for i := bookLevelInt; i < bookLevelInt+1; i++ {
											for id, quantity := range bp.Rewards[i] {
												switch {
												case strings.Contains(strings.ToLower(id), "token:athenaseasonxpboost"):

												case strings.Contains(strings.ToLower(id), "token:athenaseasonfriendxpboost"):

												case strings.Contains(strings.ToLower(id), "currency:mtx"):
													currency.Quantity += quantity
													lootItems = append(lootItems, gin.H{
														"itemType": id,
														"itemGuid": id,
														"profile":  "common_core",
														"quantity": quantity,
													})

												case strings.Contains(strings.ToLower(id), "homebasebannericon"):
													var newItem accounts.Items
													if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
														id, accountId, "common_core").First(&newItem).Error; err == nil {
														continue
													} else {
														newValue, _ := json.Marshal(map[string]interface{}{
															"level":     1,
															"item_seen": false,
														})

														newItem = accounts.Items{
															AccountID:  accountId,
															ProfileID:  "common_core",
															TemplateID: id,
															Value:      string(newValue),
															Quantity:   quantity,
															IsStat:     false,
														}

														newItems = append(newItems, newItem)

														lootItems = append(lootItems, gin.H{
															"itemType": id,
															"itemGuid": id,
															"profile":  "common_core",
															"quantity": quantity,
														})
													}

												case strings.Contains(strings.ToLower(id), "athena"):
													var newItem accounts.Items
													if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
														id, accountId, "athena").First(&newItem).Error; err == nil {
														continue
													} else {
														newValue, _ := json.Marshal(map[string]interface{}{
															"max_level_bonus": 0,
															"level":           1,
															"xp":              0,
															"item_seen":       false,
															"variants":        []gin.H{},
															"favorite":        false,
														})

														newItem = accounts.Items{
															AccountID:  accountId,
															ProfileID:  "athena",
															TemplateID: id,
															Value:      string(newValue),
															Quantity:   quantity,
															IsStat:     false,
														}

														if !strings.Contains(strings.ToLower(id), "token:athenaseasonmergedxpboosts") {
															newItems = append(newItems, newItem)
														}
														lootItems = append(lootItems, gin.H{
															"itemType": id,
															"itemGuid": id,
															"profile":  "athena",
															"quantity": quantity,
														})
													}
												}
											}
										}
										giftboxValue := gin.H{
											"max_level_bonus": 0,
											"fromAccountId":   "",
											"lootList":        lootItems,
										}

										giftboxValueJSON, _ := json.Marshal(giftboxValue)

										gb := accounts.Items{
											AccountID:  accountId,
											ProfileID:  "common_core",
											TemplateID: "GiftBox:GB_BattlePass",
											Value:      string(giftboxValueJSON),
											Quantity:   1,
											IsStat:     false,
										}

										if err := db.Create(&gb).Error; err != nil {
											log.Printf("Error creating giftbox: %v", err)
											utilities.MCP.InternalServerError().
												WithIntent(utilities.Prod).Apply(c.Writer)
											return
										}

										tx := db.Begin()

										for _, item := range newItems {
											if err := tx.Create(&item).Error; err != nil {
												tx.Rollback()
												c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add items to database"})
												return
											}
										}

										if err := tx.Commit().Error; err != nil {
											c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
											return
										}
									}
								}

								if err := db.Save(&bookXp).Error; err != nil {
									log.Printf("Error saving book_xp: %v", err)
									utilities.MCP.InternalServerError().
										WithIntent(utilities.Prod).Apply(c.Writer)
									return
								}
								if err := db.Save(&bookLevel).Error; err != nil {
									log.Printf("Error saving book_level: %v", err)
									utilities.MCP.InternalServerError().
										WithIntent(utilities.Prod).Apply(c.Writer)
									return
								}

							} else {
								log.Printf("Star Quantity: %v", star.Quantity)
								bookXpInt += star.Quantity
								bookXp.Value = strconv.Itoa(bookXpInt)
								log.Printf("Book XP: %v", bookXp.Value)
								if err := db.Save(&bookXp).Error; err != nil {
									log.Printf("Error saving book_xp: %v", err)
									utilities.MCP.InternalServerError().
										WithIntent(utilities.Prod).Apply(c.Writer)
									return
								}
								if err := db.Save(&bookLevel).Error; err != nil {
									log.Printf("Error saving book_level: %v", err)
									utilities.MCP.InternalServerError().
										WithIntent(utilities.Prod).Apply(c.Writer)
									return
								}
							}
						}
					}

					if bundles, found := bcacheManager.TryGet(); found {
						var questsToUpdate []accounts.Quests
						var profileChangesToAdd []map[string]interface{}

						for _, quest := range quests {
							if !strings.HasPrefix(quest.TemplateID, "Quest:") {
								continue
							}

							var questAttributes map[string]interface{}
							if err := json.Unmarshal([]byte(quest.Value), &questAttributes); err != nil {
								log.Printf("Error unmarshaling quest attributes: %v", err)
								continue
							}

							questTemplate := strings.TrimPrefix(quest.TemplateID, "Quest:")
							var matchFound bool
							var questOptions struct {
								AthenaSeasonProgress bool
								BattlePassProgress   bool
							}
							var questObjectives []struct {
								BackendName string
								Count       int
								Stage       int
							}
							var questRewards []struct {
								TemplateId string
								Quantity   int
							}

							for _, bundle := range bundles {
								for _, obj := range bundle.Objects {
									if obj.QuestDefinition == questTemplate {
										questOptions = struct {
											AthenaSeasonProgress bool
											BattlePassProgress   bool
										}{
											AthenaSeasonProgress: obj.Options.AthenaSeasonProgress,
											BattlePassProgress:   obj.Options.BattlePassProgress,
										}
										for _, reward := range obj.Rewards {
											questRewards = append(questRewards, struct {
												TemplateId string
												Quantity   int
											}{
												TemplateId: reward.TemplateId,
												Quantity:   reward.Quantity,
											})
										}
										for _, objective := range obj.Objectives {
											questObjectives = append(questObjectives, struct {
												BackendName string
												Count       int
												Stage       int
											}{
												BackendName: objective.BackendName,
												Count:       objective.Count,
												Stage:       objective.Stage,
											})
										}
										matchFound = true
										break
									}
								}
								if matchFound {
									break
								}
							}

							if !matchFound {
								continue
							}

							if questOptions.AthenaSeasonProgress || questOptions.BattlePassProgress {
								updateNeeded := false

								for _, obj := range questObjectives {
									objKey := "completion_" + obj.BackendName

									if questOptions.AthenaSeasonProgress {
										seasonLevelInt, _ := strconv.Atoi(level.Value)
										if seasonLevelInt >= obj.Count {
											questAttributes[objKey] = obj.Count
											questAttributes["quest_state"] = "Completed"
											updateNeeded = true
											for _, reward := range questRewards {
												log.Printf("Processing reward: %v", reward.TemplateId)
												switch {
												case strings.Contains(reward.TemplateId, "CosmeticVariantToken"):
													cvtCacheManager := caching.GetBattlePassCacheManager()
													if cvtInterface, found := cvtCacheManager.Get(reward.TemplateId); found {
														if cvt, ok := cvtInterface.(battlepass.CosmeticVariantToken); ok {
															var existingItem accounts.Items
															templateIDParts := strings.Split(cvt.TemplateID, ":")
															if len(templateIDParts) == 2 {
																templateIDParts[1] = strings.ToLower(templateIDParts[1])
																cvt.TemplateID = strings.Join(templateIDParts, ":")
															}
															if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ?",
																accountId, "athena", cvt.TemplateID).First(&existingItem).Error; err != nil {
																log.Printf("Error finding existing item: %v", err)
																if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ?",
																	accountId, "athena", strings.Replace(cvt.TemplateID, ":", ":", 1)).First(&existingItem).Error; err != nil {
																	log.Printf("Error finding existing item with replaced template ID: %v", err)
																}
															}
															var value map[string]interface{}
															if err := json.Unmarshal([]byte(existingItem.Value), &value); err != nil {
																log.Printf("Error unmarshaling existing item value: %v", err)
																continue
															}

															if variants, ok := value["variants"].([]interface{}); ok {
																variantExists := false
																for i, v := range variants {
																	variant, ok := v.(map[string]interface{})
																	if !ok {
																		continue
																	}

																	if channel, ok := variant["channel"].(string); ok && channel == cvt.Channel {
																		variantExists = true

																		variant["active"] = cvt.Value

																		if owned, ok := variant["owned"].([]interface{}); ok {
																			ownedStrings := make([]string, len(owned))
																			for j, o := range owned {
																				if str, ok := o.(string); ok {
																					ownedStrings[j] = str
																				}
																			}

																			valueExists := false
																			for _, o := range ownedStrings {
																				if o == cvt.Value {
																					valueExists = true
																					break
																				}
																			}

																			if !valueExists {
																				ownedStrings = append(ownedStrings, cvt.Value)
																				variant["owned"] = ownedStrings
																			}
																		} else {
																			variant["owned"] = []string{cvt.Value}
																		}

																		variants[i] = variant

																		break
																	}
																}

																if !variantExists {
																	variants = append(variants, map[string]interface{}{
																		"channel": cvt.Channel,
																		"active":  cvt.Value,
																		"owned":   []string{cvt.Value},
																	})
																}

																value["variants"] = variants
																value["favorite"] = false
																value["item_seen"] = false
																value["level"] = 1
																value["max_level_bonus"] = 0
																value["xp"] = 0
															} else {
																value["variants"] = []interface{}{
																	map[string]interface{}{
																		"channel": cvt.Channel,
																		"active":  cvt.Value,
																		"owned":   []string{cvt.Value},
																	},
																}
																value["favorite"] = false
																value["item_seen"] = false
																value["level"] = 1
																value["max_level_bonus"] = 0
																value["xp"] = 0
															}

															updatedValue, err := json.Marshal(value)
															if err != nil {
																log.Printf("Error marshaling updated value: %v", err)
																continue
															}

															existingItem.Value = string(updatedValue)
															if err := db.Save(&existingItem).Error; err != nil {
																log.Printf("Error updating existing item: %v", err)
															}
														}
													} else {
														log.Printf("CosmeticVariantToken not found in cache: %v", reward.TemplateId)
													}
												default:
													log.Printf("Unknown reward type: %v", reward.TemplateId)
												}
											}
										} else {
											questAttributes[objKey] = seasonLevelInt
											updateNeeded = true
										}
									} else if questOptions.BattlePassProgress {
										bookLevelInt, _ := strconv.Atoi(bookLevel.Value)
										if bookLevelInt >= obj.Count {
											questAttributes[objKey] = obj.Count
											questAttributes["quest_state"] = "Completed"
											for _, reward := range questRewards {
												log.Printf("Processing reward: %v", reward.TemplateId)
												switch {
												case strings.Contains(reward.TemplateId, "CosmeticVariantToken"):
													cvtCacheManager := caching.GetBattlePassCacheManager()
													if cvtInterface, found := cvtCacheManager.Get(reward.TemplateId); found {
														if cvt, ok := cvtInterface.(battlepass.CosmeticVariantToken); ok {
															var existingItem accounts.Items
															templateIDParts := strings.Split(cvt.TemplateID, ":")
															if len(templateIDParts) == 2 {
																templateIDParts[1] = strings.ToLower(templateIDParts[1])
																cvt.TemplateID = strings.Join(templateIDParts, ":")
															}
															if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ?",
																accountId, "athena", cvt.TemplateID).First(&existingItem).Error; err != nil {
																log.Printf("Error finding existing item: %v", err)
																if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ?",
																	accountId, "athena", strings.Replace(cvt.TemplateID, ":", ":", 1)).First(&existingItem).Error; err != nil {
																	log.Printf("Error finding existing item with replaced template ID: %v", err)
																}
															}
															var value map[string]interface{}
															if err := json.Unmarshal([]byte(existingItem.Value), &value); err != nil {
																log.Printf("Error unmarshaling existing item value: %v", err)
																continue
															}

															if variants, ok := value["variants"].([]interface{}); ok {
																variantExists := false
																for i, v := range variants {
																	variant, ok := v.(map[string]interface{})
																	if !ok {
																		continue
																	}

																	if channel, ok := variant["channel"].(string); ok && channel == cvt.Channel {
																		variantExists = true

																		variant["active"] = cvt.Value

																		if owned, ok := variant["owned"].([]interface{}); ok {
																			ownedStrings := make([]string, len(owned))
																			for j, o := range owned {
																				if str, ok := o.(string); ok {
																					ownedStrings[j] = str
																				}
																			}

																			valueExists := false
																			for _, o := range ownedStrings {
																				if o == cvt.Value {
																					valueExists = true
																					break
																				}
																			}

																			if !valueExists {
																				ownedStrings = append(ownedStrings, cvt.Value)
																				variant["owned"] = ownedStrings
																			}
																		} else {
																			variant["owned"] = []string{cvt.Value}
																		}

																		variants[i] = variant

																		break
																	}
																}

																if !variantExists {
																	variants = append(variants, map[string]interface{}{
																		"channel": cvt.Channel,
																		"active":  cvt.Value,
																		"owned":   []string{cvt.Value},
																	})
																}

																value["variants"] = variants
																value["favorite"] = false
																value["item_seen"] = false
																value["level"] = 1
																value["max_level_bonus"] = 0
																value["xp"] = 0
															} else {
																value["variants"] = []interface{}{
																	map[string]interface{}{
																		"channel": cvt.Channel,
																		"active":  cvt.Value,
																		"owned":   []string{cvt.Value},
																	},
																}
																value["favorite"] = false
																value["item_seen"] = false
																value["level"] = 1
																value["max_level_bonus"] = 0
																value["xp"] = 0
															}

															updatedValue, err := json.Marshal(value)
															if err != nil {
																log.Printf("Error marshaling updated value: %v", err)
																continue
															}

															existingItem.Value = string(updatedValue)
															if err := db.Save(&existingItem).Error; err != nil {
																log.Printf("Error updating existing item: %v", err)
															}
														}
													} else {
														log.Printf("CosmeticVariantToken not found in cache: %v", reward.TemplateId)
													}
												default:
													log.Printf("Unknown reward type: %v", reward.TemplateId)
												}
											}
											updateNeeded = true
										} else {
											questAttributes[objKey] = bookLevelInt
											updateNeeded = true
										}
									}
								}

								if updateNeeded {
									newValue, err := json.Marshal(questAttributes)
									if err != nil {
										log.Printf("Error marshaling updated quest attributes: %v", err)
										continue
									}

									quest.Value = string(newValue)
									questsToUpdate = append(questsToUpdate, quest)

									if questAttributes["quest_state"] == "Completed" {
										bundleID, ok := questAttributes["challenge_bundle_id"].(string)
										if ok {
											var bundle accounts.Quests
											if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ? AND season = ?",
												accountId, "athena", bundleID, utilities.GetConfig().CURRENT_VERSION).First(&bundle).Error; err == nil {
												var bundleAttrs map[string]interface{}
												if err := json.Unmarshal([]byte(bundle.Value), &bundleAttrs); err == nil {
													if numCompleted, ok := bundleAttrs["num_quests_completed"].(float64); ok {
														bundleAttrs["num_quests_completed"] = numCompleted + 1
														bundleAttrs["num_progress_quests_completed"] = numCompleted + 1

														newBundleValue, err := json.Marshal(bundleAttrs)
														if err == nil {
															bundle.Value = string(newBundleValue)
															questsToUpdate = append(questsToUpdate, bundle)
														}
													}
												}
											}
										}
									}

									profileChangesToAdd = append(profileChangesToAdd,
										map[string]interface{}{
											"changeType":     "itemAttrChanged",
											"itemId":         quest.TemplateID,
											"attributeName":  "quest_state",
											"attributeValue": questAttributes["quest_state"],
										},
										map[string]interface{}{
											"changeType":     "itemAttrChanged",
											"itemId":         quest.TemplateID,
											"attributeName":  "completion_" + questObjectives[0].BackendName,
											"attributeValue": questAttributes["completion_"+questObjectives[0].BackendName],
										},
									)
								}
							}
						}

						if len(questsToUpdate) > 0 {
							for _, quest := range questsToUpdate {
								if err := db.Save(&quest).Error; err != nil {
									log.Printf("Error saving updated quest: %v", err)
								}
							}
						}
						profileChanges = append(profileChanges, profileChangesToAdd...)
					}
				}
				break
			}
		}
		xp.Value = strconv.Itoa(xpInt)
		log.Printf("XP: %v", xp.Value)
		if err := db.Save(&level).Error; err != nil {
			log.Printf("Error saving level: %v", err)
			utilities.MCP.InternalServerError().
				WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}
		if err := db.Save(&xp).Error; err != nil {
			log.Printf("Error saving xp: %v", err)
			utilities.MCP.InternalServerError().
				WithIntent(utilities.Prod).Apply(c.Writer)
			return
		}

		response.ProfileChanges = profile.Revision
		response.ProfileId = "athena"
		response.ProfileChangesBaseRevision = body.AthenaProfile.ProfileChangesBaseRevision
		response.ProfileCommandRevision = profile.Revision
		response.ServerTime = now
		response.ResponseVersion = 1
		response.ProfileChanges = profileChanges
	}

	c.JSON(200, response)
}
