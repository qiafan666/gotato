package gcommon

import (
	"encoding/json"
	"github.com/qiafan666/gotato/commons/gcast"
	"strings"
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

type TestStruct struct {
	Id    int
	Age   int
	Score float64
	Name  string
}
type TestStruct2 struct {
	Id    int
	Score float64
}

func TestSliceSort(t *testing.T) {
	var struct1 = TestStruct{Id: 2, Age: 20, Score: 80.5}
	var struct2 = TestStruct{Id: 1, Age: 18, Score: 90.5}
	var struct3 = TestStruct{Id: 3, Age: 22, Score: 70.5}

	var sliceStruct = []TestStruct{struct1, struct2, struct3}

	//排序规则，按照Id排序
	SliceSort(
		sliceStruct,
		func(i, j TestStruct) bool {
			// id大的在前
			if struct1.Id > struct2.Id {
				return true
			}
			return false
		})
	// 按照成绩排序
	SliceSort(
		sliceStruct,
		func(i, j TestStruct) bool {
			// 成绩大的在前
			if struct1.Score > struct2.Score {
				return true
			}
			return false
		})
	t.Log(sliceStruct)

	t.Log(SliceToMap(sliceStruct, func(val TestStruct) int {
		return val.Id
	}))

	t.Log(SliceToMapAny(sliceStruct, func(val TestStruct) (string, TestStruct2) {
		return gcast.ToString(val.Age), TestStruct2{Id: val.Id, Score: val.Score}
	}))

	t.Log(SliceToNilMap(sliceStruct))
}

func TestIf(t *testing.T) {

	t.Log(If(true, 1, 2))

}

func TestUnderlineCamelCase(t *testing.T) {
	var underLine = "user_name_score"
	t.Log(CamelName(underLine, false))
	t.Log(CamelName(underLine, true))
	t.Log(UnderscoreName(underLine))
}

func TestStringBuff(t *testing.T) {
	t.Log(NewBuffer().Append("hello", " ", "world").Append("world").String())
}

func TestSliceConvert(t *testing.T) {
	var struct1 = TestStruct{Id: 2, Age: 20, Score: 80.5}
	var struct2 = TestStruct{Id: 1, Age: 18, Score: 90.5}
	var struct3 = TestStruct{Id: 3, Age: 22, Score: 70.5}

	var sliceStruct = []TestStruct{struct1, struct2, struct3}

	toMap := SliceToMap(sliceStruct, func(val TestStruct) int {
		return val.Id
	})

	keys := MapKeys(toMap)
	t.Log(keys)

	convert := SliceConvert(keys, func(val int) string {
		return gcast.ToString(val)
	})
	t.Log(convert)
}

func TestSliceFilter(t *testing.T) {
	var struct1 = TestStruct{Id: 2, Age: 20, Score: 80.5, Name: "apple"}
	var struct2 = TestStruct{Id: 1, Age: 18, Score: 90.5, Name: "banana"}
	var struct3 = TestStruct{Id: 3, Age: 22, Score: 70.5, Name: "orange"}

	var sliceStruct = []TestStruct{struct1, struct2, struct3}

	//模糊查询
	filter := SliceFilter(sliceStruct, func(val TestStruct) (TestStruct, bool) {
		if strings.Contains(val.Name, "ange") {
			return val, true
		}
		return val, false
	})
	t.Log(filter)

}

func TestSliceBatch(t *testing.T) {
	var struct1 = TestStruct{Id: 2, Age: 20, Score: 80.5, Name: "apple"}
	var struct2 = TestStruct{Id: 1, Age: 18, Score: 90.5, Name: "banana"}
	var struct3 = TestStruct{Id: 3, Age: 22, Score: 70.5, Name: "orange"}

	var sliceStruct = []TestStruct{struct1, struct2, struct3}

	//分批处理
	batch := SliceBatch(sliceStruct, func(t TestStruct) TestStruct {
		t.Name = strings.ToUpper(t.Name)
		return t
	})
	t.Log(batch)
}

func TestSlice2String(t *testing.T) {

	var slice1 = []int{1, 2, 3, 4, 5}
	t.Log(Slice2String(slice1, ","))
	t.Log(String2Slice(Slice2String(slice1, ","), ","))

	var slice2 = []string{"apple", "banana", "orange"}
	t.Log(Slice2String(slice2, ","))
	t.Log(String2Slice(Slice2String(slice2, ","), ","))
}
