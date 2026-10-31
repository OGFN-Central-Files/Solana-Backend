package quests

type AthenaDailyQuest struct {
	Type       string                     `json:"Type"`
	Name       string                     `json:"Name"`
	Class      string                     `json:"Class"`
	Properties AthenaDailyQuestProperties `json:"Properties"` 
}
