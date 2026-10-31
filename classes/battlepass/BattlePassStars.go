package battlepass

import (
	"encoding/json"
	"errors"
)

type ItemDefinition struct {
	AssetPathName string `json:"AssetPathName"`
	SubPathString string `json:"SubPathString"`
}

type BattlePassStars struct {
	Level          int             `json:"Level"`
	Quantity       int             `json:"Quantity"`
	ItemDefinition json.RawMessage `json:"ItemDefinition"`
}

func (b *BattlePassStars) UnmarshalJSON(data []byte) error {
	type Alias BattlePassStars
	aux := &struct {
		ItemDefinition interface{} `json:"ItemDefinition"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.ItemDefinition.(type) {
	case float64:
		b.ItemDefinition = json.RawMessage([]byte(`{"AssetPathName":"","SubPathString":""}`))
	case map[string]interface{}:
		b.ItemDefinition, _ = json.Marshal(v)
	default:
		return errors.New("unexpected type for ItemDefinition")
	}

	return nil
}
