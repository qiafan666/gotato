package gcommon

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gid"
	"slices"
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
			// 比较切片中的两个元素的 Id
			return i.Id > j.Id
		})
	t.Log("id排序后的sliceStruct:", sliceStruct)

	// 按照成绩排序
	SliceSort(
		sliceStruct,
		func(i, j TestStruct) bool {
			// 成绩大的在前
			return i.Score > j.Score
		})
	t.Log("成绩排序后的sliceStruct:", sliceStruct)

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

func TestSlice2Str(t *testing.T) {

	var slice1 = []int{1, 2, 3, 4, 5}
	t.Log(Slice2Str(slice1, ","))
	t.Log(Str2Slice(Slice2Str(slice1, ","), ","))

	var slice2 = []string{"apple", "banana", "orange"}
	t.Log(Slice2Str(slice2, ","))
	t.Log(Str2Slice(Slice2Str(slice2, ","), ","))
}

type Test1 struct {
	Id []int `json:"id"`
}

func TestKv2Str(t *testing.T) {
	t.Log(Kv2Str("msg", "key1", "value1", "key2", "value2"))

	test1 := Test1{Id: []int{1, 2, 3}}
	t.Log(test1)

	t.Log(slices.DeleteFunc(test1.Id, func(val int) bool {
		return val == 2
	}))
}

// generateKey 使用移位方式生成唯一键
func generateKey(roleId1, roleId2 int64) int64 {
	if roleId1 > roleId2 {
		return (roleId1 << 32) | roleId2
	}
	return (roleId2 << 32) | roleId1
}

func TestInt64(t *testing.T) {

	// 测试用例1: roleId1 < roleId2
	roleId1 := int64(123456789)
	roleId2 := int64(987654321)
	key1 := generateKey(roleId1, roleId2)
	key2 := generateKey(roleId2, roleId1)

	if key1 != key2 {
		t.Errorf("Expected key to be equal for (roleId1=%d, roleId2=%d) and (roleId1=%d, roleId2=%d), got key1=%d, key2=%d", roleId1, roleId2, roleId2, roleId1, key1, key2)
	}

	// 测试用例2: roleId1 > roleId2
	roleId1 = int64(987654321)
	roleId2 = int64(123456789)
	key3 := generateKey(roleId1, roleId2)

	if key3 != key1 {
		t.Errorf("Expected same key for reversed inputs, got key1=%d, key3=%d", key1, key3)
	}

	// 测试用例3: roleId1 = roleId2
	roleId1 = int64(123456789)
	roleId2 = int64(123456789)
	key4 := generateKey(roleId1, roleId2)

	if key4 != (roleId1<<32 | roleId2) {
		t.Errorf("Expected key to be %d for equal roleId1 and roleId2, got %d", (roleId1<<32 | roleId2), key4)
	}

	// 测试用例4: roleId1 和 roleId2 较小的值
	roleId1 = int64(1)
	roleId2 = int64(2)
	key5 := generateKey(roleId1, roleId2)
	expectedKey5 := (roleId2 << 32) | roleId1

	if key5 != expectedKey5 {
		t.Errorf("Expected key to be %d, got %d", expectedKey5, key5)
	}

	// 测试用例5: roleId1 和 roleId2 较大的值
	roleId1 = int64(9223372036854775807) // Max int64
	roleId2 = int64(1)
	key6 := generateKey(roleId1, roleId2)
	expectedKey6 := (roleId1 << 32) | roleId2

	if key6 != expectedKey6 {
		t.Errorf("Expected key to be %d, got %d", expectedKey6, key6)
	}
}

type Req struct {
	SendID    string `json:"send_id"        validate:"required"`
	RequestID string `json:"request_id"   validate:"required"`
	GrpID     uint8  `json:"grp_id" validate:"required"` // 消息组id
	CmdID     uint8  `json:"cmd_id" validate:"required"` // 消息的ID
	Data      []byte `json:"data"`
}

func TestEncode(t *testing.T) {
	req := &Req{
		SendID:    "123456",
		RequestID: "abcdefg",
		GrpID:     2,
		CmdID:     2,
		Data:      []byte("hello world"),
	}

	encoder := NewGobEncoder()
	encode, err := encoder.Encode(req)
	if err != nil {
		t.Error(err)
	}
	t.Log(encode)
	t.Log(base64.StdEncoding.EncodeToString(encode))

	decodeString, err := base64.StdEncoding.DecodeString(base64.StdEncoding.EncodeToString(encode))
	if err != nil {
		t.Error(err)
	}
	var decode *Req
	err = encoder.Decode(decodeString, &decode)
	if err != nil {
		t.Error(err)
	}
	t.Log(decode)
	t.Log(string(decode.Data))
}

func TestSnowflake(t *testing.T) {
	t.Log(gid.RandSnowflakeID())
	t.Log(gid.RandServerID())
	t.Log(RangeNum[float64](1.4, 11001.1214))
	t.Log(RangeNum[float32](1.4, 11001.1214))
	t.Log(RangeNum[int32](1, 100))
}

func TestAppendString(t *testing.T) {
	t.Log(AppendStr("hello", "world").Append("test"))
	t.Log(BuildStrWithSep("-", "hello", "world", "test", 1, 1231.1231))
}

func TestSlice(t *testing.T) {
	var slice = []int{1, 2, 3, 4, 5}
	SliceForEach(slice, func(val int) {
		t.Log(val)
	})
	SliceReverse(slice)
	t.Log(slice)
	SliceSort(slice, func(i, j int) bool {
		return i < j
	})
	t.Log(slice)

	SliceForEach(slice, func(val int) {
		t.Log(val)
	})
	SliceForEachReverse(slice, func(val int) {
		t.Log(val)
	})
}

func TestHeadSort(t *testing.T) {
	// 示例 1：对整数进行升序排序。
	intData := []int{5, 3, 8, 4, 2, 1}
	sortedInts := HeapSort(intData, func(a, b int) bool {
		return a < b // 升序排序
	})
	fmt.Println("排序后的整数:", sortedInts)

	// 示例 2：按字符串长度排序。
	stringData := []string{"apple", "banana", "kiwi", "grape", "orange"}
	sortedStrings := HeapSort(stringData, func(a, b string) bool {
		return len(a) < len(b) // 按字符串长度排序
	})
	fmt.Println("按长度排序后的字符串:", sortedStrings)

	// 示例 3：对结构体进行多条件排序。
	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 30},
		{"Dave", 20},
	}

	sortedPeople := HeapSort(people, func(a, b Person) bool {
		if a.Age == b.Age {
			return a.Name < b.Name // 如果年龄相同，则按名字排序
		}
		return a.Age < b.Age // 否则按年龄排序
	})
	fmt.Println("按年龄和名字排序后的人员:", sortedPeople)
}

func TestMap(t *testing.T) {

	testMap := map[string]int{
		"banana": 2,
		"apple":  1,
		"orange": 3,
		"123":    4,
		"宁":      5,
	}
	t.Log(MapSortKey(testMap, func(a, b string) bool {
		return a < b
	}))
	t.Log(MapSortValue(testMap, func(a, b int) bool {
		return a < b
	}))
	t.Log(MapKeys(testMap))

	testMap2 := map[string]int{
		"banana": 3,
	}
	e, err := MapMergeE(testMap, testMap2)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(e)
}

func TestSliceDelete(t *testing.T) {
	testSlice := []int{1, 2, 3, 4, 5}
	sliceDelete := SliceDelete(testSlice, 2)
	t.Log(sliceDelete)

	t.Log(SliceIndex(sliceDelete, 6))
	newList := SliceDeleteIndex(sliceDelete, SliceIndex(sliceDelete, 6))
	t.Log(newList)
}

func TestHideStr(t *testing.T) {
	hideStr := HideStr("hg2gggwwwdwvw")
	t.Log(hideStr)
	t.Log(RandRedPacket(10, 100))
	t.Log(RandByWeight([]int32{14121, 21130, 30}))
}

func TestRandStr(t *testing.T) {
	t.Log(RandLower(10))
	t.Log(RandUpper(10))
	t.Log(RandNum(10))
	t.Log(RandSymbol(10))
	t.Log(RandLowerUpper(10))
	t.Log(RandStr(10))
	t.Log(RandCusStr(Symbol+Number, 10))
}

func TestContainsEmoji(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"Hello, world! 😊", true},
		{"No emojis here.", false},
		{"🚀 Rocket to the moon!", true},
		{"Some text and 🐱 a cat emoji.", true},
		{"Just a smiley face :)", false},
		{"Multiple emojis 😂😂", true},
		{"Mix of text and emojis: 💡✨🌟", true},
		{"含有汉字和表情 😊", true},
		{"含有汉字但无表情", false},
		{"芙蓉王⚪圈", true},
	}

	for _, test := range tests {
		result := ContainsEmoji(test.input)
		if result != test.expected {
			t.Errorf("containsEmoji(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}
