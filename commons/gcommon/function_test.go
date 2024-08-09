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

func TestSliceContains(t *testing.T) {
	arr := []string{"apple", "banana", "orange"}
	t.Log(SliceContain(arr, "banana"))
	t.Log(SliceContains(arr, []string{"apple1", "orange2"}))
	t.Log(SliceContains(arr, []string{"apple", "orange"}))
}

func TestFunc(t *testing.T) {
	t.Log(GenerateUUID())
	t.Log(len(GenerateUUID()))
}
