package store

import "encoding/json"

func unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
