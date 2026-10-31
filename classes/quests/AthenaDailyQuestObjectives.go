package quests

type AthenaDailyQuestObjectives struct {
	BackendName            string
	ObjectiveState         string
	ItemEvent              string
	ItemReference          string
	ItemTemplateIdOverride string
	Description            string
	HudShortDescription    string
	Count                  int32
	Stage                  int32
	bHidden                bool
}
