package gerr

import (
	"encoding/json"
)

func Marshal(v any) ([]byte, error) {
	m, err := json.Marshal(v)
	return m, Wrap(err)
}

func Unmarshal(b []byte, v any) error {
	return Wrap(json.Unmarshal(b, v))
}
