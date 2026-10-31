package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"solana/classes/battlepass"
	"solana/modules/caching"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/modules/database/entities/fortnite"
	utilities "solana/utilities"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var allItems []accounts.Items

func PurchaseCatalogEntry(c *gin.Context) {
	if len(allItems) == 0 {
		if err := database.Get().Find(&allItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load items from the database"})
			return
		}
	}

	accountID := c.Param("accountId")
	profileID := c.Query("profileId")

	if profileID == "" {
		utilities.Basic.NotAcceptable().Apply(c.Writer)
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		utilities.Basic.NotAcceptable().Apply(c.Writer)
		return
	}

	offerID, ok := body["offerId"].(string)
	if !ok {
		utilities.Basic.NotAcceptable().Apply(c.Writer)
		return
	}

	db := database.Get()

	var profileData accounts.Profiles
	if err := db.Where("profile_id = ? AND account_id = ?", profileID, accountID).First(&profileData).Error; err != nil {
		utilities.Storefront.InvalidItem().Apply(c.Writer)
		return
	}

	var athenaData accounts.Profiles
	if err := db.Where("profile_id = ? AND account_id = ?", "athena", accountID).First(&athenaData).Error; err != nil {
		utilities.Storefront.InvalidItem().Apply(c.Writer)
		return
	}

	var catalog fortnite.Catalog
	if err := db.Where(`"offerId" = ?`, offerID).First(&catalog).Error; err != nil {
		utilities.Storefront.InvalidItem().Apply(c.Writer)
		return
	}

	var offerData struct {
		OfferID    string `json:"offerId"`
		ItemGrants []struct {
			TemplateID string `json:"templateId"`
		} `json:"itemGrants"`
		Prices []struct {
			CurrencyType string  `json:"currencyType"`
			FinalPrice   float64 `json:"finalPrice"`
		} `json:"prices"`
	}

	if err := json.Unmarshal([]byte(catalog.Data), &offerData); err != nil {
		utilities.Storefront.InvalidItem().Apply(c.Writer)
		return
	}

	if len(offerData.ItemGrants) == 0 && !strings.Contains(catalog.Storefront, "BRSeason") {
		utilities.Storefront.InvalidItem().Apply(c.Writer)
		return
	}

	for _, itemGrant := range offerData.ItemGrants {
		var existingItem accounts.Items
		if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
			itemGrant.TemplateID, accountID, "athena").First(&existingItem).Error; err == nil {

			utilities.Storefront.AlreadyOwned().Apply(c.Writer)
			return
		} else {

		}
	}

	var currency accounts.Items
	if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
		"Currency:MtxPurchased", accountID, profileID).First(&currency).Error; err != nil {

		currency = accounts.Items{
			AccountID:  accountID,
			ProfileID:  profileID,
			TemplateID: "Currency:MtxPurchased",
			Value:      `{"platform":"EpicPC"}`,
			Quantity:   0,
			IsStat:     false,
		}
		db.Create(&currency)
	}

	price := offerData.Prices[0].FinalPrice
	currencyType := offerData.Prices[0].CurrencyType

	if currencyType == "MtxCurrency" {
		var currencyValue map[string]interface{}
		json.Unmarshal([]byte(currency.Value), &currencyValue)
		if float64(currency.Quantity) <= price {
			log.Println("Currency is insufficient")
			log.Println(currency.Quantity)
			log.Println(price)
			utilities.Storefront.CurrencyInsufficient().Apply(c.Writer)
			return
		}
	} else if price > 0 {
		log.Println("Price is not zero")
		utilities.Storefront.CurrencyInsufficient().Apply(c.Writer)
		return
	}

	if currencyType == "MtxCurrency" {
		currency.Quantity -= int(price)
		db.Save(&currency)
	}

	notifications := []gin.H{}
	athenaChanges := []map[string]interface{}{}
	commonCoreChanges := []map[string]interface{}{}
	profileChanges := []map[string]interface{}{}
	lootItems := []map[string]interface{}{}
	newItems := []accounts.Items{}

	if strings.Contains(catalog.Storefront, "BRSeason") {
		ua := utilities.Parse(c.GetHeader("User-Agent"))
		cacheManager := caching.GetBattlePassCacheManager()
		if _, found := cacheManager.Get(catalog.Storefront); found {
			if strconv.Itoa(ua.Season) != strings.Split(catalog.Storefront, "BRSeason")[1] {
				utilities.Storefront.InvalidItem().Apply(c.Writer)
				return
			} else {
				bpInterface, _ := cacheManager.Get(catalog.Storefront)
				bp, ok := bpInterface.(battlepass.BattlePass)
				if !ok {
					utilities.Storefront.InvalidItem().Apply(c.Writer)
					return
				}
				var bpOwned accounts.Items
				if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
					"book_purchased", accountID, "athena").First(&bpOwned).Error; err == nil {
					var bpOwnedValue string
					if err := json.Unmarshal([]byte(bpOwned.Value), &bpOwnedValue); err == nil {
						if bpOwnedValue == "true" {
							utilities.Storefront.AlreadyOwned().Apply(c.Writer)
							return
						}
					}
					cacheManager := caching.GetChallengeBundleCacheManager()

					var Count int
					var bookLevel accounts.Items
					if err := db.Where(`template_id = ? AND account_id = ?`, "book_level", accountID).First(&bookLevel).Error; err != nil {
						utilities.Storefront.InvalidItem().Apply(c.Writer)
						return
					}

					var seasonLevel accounts.Items
					if err := db.Where(`template_id = ? AND account_id = ?`, "level", accountID).First(&seasonLevel).Error; err != nil {
						utilities.Storefront.InvalidItem().Apply(c.Writer)
						return
					}

					currentBookLevel, err := strconv.Atoi(bookLevel.Value)
					if err != nil {
						utilities.Storefront.InvalidItem().Apply(c.Writer)
						return
					}

					if strings.Contains(catalog.Name, "BattleBundle") {
						Count = 25
					} else if strings.Contains(catalog.Name, "BattlePass") {
						Count = 1
					} else if strings.Contains(catalog.Name, "SingleTier") {
						if purchaseQuantity, ok := body["purchaseQuantity"]; ok {
							if purchaseQuantityInt, err := strconv.Atoi(fmt.Sprintf("%v", purchaseQuantity)); err == nil {
								Count = purchaseQuantityInt
							} else {
								utilities.Storefront.InvalidItem().Apply(c.Writer)
								return
							}
						}
					}

					bookLevel.Value = strconv.Itoa(Count + currentBookLevel)

					if strings.Contains(catalog.Name, "SingleTier") {
						bookLevelInt, err := strconv.Atoi(bookLevel.Value)
						if err != nil {
							utilities.Storefront.InvalidItem().Apply(c.Writer)
							return
						}
						bookLevelInt -= Count
						for i := bookLevelInt; i < bookLevelInt+Count; i++ {
							for id, quantity := range bp.Rewards[i] {
								switch {
								case strings.Contains(strings.ToLower(id), "token:athenaseasonxpboost"):

								case strings.Contains(strings.ToLower(id), "token:athenaseasonfriendxpboost"):

								case strings.Contains(strings.ToLower(id), "currency:mtx"):
									currency.Quantity += quantity
									profileChanges = append(profileChanges, gin.H{
										"changeType": "itemQuantityChanged",
										"itemId":     "Currency:MtxPurchased",
										"quantity":   quantity,
									})
									lootItems = append(lootItems, gin.H{
										"itemType": id,
										"itemGuid": id,
										"profile":  "common_core",
										"quantity": quantity,
									})

								case strings.Contains(strings.ToLower(id), "homebasebannericon"):
									var newItem accounts.Items
									if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
										id, accountID, "common_core").First(&newItem).Error; err == nil {
										continue
									} else {
										newValue, _ := json.Marshal(map[string]interface{}{
											"level":     1,
											"item_seen": false,
										})

										newItem = accounts.Items{
											AccountID:  accountID,
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
										commonCoreChanges = append(commonCoreChanges, gin.H{
											"changeType": "itemAdded",
											"itemId":     id,
											"item": gin.H{
												"templateId": id,
												"attributes": gin.H{
													"level":     1,
													"item_seen": false,
												},
												"quantity": quantity,
											},
										})
									}

								case strings.Contains(strings.ToLower(id), "athena"):
									var newItem accounts.Items
									if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
										id, accountID, "athena").First(&newItem).Error; err == nil {
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
											AccountID:  accountID,
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
										athenaChanges = append(athenaChanges, gin.H{
											"changeType": "itemAdded",
											"itemId":     id,
											"item": gin.H{
												"templateId": id,
												"attributes": gin.H{
													"max_level_bonus": 0,
													"level":           1,
													"xp":              0,
													"item_seen":       false,
													"variants":        []gin.H{},
													"favorite":        false,
												},
												"quantity": quantity,
											},
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

						profileChanges = append(profileChanges, gin.H{
							"changeType": "itemAdded",
							"itemId":     "GiftBox:GB_BattlePass",
							"item": gin.H{
								"templateId": "GiftBox:GB_BattlePass",
								"attributes": giftboxValue,
								"quantity":   1,
							},
						})

						gb := accounts.Items{
							AccountID:  accountID,
							ProfileID:  "common_core",
							TemplateID: "GiftBox:GB_BattlePass",
							Value:      string(giftboxValueJSON),
							Quantity:   1,
							IsStat:     false,
						}

						db.Create(&gb)
					} else {
						for i := 0; i < Count; i++ {
							for id, quantity := range bp.Rewards[i] {
								switch {
								case strings.Contains(strings.ToLower(id), "token:athenaseasonxpboost"):

								case strings.Contains(strings.ToLower(id), "token:athenaseasonfriendxpboost"):

								case strings.Contains(strings.ToLower(id), "currency:mtx"):

								case strings.Contains(strings.ToLower(id), "homebasebannericon"):

								case strings.Contains(strings.ToLower(id), "athena"):
									var newItem accounts.Items
									if err := db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
										id, accountID, "athena").First(&newItem).Error; err == nil {
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
											AccountID:  accountID,
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
										athenaChanges = append(athenaChanges, gin.H{
											"changeType": "itemAdded",
											"itemId":     id,
											"item": gin.H{
												"templateId": id,
												"attributes": gin.H{
													"max_level_bonus": 0,
													"level":           1,
													"xp":              0,
													"item_seen":       false,
													"variants":        []gin.H{},
													"favorite":        false,
												},
												"quantity": quantity,
											},
										})
									}
								case strings.Contains(strings.ToLower(id), "challengebundleschedule"):
									if bundles, found := cacheManager.TryGet(); found {
										scheduleToGrantedBundles := make(map[string][]string)
										var allQuestsToSave []accounts.Quests
										now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")

										processedQuests := make(map[string]bool)
										processedBundles := make(map[string]bool)
										processedSchedules := make(map[string]bool)

										for _, bundle := range bundles {
											if bundle.ChallengeBundleSchedule != id {
												continue
											}

											bundleName := "ChallengeBundle:" + bundle.TemplateId
											grantedQuestInstanceIds := []string{}
											challengeBundleScheduleId := bundle.ChallengeBundleSchedule

											var existingBundle accounts.Quests
											bundleExists := false
											if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ? AND season = ?",
												accountID, "athena", bundleName, ua.Season).First(&existingBundle).Error; err == nil {
												bundleExists = true
												processedBundles[bundleName] = true
											}

											if !processedBundles[bundleName] {
												scheduleToGrantedBundles[challengeBundleScheduleId] = append(
													scheduleToGrantedBundles[challengeBundleScheduleId],
													bundleName,
												)
											}

											for _, quest := range bundle.Objects {
												questTemplateId := "Quest:" + quest.QuestDefinition

												if processedQuests[questTemplateId] {
													continue
												}

												var existingQuest accounts.Quests
												if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ? AND season = ?",
													accountID, "athena", questTemplateId, ua.Season).First(&existingQuest).Error; err == nil {
													processedQuests[questTemplateId] = true
													continue
												}

												objectiveStates := make(map[string]interface{})
												for _, obj := range quest.Objectives {
													if quest.Options.ProgressOnBattlePassPurchased {
														objectiveStates["completion_"+obj.BackendName] = obj.Count
													} else if quest.Options.BattlePassProgress {
														bookLevelInt, _ := strconv.Atoi(bookLevel.Value)
														if bookLevelInt >= obj.Count {
															objectiveStates["completion_"+obj.BackendName] = obj.Count
														}
													} else if quest.Options.AthenaSeasonProgress {
														seasonLevelInt, _ := strconv.Atoi(seasonLevel.Value)
														if seasonLevelInt >= obj.Count {
															objectiveStates["completion_"+obj.BackendName] = obj.Count
														} else {
															objectiveStates["completion_"+obj.BackendName] = seasonLevelInt
														}
													} else {
														objectiveStates["completion_"+obj.BackendName] = obj.Stage
													}
												}

												combinedAttributes := map[string]interface{}{
													"creation_time":          now,
													"level":                  -1,
													"item_seen":              true,
													"playlists":              []string{},
													"sent_new_notification":  true,
													"challenge_bundle_id":    bundleName,
													"xp_reward_scalar":       1,
													"quest_state":            "Active",
													"last_state_change_time": now,
													"quest_rarity":           "uncommon",
													"favorite":               false,
												}

												if quest.Options.ProgressOnBattlePassPurchased {
													combinedAttributes["quest_state"] = "Completed"
												}

												if quest.Options.BattlePassProgress {
													bookLevelInt, _ := strconv.Atoi(bookLevel.Value)
													for _, obj := range quest.Objectives {
														if bookLevelInt >= obj.Count {
															combinedAttributes["quest_state"] = "Completed"
															for _, reward := range quest.Rewards {
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
																				accountID, "athena", cvt.TemplateID).First(&existingItem).Error; err != nil {
																				log.Printf("Error finding existing item: %v", err)
																				if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ?",
																					accountID, "athena", strings.Replace(cvt.TemplateID, ":", ":", 1)).First(&existingItem).Error; err != nil {
																					log.Printf("Error finding existing item with replaced template ID: %v", err)
																				}
																			}
																			var value map[string]interface{}
																			if err := json.Unmarshal([]byte(existingItem.Value), &value); err != nil {
																				log.Printf("Error unmarshaling existing item value: %v", err)
																				continue
																			}
																			log.Printf("Existing item value: %v", value)
																			if variants, ok := value["variants"].([]interface{}); ok {
																				variants = append(variants, map[string]interface{}{
																					"channel": cvt.Channel,
																					"active":  cvt.Value,
																					"owned":   []string{cvt.Value},
																				})
																				value["variants"] = variants
																			} else {
																				value["variants"] = []interface{}{
																					map[string]interface{}{
																						"channel": cvt.Channel,
																						"active":  cvt.Value,
																						"owned":   []string{cvt.Value},
																					},
																				}
																			}
																			newValue, _ := json.Marshal(value)
																			existingItem.Value = string(newValue)
																			if err := db.Model(&existingItem).Updates(existingItem).Error; err != nil {
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
														}
													}
												}

												if quest.Options.AthenaSeasonProgress {
													seasonLevelInt, _ := strconv.Atoi(seasonLevel.Value)
													for _, obj := range quest.Objectives {
														if seasonLevelInt >= obj.Count {
															combinedAttributes["quest_state"] = "Completed"
															for _, reward := range quest.Rewards {
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
																				accountID, "athena", cvt.TemplateID).First(&existingItem).Error; err != nil {
																				log.Printf("Error finding existing item: %v", err)
																				if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ?",
																					accountID, "athena", strings.Replace(cvt.TemplateID, ":", ":", 1)).First(&existingItem).Error; err != nil {
																					log.Printf("Error finding existing item with replaced template ID: %v", err)
																				}
																			}
																			var value map[string]interface{}
																			if err := json.Unmarshal([]byte(existingItem.Value), &value); err != nil {
																				log.Printf("Error unmarshaling existing item value: %v", err)
																				continue
																			}
																			if variants, ok := value["variants"].([]interface{}); ok {
																				variants = append(variants, map[string]interface{}{
																					"channel": cvt.Channel,
																					"active":  cvt.Value,
																					"owned":   []string{cvt.Value},
																				})
																				value["variants"] = variants
																			} else {
																				value["variants"] = []interface{}{
																					map[string]interface{}{
																						"channel": cvt.Channel,
																						"active":  cvt.Value,
																						"owned":   []string{cvt.Value},
																					},
																				}
																			}
																			newValue, _ := json.Marshal(value)
																			existingItem.Value = string(newValue)
																			if err := db.Model(&existingItem).Updates(existingItem).Error; err != nil {
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
														}
													}
												}

												for key, value := range objectiveStates {
													combinedAttributes[key] = value
												}

												grantedQuestInstanceIds = append(grantedQuestInstanceIds, questTemplateId)
												processedQuests[questTemplateId] = true

												questUUID := strings.Replace(uuid.New().String(), "-", "", -1)
												attributesJSON, _ := json.Marshal(combinedAttributes)
												athenaChanges = append(athenaChanges, gin.H{
													"changeType": "itemAdded",
													"itemId":     questTemplateId,
													"item": gin.H{
														"templateId": questTemplateId,
														"attributes": combinedAttributes,
														"quantity":   quantity,
													},
												})
												allQuestsToSave = append(allQuestsToSave, accounts.Quests{
													UUID:       questUUID,
													AccountID:  accountID,
													ProfileID:  "athena",
													Daily:      false,
													TemplateID: questTemplateId,
													Value:      string(attributesJSON),
													Season:     ua.Season,
												})
											}

											if !bundleExists && !processedBundles[bundleName] {
												bundleItemAttributes := map[string]interface{}{
													"has_unlock_by_completion":      false,
													"num_quests_completed":          0,
													"level":                         0,
													"grantedquestinstanceids":       grantedQuestInstanceIds,
													"item_seen":                     true,
													"max_allowed_bundle_level":      0,
													"num_granted_bundle_quests":     len(grantedQuestInstanceIds),
													"max_level_bonus":               0,
													"challenge_bundle_schedule_id":  challengeBundleScheduleId,
													"num_progress_quests_completed": 0,
													"xp":                            0,
													"favorite":                      false,
												}

												bundleUUID := strings.Replace(uuid.New().String(), "-", "", -1)
												bundleJSON, _ := json.Marshal(bundleItemAttributes)
												processedBundles[bundleName] = true

												athenaChanges = append(athenaChanges, gin.H{
													"changeType": "itemAdded",
													"itemId":     bundleName,
													"item": gin.H{
														"templateId": bundleName,
														"attributes": bundleItemAttributes,
														"quantity":   quantity,
													},
												})

												allQuestsToSave = append(allQuestsToSave, accounts.Quests{
													UUID:       bundleUUID,
													AccountID:  accountID,
													ProfileID:  "athena",
													Daily:      false,
													TemplateID: bundleName,
													Value:      string(bundleJSON),
													Season:     ua.Season,
												})
											}
										}

										for scheduleId, grantedBundles := range scheduleToGrantedBundles {
											if processedSchedules[scheduleId] {
												continue
											}

											var existingSchedule accounts.Quests
											if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ? AND season = ?",
												accountID, "athena", scheduleId, ua.Season).First(&existingSchedule).Error; err == nil {
												processedSchedules[scheduleId] = true
												continue
											}

											scheduleItemAttributes := map[string]interface{}{
												"unlock_epoch":    now,
												"max_level_bonus": 0,
												"level":           1,
												"item_seen":       true,
												"xp":              0,
												"favorite":        false,
												"granted_bundles": grantedBundles,
											}

											scheduleUUID := strings.Replace(uuid.New().String(), "-", "", -1)
											scheduleJSON, _ := json.Marshal(scheduleItemAttributes)
											processedSchedules[scheduleId] = true

											allQuestsToSave = append(allQuestsToSave, accounts.Quests{
												UUID:       scheduleUUID,
												AccountID:  accountID,
												ProfileID:  "athena",
												Daily:      false,
												TemplateID: scheduleId,
												Value:      string(scheduleJSON),
												Season:     ua.Season,
											})

											athenaChanges = append(athenaChanges, gin.H{
												"changeType": "itemAdded",
												"itemId":     scheduleId,
												"item": gin.H{
													"templateId": scheduleId,
													"attributes": scheduleItemAttributes,
													"quantity":   quantity,
												},
											})
										}

										if len(allQuestsToSave) > 0 {
											db.Create(&allQuestsToSave)
										}
									} else {
										fmt.Println("no bundles found")
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

						profileChanges = append(profileChanges, gin.H{
							"changeType": "itemAdded",
							"itemId":     "GiftBox:GB_BattlePassPurchased",
							"item": gin.H{
								"templateId": "GiftBox:GB_BattlePassPurchased",
								"attributes": giftboxValue,
								"quantity":   1,
							},
						})

						gb := accounts.Items{
							AccountID:  accountID,
							ProfileID:  "common_core",
							TemplateID: "GiftBox:GB_BattlePassPurchased",
							Value:      string(giftboxValueJSON),
							Quantity:   1,
							IsStat:     false,
						}

						db.Create(&gb)

						athenaChanges = append(athenaChanges, gin.H{
							"changeType": "statModified",
							"name":       "book_purchased",
							"value":      true,
						})

						bpOwnedValue = "true"
						bpOwned.Value = bpOwnedValue
						db.Save(&bpOwned)
					}

					bookLevelInt, err := strconv.Atoi(bookLevel.Value)
					if err != nil {
						utilities.Storefront.InvalidItem().Apply(c.Writer)
						return
					}
					athenaChanges = append(athenaChanges, gin.H{
						"changeType": "statModified",
						"name":       "book_level",
						"value":      bookLevelInt,
					})
					db.Save(&bookLevel)
				}

				profileData.Revision++
				athenaData.Revision++
				db.Save(&profileData)
				db.Save(&athenaData)

				if currencyType == "MtxCurrency" {
					profileChanges = append(profileChanges, map[string]interface{}{
						"changeType": "itemQuantityChanged",
						"itemId":     "Currency:MtxPurchased",
						"quantity":   currency.Quantity,
					})
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

				profileChanges = append(profileChanges, commonCoreChanges...)

				c.JSON(http.StatusOK, gin.H{
					"profileRevision":            profileData.Revision,
					"profileId":                  profileID,
					"profileChangesBaseRevision": profileData.Revision - 1,
					"profileChanges":             profileChanges,
					"notifications": gin.H{
						"type":    "CatalogPurchase",
						"primary": true,
						"lootResult": gin.H{
							"items": lootItems,
						},
					},
					"profileCommandRevision": profileData.Revision,
					"serverTime":             time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
					"multiUpdate": []gin.H{
						{
							"profileRevision":            athenaData.Revision,
							"profileId":                  "athena",
							"profileChangesBaseRevision": athenaData.Revision - 1,
							"profileChanges":             athenaChanges,
							"profileCommandRevision":     athenaData.Revision,
						},
					},
					"responseVersion": 1,
				})
			}
		}
		return
	}

	var mtxAffiliate accounts.Items
	db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
		"mtx_affiliate", accountID, profileID).First(&mtxAffiliate)

	var platform accounts.Items
	db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
		"current_mtx_platform", accountID, "common_core").First(&platform)

	var mtxHistory accounts.Items
	db.Where("template_id = ? AND account_id = ? AND profile_id = ?",
		"mtx_history", accountID, profileID).First(&mtxHistory)

	for _, itemGrant := range offerData.ItemGrants {
		if len(itemGrant.TemplateID) > 0 && itemGrant.TemplateID[0:6] == "Athena" {
			var foundItem bool = false
			for _, item := range allItems {
				if item.TemplateID == itemGrant.TemplateID {
					foundItem = true
					var itemValue map[string]interface{}
					json.Unmarshal([]byte(item.Value), &itemValue)

					variants := []interface{}{}
					if variantsData, ok := itemValue["variants"]; ok {
						if variantsBytes, err := json.Marshal(variantsData); err == nil {
							var variantsArray []map[string]interface{}
							json.Unmarshal(variantsBytes, &variantsArray)

							for _, v := range variantsArray {
								variants = append(variants, map[string]interface{}{
									"channel": v["channel"],
									"active":  v["active"],
									"owned":   v["owned"],
								})
							}
						}
					}

					newValue, _ := json.Marshal(map[string]interface{}{
						"max_level_bonus": 0,
						"level":           1,
						"xp":              0,
						"item_seen":       false,
						"variants":        variants,
						"favorite":        false,
					})

					newItem := accounts.Items{
						AccountID:  accountID,
						ProfileID:  "athena",
						TemplateID: itemGrant.TemplateID,
						Value:      string(newValue),
						Quantity:   1,
						IsStat:     false,
					}

					newItems = append(newItems, newItem)
					break
				}
			}

			if !foundItem {
				newValue, _ := json.Marshal(map[string]interface{}{
					"max_level_bonus": 0,
					"level":           1,
					"xp":              0,
					"item_seen":       false,
					"variants":        []interface{}{},
					"favorite":        false,
				})

				newItem := accounts.Items{
					AccountID:  accountID,
					ProfileID:  "athena",
					TemplateID: itemGrant.TemplateID,
					Value:      string(newValue),
					Quantity:   1,
					IsStat:     false,
				}

				newItems = append(newItems, newItem)
			}
		}
	}

	tx := db.Begin()

	for _, item := range newItems {
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add items to database"})
			return
		}

		allItems = append(allItems, item)
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	profileData.Revision++
	athenaData.Revision++
	db.Save(&profileData)
	db.Save(&athenaData)

	if currencyType == "MtxCurrency" {
		profileChanges = append(profileChanges, map[string]interface{}{
			"changeType": "itemQuantityChanged",
			"itemId":     "Currency:MtxPurchased",
			"quantity":   currency.Quantity,
		})
	}

	for _, itemGrant := range offerData.ItemGrants {
		lootItems = append(lootItems, map[string]interface{}{
			"itemType":    itemGrant.TemplateID,
			"itemGuid":    itemGrant.TemplateID,
			"itemProfile": "athena",
			"quantity":    1,
		})
	}

	for _, itemGrant := range offerData.ItemGrants {
		athenaChanges = append(athenaChanges, map[string]interface{}{
			"changeType": "itemAdded",
			"itemId":     itemGrant.TemplateID,
			"item": map[string]interface{}{
				"templateId": itemGrant.TemplateID,
				"attributes": map[string]interface{}{
					"max_level_bonus": 0,
					"level":           1,
					"xp":              0,
					"item_seen":       false,
					"variants":        []interface{}{},
					"favorite":        false,
				},
				"quantity": 1,
			},
		})
	}

	notifications = append(notifications, gin.H{
		"type":    "CatalogPurchase",
		"primary": true,
		"lootResult": gin.H{
			"items": lootItems,
		},
	})

	c.JSON(http.StatusOK, gin.H{
		"profileRevision":            profileData.Revision,
		"profileId":                  profileID,
		"profileChangesBaseRevision": profileData.Revision - 1,
		"profileChanges":             profileChanges,
		"notifications":              notifications,
		"profileCommandRevision":     profileData.Revision,
		"serverTime":                 time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
		"multiUpdate": []gin.H{
			{
				"profileRevision":            athenaData.Revision,
				"profileId":                  "athena",
				"profileChangesBaseRevision": athenaData.Revision - 1,
				"profileChanges":             athenaChanges,
				"profileCommandRevision":     athenaData.Revision,
			},
		},
		"responseVersion": 1,
	})
}
