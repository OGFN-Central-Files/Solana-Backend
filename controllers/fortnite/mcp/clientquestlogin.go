package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"solana/classes/mcp"
	"solana/managers"
	"solana/modules/caching"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	utilities "solana/utilities"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ClientQuestLogin(c *gin.Context) {
	accountId := c.Param("accountId")
	profileId := c.Query("profileId")
	db := database.Get()

	if profileId == "" {
		utilities.Internal.ValidationFailed().
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	ua := utilities.Parse(c.GetHeader("User-Agent"))

	if ua == nil {
		utilities.Basic.BadRequest().Apply(c.Writer)
		return
	}

	now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	var response mcp.DefaultMCPResponse

	var user accounts.Users
	if err := db.Where("\"accountId\" = ?", accountId).First(&user).Error; err != nil {
		utilities.MCP.ProfileNotFound().
			WithMessageVar([]string{accountId}).
			WithIntent(utilities.Prod).Apply(c.Writer)
		return
	}

	var profile accounts.Profiles
	var items []accounts.Items
	var quests []accounts.Quests
	profileFound := true

	if err := db.Where("account_id = ? AND profile_id = ?", accountId, profileId).First(&profile).Error; err != nil {
		profileFound = false
	} else if profileFound {
		db.Where("account_id = ? AND profile_id = ?", accountId, profileId).Find(&items)
		if profileId == "athena" {
			db.Where("account_id = ? and profile_id = ? ", accountId, "athena").Find(&quests)
		}
	}

	if !profileFound {
		response = mcp.DefaultMCPResponse{
			ProfileRevision:            0,
			ProfileId:                  profileId,
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
						ProfileId:  profileId,
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
		profile.Revision++
		db.Save(&profile)

		profileData := mcp.Profile{
			Created:         now,
			Updated:         now,
			Rvn:             profile.Revision,
			WipeNumber:      1,
			AccountId:       accountId,
			ProfileId:       profile.ProfileID,
			Version:         "no_version",
			CommandRevision: profile.Revision,
			Items:           make(map[string]interface{}),
			Stats: mcp.Stats{
				Attributes: make(map[string]interface{}),
			},
		}

		for _, item := range items {
			if !item.IsStat {
				var valueMap interface{}
				json.Unmarshal([]byte(item.Value), &valueMap)
				itemData := mcp.BaseItem{
					Attributes: valueMap,
					TemplateId: item.TemplateID,
					Quantity:   item.Quantity,
				}
				if strings.Contains(item.TemplateID, "loadout") {
					itemData.TemplateId = "CosmeticLocker:cosmeticlocker_athena"
				}
				if profileId == "creative" {
					profileData.Items[item.UUID] = itemData
				} else {
					profileData.Items[item.TemplateID] = itemData
				}
			} else {
				var statValue interface{}
				json.Unmarshal([]byte(item.Value), &statValue)
				profileData.Stats.Attributes[item.TemplateID] = statValue
			}
		}

		for _, quest := range quests {
			var questAttributes map[string]interface{}
			json.Unmarshal([]byte(quest.Value), &questAttributes)
			profileData.Items[quest.TemplateID] = mcp.BaseItem{
				TemplateId: quest.TemplateID,
				Attributes: questAttributes,
				Quantity:   1,
			}
		}

		if profileId == "athena" && managers.IsDailyQuestCacheLoaded() {
			cacheManager := caching.GetDailyQuestCacheManager()
			var questManager accounts.Items
			var questManagerFound bool = true
			if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ?", accountId, "athena", "quest_manager").First(&questManager).Error; err != nil {
				questManagerFound = false
			}

			var questManagerValue map[string]interface{}

			if questManagerFound {
				var rawData interface{}
				if val, exists := profileData.Stats.Attributes["quest_manager"]; exists {
					rawData = val
				}

				if rawData != nil {
					jsonData, err := json.Marshal(rawData)
					if err != nil {
						log.Println("Error marshaling quest manager data:", err)
					} else {
						err = json.Unmarshal(jsonData, &questManagerValue)
						if err != nil {
							log.Println("Error unmarshaling quest manager data:", err)
						}
					}
				}
			}
			currentDate := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")

			if questManagerFound && questManagerValue != nil && questManagerValue["dailyLoginInterval"] != currentDate {
				questManagerValue["dailyLoginInterval"] = currentDate
				if val, ok := questManagerValue["dailyQuestRerolls"].(int); ok {
					questManagerValue["dailyQuestRerolls"] = val + 1
				} else {
					questManagerValue["dailyQuestRerolls"] = 1
				}

				skib, err := json.Marshal(questManagerValue)
				if err != nil {
					log.Println("Error marshaling updated quest manager data:", err)
				} else {
					if err := db.Model(&questManager).Where("account_id = ? AND profile_id = ? AND template_id = ?",
						accountId, "athena", "quest_manager").
						Update("value", skib).Error; err != nil {
						log.Println("Error updating quest manager in database:", err)
					}
					profileData.Stats.Attributes["quest_manager"] = questManagerValue
				}

				if quests, found := cacheManager.TryGet(); found {
					selectedQuests := quests[:3]

					for _, quest := range selectedQuests {

						questTemplateId := "Quest:" + quest.Name

						var existingQuest accounts.Quests
						questExists := false
						if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ? AND season = ? AND daily = ?",
							accountId, "athena", questTemplateId, ua.Season, true).First(&existingQuest).Error; err == nil {
							questExists = true
						}

						if questExists {
							continue
						}

						now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
						objectiveStates := make(map[string]interface{})
						for _, obj := range quest.Properties.Objectives {
							objectiveStates["completion_"+obj.BackendName] = obj.Stage
						}

						combinedAttributes := map[string]interface{}{
							"creation_time":                 now,
							"level":                         -1,
							"item_seen":                     true,
							"playlists":                     []string{},
							"sent_new_notification":         true,
							"challenge_bundle_id":           "",
							"xp_reward_scalar":              1,
							"challenge_linked_quest_given":  "",
							"quest_pool":                    "",
							"quest_state":                   "Active",
							"bucket":                        "",
							"last_state_change_time":        now,
							"challenge_linked_quest_parent": "",
							"max_level_bonus":               0,
							"xp":                            quest.Properties.SeasonXP,
							"quest_rarity":                  "uncommon",
							"favorite":                      false,
						}

						for key, value := range objectiveStates {
							combinedAttributes[key] = value
						}

						for key, value := range objectiveStates {
							combinedAttributes[key] = value
						}

						questUUID := strings.Replace(uuid.New().String(), "-", "", -1)
						attributesJSON, _ := json.Marshal(combinedAttributes)

						newQuest := accounts.Quests{
							UUID:       questUUID,
							AccountID:  accountId,
							ProfileID:  "athena",
							TemplateID: questTemplateId,
							Value:      string(attributesJSON),
							Daily:      true,
							Season:     ua.Season,
						}

						db.Create(&newQuest)

						profileData.Items[questTemplateId] = mcp.BaseItem{
							TemplateId: questTemplateId,
							Attributes: combinedAttributes,
							Quantity:   1,
						}
					}
				} else {
					fmt.Println("no quests found")
				}
			}
		}

		if profileId == "athena" && managers.IsMissionBundleCacheLoaded() {
			cacheManager := caching.GetChallengeBundleCacheManager()

			if bundles, found := cacheManager.TryGet(); found {
				scheduleToGrantedBundles := make(map[string][]string)
				var allQuestsToSave []accounts.Quests
				now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")

				for _, bundle := range bundles {
					season10Regex := regexp.MustCompile(`(?:MissionBundle_S\d+_(\d+)[A-Za-z]?)`)
					defaultRegex := regexp.MustCompile(`Week_(\d{3})`)

					var selectedRegex *regexp.Regexp
					if ua.Season == 10 {
						selectedRegex = season10Regex
					} else {
						selectedRegex = defaultRegex
					}

					match := selectedRegex.FindStringSubmatch(bundle.TemplateId)
					if match == nil {
						continue
					}

					var weekStr string
					if len(match) > 1 && match[1] != "" {
						weekStr = match[1]
					} else if len(match) > 2 && match[2] != "" {
						weekStr = match[2]
					} else {
						continue
					}

					questWeekNumber, err := strconv.Atoi(weekStr)
					if err != nil || questWeekNumber > utilities.GetConfig().Current_QuestWeek {
						continue
					}

					bundleName := "ChallengeBundle:" + bundle.TemplateId
					grantedQuestInstanceIds := []string{}
					challengeBundleScheduleId := bundle.ChallengeBundleSchedule

					scheduleToGrantedBundles[challengeBundleScheduleId] = append(
						scheduleToGrantedBundles[challengeBundleScheduleId],
						bundleName,
					)

					for _, quest := range bundle.Objects {
						if quest.Options.IsBattlePass {
							continue
						}

						questTemplateId := "Quest:" + quest.QuestDefinition

						var existingQuest accounts.Quests
						questExists := false
						if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ? AND season = ?",
							accountId, "athena", questTemplateId, ua.Season).First(&existingQuest).Error; err == nil {
							questExists = true
						}

						if questExists {
							continue
						}

						objectiveStates := make(map[string]interface{})
						for _, obj := range quest.Objectives {
							objectiveStates["completion_"+obj.BackendName] = obj.Stage
						}
						// brseason10
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

						for key, value := range objectiveStates {
							combinedAttributes[key] = value
						}

						grantedQuestInstanceIds = append(grantedQuestInstanceIds, questTemplateId)

						questUUID := strings.Replace(uuid.New().String(), "-", "", -1)
						attributesJSON, _ := json.Marshal(combinedAttributes)

						allQuestsToSave = append(allQuestsToSave, accounts.Quests{
							UUID:       questUUID,
							AccountID:  accountId,
							ProfileID:  "athena",
							Daily:      false,
							TemplateID: questTemplateId,
							Value:      string(attributesJSON),
							Season:     ua.Season,
						})

						profileData.Items[questTemplateId] = mcp.BaseItem{
							TemplateId: questTemplateId,
							Attributes: combinedAttributes,
							Quantity:   1,
						}
					}

					bundleExists := false
					var existingBundle accounts.Quests
					if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ? AND season = ?",
						accountId, "athena", bundleName, ua.Season).First(&existingBundle).Error; err == nil {
						bundleExists = true
					}

					if !bundleExists {
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

						allQuestsToSave = append(allQuestsToSave, accounts.Quests{
							UUID:       bundleUUID,
							AccountID:  accountId,
							ProfileID:  "athena",
							Daily:      false,
							TemplateID: bundleName,
							Value:      string(bundleJSON),
							Season:     ua.Season,
						})

						profileData.Items[bundleName] = mcp.BaseItem{
							TemplateId: bundleName,
							Attributes: bundleItemAttributes,
							Quantity:   1,
						}
					}
				}

				for scheduleId, grantedBundles := range scheduleToGrantedBundles {
					scheduleItemExists := false
					var existingSchedule accounts.Quests
					if err := db.Where("account_id = ? AND profile_id = ? AND template_id = ? AND season = ?",
						accountId, "athena", scheduleId, ua.Season).First(&existingSchedule).Error; err == nil {
						scheduleItemExists = true
					}

					if !scheduleItemExists {
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

						allQuestsToSave = append(allQuestsToSave, accounts.Quests{
							UUID:       scheduleUUID,
							AccountID:  accountId,
							ProfileID:  "athena",
							Daily:      false,
							TemplateID: scheduleId,
							Value:      string(scheduleJSON),
							Season:     ua.Season,
						})

						profileData.Items[scheduleId] = mcp.BaseItem{
							TemplateId: scheduleId,
							Attributes: scheduleItemAttributes,
							Quantity:   1,
						}
					}
				}

				if len(allQuestsToSave) > 0 {
					db.Create(&allQuestsToSave)
				}
			} else {
				fmt.Println("no bundles found")
			}
		}

		if user.HasFL && profileId == "athena" && IsCacheLoaded() {
			cache := GetCachedItems()
			cache.Range(func(templateID, baseItem interface{}) bool {
				if _, exists := profileData.Items[templateID.(string)]; !exists {
					profileData.Items[templateID.(string)] = baseItem.(mcp.BaseItem)
				}
				return true
			})
		}

		response = mcp.DefaultMCPResponse{
			ProfileRevision:            profile.Revision,
			ProfileId:                  profileId,
			ProfileChangesBaseRevision: profile.Revision,
			ProfileCommandRevision:     profile.Revision,
			ServerTime:                 now,
			ResponseVersion:            1,
			ProfileChanges: []map[string]interface{}{
				{
					"changeType": "fullProfileUpdate",
					"profile":    profileData,
				},
			},
		}
	}

	c.JSON(200, response)
}
