package sval

import "testing"

func TestSValue(t *testing.T) {
	// Test Int32 encoding and decoding
	int32Val := int32(12345)
	sInt32 := Int32(int32Val)
	if sInt32.Int32() != int32Val {
		t.Errorf("Expected Int32: %d, got: %d", int32Val, sInt32.Int32())
	}

	// Test Int64 encoding and decoding
	int64Val := int64(1234567890123)
	sInt64 := Int64(int64Val)
	if sInt64.Int64() != int64Val {
		t.Errorf("Expected Int64: %d, got: %d", int64Val, sInt64.Int64())
	}

	// Test string encoding and decoding
	strVal := "hello world"
	sStr := Str(strVal)
	if sStr.Str() != strVal {
		t.Errorf("Expected string: %s, got: %s", strVal, sStr.Str())
	}

	// Test boolean encoding and decoding
	boolVal := true
	sBool := Bool(boolVal)
	if sBool.Bool() != boolVal {
		t.Errorf("Expected Bool: %v, got: %v", boolVal, sBool.Bool())
	}

	// Test object encoding and decoding
	type TestObj struct {
		A int    `json:"a"`
		B string `json:"b"`
	}
	obj := TestObj{A: 42, B: "test"}
	sObj := Obj(obj)
	var decodedObj TestObj
	sObj.Obj(&decodedObj)
	if obj != decodedObj {
		t.Errorf("Expected Obj: %+v, got: %+v", obj, decodedObj)
	}
}

func TestM(t *testing.T) {
	// Create and populate M
	m := M{
		"key1": Int32(123),
		"key2": Int64(4567890123),
		"key3": Str("value"),
		"key4": Bool(true),
	}

	// Test Int32 retrieval
	if m.Int32("key1") != 123 {
		t.Errorf("Expected key1: %d, got: %d", 123, m.Int32("key1"))
	}

	// Test Int64 retrieval
	if m.Int64("key2") != 4567890123 {
		t.Errorf("Expected key2: %d, got: %d", 4567890123, m.Int64("key2"))
	}

	// Test string retrieval
	if m.Str("key3") != "value" {
		t.Errorf("Expected key3: %s, got: %s", "value", m.Str("key3"))
	}

	// Test boolean retrieval
	if m.Bool("key4") != true {
		t.Errorf("Expected key4: %v, got: %v", true, m.Bool("key4"))
	}

	// Test Clone
	cloned := m.Clone()
	if len(cloned) != len(m) || cloned["key1"] != m["key1"] {
		t.Errorf("Clone mismatch. Expected: %+v, got: %+v", m, cloned)
	}

	// Test Obj retrieval
	type TestObj struct {
		A int    `json:"a"`
		B string `json:"b"`
	}
	obj := TestObj{A: 42, B: "test"}
	m["key5"] = Obj(obj)
	var decodedObj TestObj
	if !m.Obj("key5", &decodedObj) || obj != decodedObj {
		t.Errorf("Expected Obj: %+v, got: %+v", obj, decodedObj)
	}
}

func TestInvalidSValue(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid SValue, but it did not happen")
		}
	}()

	// Create an invalid SValue and trigger a panic
	invalid := SValue("invalid")
	invalid.Int32()
}
