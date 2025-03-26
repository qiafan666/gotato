package gerr

import (
	"github.com/qiafan666/gotato/commons/gson"
)

func Marshal(v any) ([]byte, error) {
	m, err := gson.Marshal(v)
	return m, Wrap(err)
}

func Unmarshal(b []byte, v any) error {
	return Wrap(gson.Unmarshal(b, v))
}
