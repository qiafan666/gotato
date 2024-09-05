package jsonutil

import (
	"encoding/json"
	"github.com/qiafan666/gotato/commons/gerr"
)

func JsonMarshal(v any) ([]byte, error) {
	m, err := json.Marshal(v)
	return m, gerr.Wrap(err)
}

func JsonUnmarshal(b []byte, v any) error {
	return gerr.Wrap(json.Unmarshal(b, v))
}

func StructToJsonString(param any) string {
	dataType, _ := JsonMarshal(param)
	dataString := string(dataType)
	return dataString
}

// The incoming parameter must be a pointer
func JsonStringToStruct(s string, args any) error {
	err := json.Unmarshal([]byte(s), args)
	return err
}
