package dsmcp

import (
	"encoding/json"
	"net/http"
	"solana/classes/mcp"
	"solana/classes/quests"
	"solana/managers"
	"solana/modules/caching"
	"solana/modules/database"
	"solana/modules/database/entities/accounts"
	"solana/utilities"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func UpdateQuests(c *gin.Context) {
	accountId := c.Param("accountId")
	profileId := c.Query("profileId")

	db := database.Get()

	if profileId == "" {
		utilities.Internal.ValidationFailed().
			WithIntent(utilities.Prod).Apply(c.Writer)
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

	var body []quests.AthenaUpdateQuests
	if err := c.ShouldBindJSON(&body); err != nil {
		utilities.MCP.InvalidPayload().
			WithIntent(utilities.Prod).
			Apply(c.Writer)
		return
	}

	var profile accounts.Profiles
	var quests []accounts.Quests
	profileFound := true

	if err := db.Where("account_id = ? AND profile_id = ?", accountId, profileId).First(&profile).Error; err != nil {
		profileFound = false
	} else if profileFound {
		db.Where("account_id = ? AND profile_id = ?", accountId, profileId).Find(&quests)
	}

	questMap := make(map[string]accounts.Quests)
	for _, q := range quests {
		questMap[q.TemplateID] = q
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

		profileChanges := make([]map[string]interface{}, 0)

		cacheManager := caching.GetDailyQuestCacheManager()

		cachedQuests, hasCachedQuests := cacheManager.TryGet()

		for _, quest := range body {
			template, exists := questMap["Quest:"+quest.BackendName]
			if !exists {
				continue
			}

			if strings.Contains(quest.BackendName, "AthenaDailyQuest") && managers.IsDailyQuestCacheLoaded() {
				var questData map[string]interface{}
				if err := json.Unmarshal([]byte(template.Value), &questData); err != nil {
					continue
				}

				if !hasCachedQuests {
					continue
				}

				found := false
				for _, cachedQuest := range cachedQuests {
					if cachedQuest.Name == quest.BackendName && questData["quest_state"] != "Completed" {
						found = true

						objectiveStates := make(map[string]int)

						for _, obj := range cachedQuest.Properties.Objectives {
							objectiveStates["completion_"+obj.BackendName] = int(obj.Count)
						}

						completionKey := "completion_" + quest.Objective

						questCompletion, ok := objectiveStates[completionKey]
						if !ok {
							questCompletion = 0
						}

						if questCompletion <= objectiveStates[completionKey] {
							questCompletion = int(quest.Count)
							questData[completionKey] = questCompletion
						}

						if questCompletion >= objectiveStates[completionKey] {
							updatedValue, err := json.Marshal(questData)
							if err != nil {
								continue
							}
							template.Value = string(updatedValue)
							profileChanges = append(profileChanges, map[string]interface{}{
								"changeType": "questCompleted",
								"quest":      questData,
							})
						} else {
							updatedValue, err := json.Marshal(questData)
							if err != nil {
								continue
							}
							template.Value = string(updatedValue)
							profileChanges = append(profileChanges, map[string]interface{}{
								"changeType": "questUpdated",
								"quest":      questData,
							})
							db.Save(&template)
						}
					}
				}

				if !found {
				}
			} else if strings.Contains(quest.BackendName, "Quest_BR") && !strings.Contains(quest.BackendName, "Quest_BR_Daily") && managers.IsMissionBundleCacheLoaded() {
				var questData map[string]interface{}
				if err := json.Unmarshal([]byte(template.Value), &questData); err != nil {
					utilities.LogInfo("Failed to parse quest template", "error", err, "questName", quest.BackendName)
					return
				}

				cacheManager := caching.GetChallengeBundleCacheManager()

				if bundles, found := cacheManager.TryGet(); found {
					for _, bundle := range bundles { // should be in a bundle if not then its most likely a daily?
						for _, bundleQuest := range bundle.Objects {
							if bundleQuest.QuestDefinition == quest.BackendName {
								objectiveStates := make(map[string]int)

								for _, obj := range bundleQuest.Objectives {
									objectiveStates["completion_"+obj.BackendName] = obj.Count
								}

								completionKey := "completion_" + quest.Objective // ex: completion_battlepass_use_item_zipline

								targetObjectiveCount, exists := objectiveStates[completionKey]
								if !exists {
									continue
								}

								_, ok := questData[completionKey].(float64)
								if !ok {
									utilities.Basic.BadRequest().Apply(c.Writer)
									return
								}

								newProgress := quest.Count
								questData[completionKey] = newProgress

								if float64(newProgress) >= float64(targetObjectiveCount) { // prob dont need to do float64 but wtv
									questData["quest_state"] = "Completed"
									updatedValue, err := json.Marshal(questData)
									if err != nil {
										continue
									}
									template.Value = string(updatedValue)
									profileChanges = append(profileChanges, map[string]interface{}{
										"changeType": "questCompleted",
										"quest":      questData,
									})
								} else {
									questData["quest_state"] = "Active" // we do this just incase quest for some reason gets reverted?
									updatedValue, err := json.Marshal(questData)
									if err != nil {
										continue
									}
									template.Value = string(updatedValue)
									profileChanges = append(profileChanges, map[string]interface{}{
										"changeType": "questUpdated",
										"quest":      questData,
									})
								}

								db.Save(&template)
							}
						}
					}
				}
			}
		}

		response = mcp.DefaultMCPResponse{
			ProfileRevision:            profile.Revision,
			ProfileId:                  profileId,
			ProfileChangesBaseRevision: profile.Revision - 1,
			ProfileCommandRevision:     profile.Revision,
			ServerTime:                 now,
			ResponseVersion:            1,
			ProfileChanges:             profileChanges,
		}
	}

	c.JSON(http.StatusOK, response)
}
