package caching

import (
	"sync"
	"time"
)

type BattlePassCacheManager struct {
	cache      sync.Map
	expiration time.Duration
}

var bpCache *BattlePassCacheManager
var bpOnce sync.Once

func GetBattlePassCacheManager() *BattlePassCacheManager {
	bpOnce.Do(func() {
		bpCache = &BattlePassCacheManager{
			expiration: 7 * 24 * time.Hour,
		}
	})
	return bpCache
}

func (b *BattlePassCacheManager) Set(key string, value interface{}) {
	b.cache.Store(key, value)
}

func (b *BattlePassCacheManager) Get(key string) (interface{}, bool) {
	return b.cache.Load(key)
}

func (b *BattlePassCacheManager) Remove(key string) {
	b.cache.Delete(key)
}

func (b *BattlePassCacheManager) IsLoaded(key string) bool {
	_, found := b.cache.Load(key)
	return found
}
