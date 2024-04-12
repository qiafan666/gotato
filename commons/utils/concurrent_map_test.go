package utils

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"testing"
)

type Animal struct {
	name string
}

func TestMapCreation(t *testing.T) {
	m := NewCMap[string]()
	if m.shards == nil {
		t.Error("map is null.")
	}

	if m.Count() != 0 {
		t.Error("new map should be empty.")
	}
}

func TestInsert(t *testing.T) {
	m := NewCMap[Animal]()
	elephant := Animal{"elephant"}
	monkey := Animal{"monkey"}

	m.Set("elephant", elephant)
	m.Set("monkey", monkey)

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}
}

func TestInsertAbsent(t *testing.T) {
	m := NewCMap[Animal]()
	elephant := Animal{"elephant"}
	monkey := Animal{"monkey"}

	m.Set("elephant", elephant)
	if _, ok := m.SetIfAbsent("monkey", monkey); !ok {
		t.Log("not exist returned false")
	}
	if value, ok := m.SetIfAbsent("elephant", elephant); ok {
		t.Log(value.name)
	}
}

func TestGet(t *testing.T) {
	m := NewCMap[Animal]()

	// Get a missing element.
	val, ok := m.Get("Money")

	if ok == true {
		t.Error("ok should be false when item is missing from map.")
	}

	if (val != Animal{}) {
		t.Error("Missing values should return as null.")
	}

	elephant := Animal{"elephant"}
	m.Set("elephant", elephant)

	// Retrieve inserted element.
	elephant, ok = m.Get("elephant")
	if ok == false {
		t.Error("ok should be true for item stored within the map.")
	}

	if elephant.name != "elephant" {
		t.Error("item was modified.")
	}
}

func TestHas(t *testing.T) {
	m := NewCMap[Animal]()

	// Get a missing element.
	if m.Has("Money") == true {
		t.Error("element shouldn't exists")
	}

	elephant := Animal{"elephant"}
	m.Set("elephant", elephant)

	if m.Has("elephant") == false {
		t.Error("element exists, expecting Has to return True.")
	}
}

func TestRemove(t *testing.T) {
	m := NewCMap[Animal]()

	monkey := Animal{"monkey"}
	m.Set("monkey", monkey)

	m.Remove("monkey")

	if m.Count() != 0 {
		t.Error("Expecting count to be zero once item was removed.")
	}

	temp, ok := m.Get("monkey")

	if ok != false {
		t.Error("Expecting ok to be false for missing items.")
	}

	if (temp != Animal{}) {
		t.Error("Expecting item to be nil after its removal.")
	}

	m.Remove("noone")
}

func TestRemoveCb(t *testing.T) {
	m := NewCMap[Animal]()

	monkey := Animal{"monkey"}
	m.Set("monkey", monkey)
	elephant := Animal{"elephant"}
	m.Set("elephant", elephant)

	var (
		mapKey   string
		mapVal   Animal
		wasFound bool
	)
	cb := func(key string, val Animal, exists bool) bool {
		mapKey = key
		mapVal = val
		wasFound = exists

		return val.name == "monkey"
	}

	// Monkey should be removed
	result := m.RemoveCb("monkey", cb)
	if !result {
		t.Errorf("Result was not true")
	}

	if mapKey != "monkey" {
		t.Error("Wrong key was provided to the callback")
	}

	if mapVal != monkey {
		t.Errorf("Wrong value was provided to the value")
	}

	if !wasFound {
		t.Errorf("Key was not found")
	}

	if m.Has("monkey") {
		t.Errorf("Key was not removed")
	}

	// Elephant should not be removed
	result = m.RemoveCb("elephant", cb)
	if result {
		t.Errorf("Result was true")
	}

	if mapKey != "elephant" {
		t.Error("Wrong key was provided to the callback")
	}

	if mapVal != elephant {
		t.Errorf("Wrong value was provided to the value")
	}

	if !wasFound {
		t.Errorf("Key was not found")
	}

	if !m.Has("elephant") {
		t.Errorf("Key was removed")
	}

	// Unset key should remain unset
	result = m.RemoveCb("horse", cb)
	if result {
		t.Errorf("Result was true")
	}

	if mapKey != "horse" {
		t.Error("Wrong key was provided to the callback")
	}

	if (mapVal != Animal{}) {
		t.Errorf("Wrong value was provided to the value")
	}

	if wasFound {
		t.Errorf("Key was found")
	}

	if m.Has("horse") {
		t.Errorf("Key was created")
	}
}

func TestPop(t *testing.T) {
	m := NewCMap[Animal]()

	monkey := Animal{"monkey"}
	m.Set("monkey", monkey)

	v, exists := m.Pop("monkey")

	if !exists || v != monkey {
		t.Error("Pop didn't find a monkey.")
	}

	v2, exists2 := m.Pop("monkey")

	if exists2 || v2 == monkey {
		t.Error("Pop keeps finding monkey")
	}

	if m.Count() != 0 {
		t.Error("Expecting count to be zero once item was Pop'ed.")
	}

	temp, ok := m.Get("monkey")

	if ok != false {
		t.Error("Expecting ok to be false for missing items.")
	}

	if (temp != Animal{}) {
		t.Error("Expecting item to be nil after its removal.")
	}
}

func TestCount(t *testing.T) {
	m := NewCMap[Animal]()
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	if m.Count() != 100 {
		t.Error("Expecting 100 element within map.")
	}
}

func TestItems(t *testing.T) {
	m := NewCMap[Animal]()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	items := m.Items()

	if len(items) != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

// 定义一个实现了 fmt.Stringer 接口的类型
type key struct {
	ID   int
	Name string
}

// 实现 fmt.Stringer 接口
func (k key) String() string {
	return fmt.Sprintf("%d-%s", k.ID, k.Name)
}

func TestStringer(t *testing.T) {
	// 使用 NewStringer 创建一个 ConcurrentMap 实例，键的类型是 MyKey，值的类型是 int
	cm := NewStringer[key, int]()

	// 添加键值对
	key1 := key{ID: 1, Name: "Alice"}
	cm.Set(key1, 100)

	key2 := key{ID: 2, Name: "Bob"}
	cm.Set(key2, 200)

	// 获取键的值并检查是否正确
	value1, ok1 := cm.Get(key1)
	if !ok1 || value1 != 100 {
		t.Errorf("Expected value 100 for key %v, got %v", key1, value1)
	}

	value2, ok2 := cm.Get(key2)
	if !ok2 || value2 != 200 {
		t.Errorf("Expected value 200 for key %v, got %v", key2, value2)
	}

	// 检查 ConcurrentMap 的元素总数
	expectedCount := 2
	actualCount := cm.Count()
	if actualCount != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, actualCount)
	}

	// 移除一个键值对
	cm.Remove(key1)

	// 检查元素是否已被移除
	_, ok1 = cm.Get(key1)
	if ok1 {
		t.Errorf("Key %v should have been removed", key1)
	}

	// 检查 ConcurrentMap 的元素总数
	expectedCount = 1
	actualCount = cm.Count()
	if actualCount != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, actualCount)
	}
}

func TestConcurrent(t *testing.T) {
	m := NewCMap[int]()
	ch := make(chan int)
	const iterations = 1000
	var a [iterations]int

	// Using go routines insert 1000 ints into our map.
	go func() {
		for i := 0; i < iterations/2; i++ {
			// Add item to map.
			m.Set(strconv.Itoa(i), i)

			// Retrieve item from map.
			val, _ := m.Get(strconv.Itoa(i))

			// Write to channel inserted value.
			ch <- val
		} // Call go routine with current index.
	}()

	go func() {
		for i := iterations / 2; i < iterations; i++ {
			// Add item to map.
			m.Set(strconv.Itoa(i), i)

			// Retrieve item from map.
			val, _ := m.Get(strconv.Itoa(i))

			// Write to channel inserted value.
			ch <- val
		} // Call go routine with current index.
	}()

	// Wait for all go routines to finish.
	counter := 0
	for elem := range ch {
		a[counter] = elem
		counter++
		if counter == iterations {
			break
		}
	}

	// Sorts array, will make is simpler to verify all inserted values we're returned.
	sort.Ints(a[0:iterations])

	// Make sure map contains 1000 elements.
	if m.Count() != iterations {
		t.Error("Expecting 1000 elements.")
	}

	// Make sure all inserted values we're fetched from map.
	for i := 0; i < iterations; i++ {
		if i != a[i] {
			t.Error("missing value", i)
		}
	}
}

func TestJsonMarshal(t *testing.T) {

	expected := "{\"a\":1,\"b\":2}"
	m := NewWithCustomShardingFunction[string, int](func(key string) uint32 {
		// 根据自己的需求定义分片函数的逻辑
		// 将字符串键转换为整数键
		// 然后对分片数量（例如，32）取模，确定映射到的分片索引
		var sum uint32
		for i := 0; i < len(key); i++ {
			sum += uint32(key[i])
		}
		return sum % 32
	})
	m.Set("a", 1)
	m.Set("b", 2)
	j, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}

	if string(j) != expected {
		t.Error("json", string(j), "differ from expected", expected)
		return
	}
}

func TestMInsert(t *testing.T) {
	animals := map[string]Animal{
		"elephant": {"elephant"},
		"monkey":   {"monkey"},
	}
	m := NewCMap[Animal]()
	m.MSet(animals)

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}
}

func TestFnv32(t *testing.T) {
	key := []byte("ABC")

	hasher := fnv.New32()
	_, err := hasher.Write(key)
	if err != nil {
		t.Errorf(err.Error())
	}
	if fnv32(string(key)) != hasher.Sum32() {
		t.Errorf("Bundled fnv32 produced %d, expected result from hash/fnv32 is %d", fnv32(string(key)), hasher.Sum32())
	}

}

func TestUpsert(t *testing.T) {
	dolphin := Animal{"dolphin"}
	whale := Animal{"whale"}
	tiger := Animal{"tiger"}
	lion := Animal{"lion"}

	cb := func(exists bool, valueInMap Animal, newValue Animal) Animal {
		if !exists {
			return newValue
		}
		valueInMap.name += newValue.name
		return valueInMap
	}

	m := NewCMap[Animal]()
	m.Set("marine", dolphin)
	m.Upsert("marine", whale, cb)
	m.Upsert("predator", tiger, cb)
	m.Upsert("predator", lion, cb)

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}

	marineAnimals, ok := m.Get("marine")
	if marineAnimals.name != "dolphinwhale" || !ok {
		t.Error("Set, then Upsert failed")
	}

	predators, ok := m.Get("predator")
	if !ok || predators.name != "tigerlion" {
		t.Error("Upsert, then Upsert failed")
	}
}
