package quests

import "time"

type AthenaQuest struct {
	CreationTime               time.Time `json:"creation_time"`
	Level                      int       `json:"level"`
	ItemSeen                   bool      `json:"item_seen"`
	Playlists                  []string  `json:"playlists"`
	SentNewNotification        bool      `json:"sent_new_notification"`
	ChallengeBundleID          string    `json:"challenge_bundle_id"`
	XPRewardScalar             float64   `json:"xp_reward_scalar"`
	ChallengeLinkedQuestGiven  string    `json:"challenge_linked_quest_given"`
	QuestPool                  string    `json:"quest_pool"`
	QuestState                 string    `json:"quest_state"`
	Bucket                     string    `json:"bucket"`
	LastStateChangeTime        time.Time `json:"last_state_change_time"`
	ChallengeLinkedQuestParent string    `json:"challenge_linked_quest_parent"`
	MaxLevelBonus              int       `json:"max_level_bonus"`
	XP                         int       `json:"xp"`
	QuestRarity                string    `json:"quest_rarity"`
	Favorite                   bool      `json:"favorite"`
}
