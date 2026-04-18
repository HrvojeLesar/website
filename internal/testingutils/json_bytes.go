package testingutils

import "encoding/json"

func JsonBytes(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
