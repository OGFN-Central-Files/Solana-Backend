package battlepass

type AthenaSeasonXPCurve struct {
	Type      string `json:"Type"`
	Name      string `json:"Name"`
	Class     string `json:"Class"`
	Flags     string `json:"Flags"`
	RowStruct struct {
		ObjectName string `json:"ObjectName"`
		ObjectPath string `json:"ObjectPath"`
	} `json:"Properties.RowStruct"`
	Rows map[string]struct {
		Level         int `json:"Level"`
		XpToNextLevel int `json:"XpToNextLevel"`
		XpTotal       int `json:"XpTotal"`
	} `json:"Rows"`
}
