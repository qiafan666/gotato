package gson

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func Get(data string, path string) string {
	return gjson.Get(data, path).String()
}
