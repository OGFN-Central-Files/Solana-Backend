package models

type Item struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Set  struct {
		Value        string `json:"value"`
		Text         string `json:"text"`
		BackendValue string `json:"backendValue"`
	} `json:"set"`
}
