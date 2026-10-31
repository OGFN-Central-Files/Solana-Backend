package caching

import (
	"solana/classes/quests"
	"sync"
	"time"
)

type DailyQuestCacheManager struct {
	cache      sync.Map
	expiration time.Duration
}

var questCache *DailyQuestCacheManager
var dailyOnce sync.Once

func GetDailyQuestCacheManager() *DailyQuestCacheManager {
	dailyOnce.Do(func() {
		questCache = &DailyQuestCacheManager{
			expiration: 7 * 24 * time.Hour,
		}
	})
	return questCache
}

func (c *DailyQuestCacheManager) TryGet() ([]quests.AthenaDailyQuest, bool) {
	if val, found := c.cache.Load("DailyQuest"); found {
		if quests, ok := val.([]quests.AthenaDailyQuest); ok {
			return quests, true
		}
	}
	return nil, false
}

func (c *DailyQuestCacheManager) Set(quests []quests.AthenaDailyQuest) {
	c.cache.Store("DailyQuest", quests)
	go func() {
		time.Sleep(c.expiration)
		c.Remove()
	}()
}

func (c *DailyQuestCacheManager) Remove() {
	c.cache.Delete("DailyQuest")
}

func (c *DailyQuestCacheManager) IsLoaded() bool {
	_, found := c.cache.Load("DailyQuest")
	return found
}
