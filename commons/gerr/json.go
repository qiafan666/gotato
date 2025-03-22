package gerr

import (
	"github.com/qiafan666/gotato/commons/gjson"
)

func Marshal(v any) ([]byte, error) {
	m, err := gjson.Marshal(v)
	return m, Wrap(err)
}

func Unmarshal(b []byte, v any) error {
	return Wrap(gjson.Unmarshal(b, v))
}
