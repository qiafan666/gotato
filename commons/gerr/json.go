package gerr

import "github.com/qiafan666/gotato/commons/gcommon"

func Marshal(v any) ([]byte, error) {
	m, err := gcommon.Marshal(v)
	return m, Wrap(err)
}

func Unmarshal(b []byte, v any) error {
	return Wrap(gcommon.Unmarshal(b, v))
}
