package managers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"solana/classes/quests"
	"solana/modules/caching"
	"solana/utilities"

	"github.com/fatih/color"
)

func CacheMissionBundles() (*[]quests.AthenaChallengeBundle, error) {
	var questBundle []quests.AthenaChallengeBundle

	file, err := os.ReadFile("Static/Memory/Season10/Quests/Battlepass/Missions.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	err = json.Unmarshal(file, &questBundle)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %v", err)
	}

	caching.GetChallengeBundleCacheManager().Set(questBundle)

	utilities.LogWithTimestamp(color.GreenString, "Cached quests")

	return &questBundle, nil
}

func CacheDailyQuests() (*[]quests.AthenaDailyQuest, error) {
	var allQuests []quests.AthenaDailyQuest
	directory := "Static/Memory/Season10/Quests/Repeatable/"

	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(directory, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			utilities.LogError(fmt.Sprintf("failed to read file %s: %v", filePath, err))
			continue
		}

		var quests []quests.AthenaDailyQuest
		if err := json.Unmarshal(content, &quests); err != nil {
			utilities.LogError(fmt.Sprintf("failed to unmarshal file %s: %v", filePath, err))
			continue
		}

		allQuests = append(allQuests, quests...)
	}

	caching.GetDailyQuestCacheManager().Set(allQuests)

	utilities.LogWithTimestamp(color.GreenString, fmt.Sprintf("Cached %d daily quests from %d files", len(allQuests), len(files)))

	return &allQuests, nil
}

func IsMissionBundleCacheLoaded() bool {
	return caching.GetChallengeBundleCacheManager().IsLoaded()
}

func IsDailyQuestCacheLoaded() bool {
	return caching.GetDailyQuestCacheManager().IsLoaded()
}
