package gerr

import (
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(v any) ([]byte, error) {
	m, err := json.Marshal(v)
	return m, Wrap(err)
}

func Unmarshal(b []byte, v any) error {
	return Wrap(json.Unmarshal(b, v))
}
