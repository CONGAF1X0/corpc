package serializer

import "encoding/json"

type JsonSerializer struct{}

func (_ JsonSerializer) Marshal(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}

func (_ JsonSerializer) Unmarshal(data []byte, message interface{}) error {
	return json.Unmarshal(data, message)
}
