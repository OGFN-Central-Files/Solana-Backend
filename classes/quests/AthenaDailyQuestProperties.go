package quests

type AthenaDailyQuestProperties struct {
	DisplayName       string
	Description       string
	SeasonXP          int32
	SeasonBattleStars int32
	Objectives        []AthenaDailyQuestObjectives
}
