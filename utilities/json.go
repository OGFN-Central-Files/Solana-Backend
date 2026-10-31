package utilities

import (
	"encoding/json"
	"log"
)

func MustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Error marshaling: %v", err)
	}
	return data
}
