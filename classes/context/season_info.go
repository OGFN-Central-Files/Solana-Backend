package context

type SeasonInfo struct {
	OS     string `json:"os"`
	Season int    `json:"season"`
	Build  string `json:"build"`
	CL     string `json:"cl"`
	Lobby  string `json:"lobby"`
}
