package caching

import (
	"solana/classes/quests"
	"sync"
	"time"
)

type ChallengeBundleCacheManager struct {
	cache      sync.Map
	expiration time.Duration
}

var challengeCache *ChallengeBundleCacheManager
var once sync.Once

func GetChallengeBundleCacheManager() *ChallengeBundleCacheManager {
	once.Do(func() {
		challengeCache = &ChallengeBundleCacheManager{
			expiration: 7 * 24 * time.Hour,
		}
	})
	return challengeCache
}

func (c *ChallengeBundleCacheManager) TryGet() ([]quests.AthenaChallengeBundle, bool) {
	if val, found := c.cache.Load("AthenaChallengeBundle"); found {
		if bundle, ok := val.([]quests.AthenaChallengeBundle); ok {
			return bundle, true
		}
	}
	return nil, false
}

func (c *ChallengeBundleCacheManager) Set(bundle []quests.AthenaChallengeBundle) {
	c.cache.Store("AthenaChallengeBundle", bundle)
	go func() {
		time.Sleep(c.expiration)
		c.Remove()
	}()
}

func (c *ChallengeBundleCacheManager) Remove() {
	c.cache.Delete("AthenaChallengeBundle")
}

func (c *ChallengeBundleCacheManager) IsLoaded() bool {
	_, found := c.cache.Load("AthenaChallengeBundle")
	return found
}
