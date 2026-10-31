package managers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"solana/classes/battlepass"
	"solana/modules/caching"
	"solana/utilities"

	"github.com/fatih/color"
)

func CacheBattlePass() (*[]battlepass.BattlePass, error) {
	var allBattlePasses []battlepass.BattlePass

	for season := 1; season <= 34; season++ {
		dirPath := fmt.Sprintf("Static/Memory/Season%d/", season)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		path := fmt.Sprintf("%s/bp.json", dirPath)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		file, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read file %s: %v", path, err)
			return nil, fmt.Errorf("failed to read file %s: %v", path, err)
		}

		file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

		var bp battlepass.BattlePass
		err = json.Unmarshal(file, &bp)
		if err != nil {
			log.Printf("Failed to unmarshal file %s: %v", path, err)
			return nil, fmt.Errorf("failed to unmarshal file %s: %v", path, err)
		}

		allBattlePasses = append(allBattlePasses, bp)
		caching.GetBattlePassCacheManager().Set(fmt.Sprintf("BRSeason%d", season), bp)
	}

	utilities.LogWithTimestamp(color.GreenString, "Cached all battle passes")

	return &allBattlePasses, nil
}

func IsBattlePassCacheLoaded() bool {
	return caching.GetBattlePassCacheManager().IsLoaded("BRSeason10")
}

func CacheBattlePassStars() (*[]battlepass.BattlePassStars, error) {
	var allBattlePassStars []battlepass.BattlePassStars

	for season := 1; season <= 34; season++ {
		dirPath := fmt.Sprintf("Static/Memory/Season%d/", season)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		path := fmt.Sprintf("%s/stars.json", dirPath)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		file, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read file %s: %v", path, err)
			return nil, fmt.Errorf("failed to read file %s: %v", path, err)
		}

		file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

		var Stars []battlepass.BattlePassStars
		err = json.Unmarshal(file, &Stars)
		if err != nil {
			log.Printf("Failed to unmarshal file %s: %v", path, err)
			return nil, fmt.Errorf("failed to unmarshal file %s: %v", path, err)
		}

		allBattlePassStars = append(allBattlePassStars, Stars...)
		caching.GetBattlePassCacheManager().Set(fmt.Sprintf("BRSeason%dStars", season), Stars)
	}

	utilities.LogWithTimestamp(color.GreenString, "Cached all battle pass stars")

	return &allBattlePassStars, nil
}

func IsBattlePassStarsCacheLoaded() bool {
	return caching.GetBattlePassCacheManager().IsLoaded("BRSeason10Stars")
}

func CacheCosmeticVariantTokens() (*[]battlepass.CosmeticVariantToken, error) {
	var allCVT []battlepass.CosmeticVariantToken

	dirPath := "Static/Athena"
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory %s does not exist: %v", dirPath, err)
	}

	path := fmt.Sprintf("%s/cvt.json", dirPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist: %v", path, err)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read file %s: %v", path, err)
		return nil, fmt.Errorf("failed to read file %s: %v", path, err)
	}

	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	var CVT map[string]battlepass.CosmeticVariantToken
	err = json.Unmarshal(file, &CVT)
	if err != nil {
		log.Printf("Failed to unmarshal file %s: %v", path, err)
		return nil, fmt.Errorf("failed to unmarshal file %s: %v", path, err)
	}

	for key, token := range CVT {
		allCVT = append(allCVT, token)
		caching.GetBattlePassCacheManager().Set(fmt.Sprintf("CosmeticVariantToken:%s", key), token)
	}

	utilities.LogWithTimestamp(color.GreenString, "Cached all cosmetic variant tokens")

	return &allCVT, nil
}

func IsCosmeticVariantTokenCacheLoaded() bool {
	return caching.GetBattlePassCacheManager().IsLoaded("CosmeticVariantToken:VTID_002_CarbideBlue_Stage1")
}

func CacheAthenaSeasonXPCurve() (*[]battlepass.AthenaSeasonXPCurve, error) {
	var allXPCurves []battlepass.AthenaSeasonXPCurve

	for season := 1; season <= 34; season++ {
		dirPath := fmt.Sprintf("Static/Memory/Season%d/", season)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		path := fmt.Sprintf("%s/xp.json", dirPath)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		file, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read file %s: %v", path, err)
			return nil, fmt.Errorf("failed to read file %s: %v", path, err)
		}

		file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

		var Curve []battlepass.AthenaSeasonXPCurve
		err = json.Unmarshal(file, &Curve)
		if err != nil {
			log.Printf("Failed to unmarshal file %s: %v", path, err)
			return nil, fmt.Errorf("failed to unmarshal file %s: %v", path, err)
		}

		allXPCurves = append(allXPCurves, Curve...)
		caching.GetBattlePassCacheManager().Set(fmt.Sprintf("AthenaSeason%dXPCurve", season), Curve)
	}

	utilities.LogWithTimestamp(color.GreenString, "Cached all Athena season XP curves")

	return &allXPCurves, nil
}

func IsAthenaSeasonXPCurveCacheLoaded() bool {
	return caching.GetBattlePassCacheManager().IsLoaded("AthenaSeason10XPCurve")
}
