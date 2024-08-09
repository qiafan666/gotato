package gcommon

import (
	"encoding/json"
	"testing"
)

type Test struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestStructToMap(t *testing.T) {

	test := Test{
		Name: "John",
		Age:  30,
	}

	toMap := StructToMap(test)
	marshal, err := json.Marshal(toMap)
	if err != nil {
		return
	}
	t.Log(string(marshal))
}
