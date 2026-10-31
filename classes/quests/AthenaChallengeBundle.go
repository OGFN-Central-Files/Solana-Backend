package quests

type AthenaChallengeBundle struct {
	TemplateId              string
	ChallengeBundleSchedule string
	CompletionReward        []AthenaCompletionRewards
	Objects                 []AthenaChallengeBundleObjects
}
