package gcache

import (
	"testing"
	"time"
)

type TestStruct struct {
	Num      int
	Children []*TestStruct
}

func TestCache(t *testing.T) {
	tc := NewCache(DefaultExpiration, 0)

	a, found := tc.Get("a")
	if found || a != nil {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, found := tc.Get("b")
	if found || b != nil {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, found := tc.Get("c")
	if found || c != nil {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", "b", DefaultExpiration)
	tc.Set("c", 3.5, DefaultExpiration)

	x, found := tc.Get("a")
	if !found {
		t.Error("a was not found while getting a2")
	}
	if x == nil {
		t.Error("x for a is nil")
	} else if a2 := x.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}

	x, found = tc.Get("b")
	if !found {
		t.Error("b was not found while getting b2")
	}
	if x == nil {
		t.Error("x for b is nil")
	} else if b2 := x.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}

	x, found = tc.Get("c")
	if !found {
		t.Error("c was not found while getting c2")
	}
	if x == nil {
		t.Error("x for c is nil")
	} else if c2 := x.(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
}

func TestCacheTimes(t *testing.T) {
	var found bool

	tc := NewCache(50*time.Millisecond, 1*time.Millisecond)
	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", 2, DefaultExpiration)
	tc.Set("c", 3, 20*time.Millisecond)
	tc.Set("d", 4, 70*time.Millisecond)

	<-time.After(25 * time.Millisecond)
	_, found = tc.Get("c")
	if found {
		t.Error("Found c when it should have been automatically deleted")
	}

	<-time.After(30 * time.Millisecond)
	_, found = tc.Get("a")
	if found {
		t.Error("Found a when it should have been automatically deleted")
	}

	_, found = tc.Get("b")
	if !found {
		t.Error("Did not find b even though it was set to never expire")
	}

	_, found = tc.Get("d")
	if !found {
		t.Error("Did not find d even though it was set to expire later than the default")
	}

	<-time.After(20 * time.Millisecond)
	_, found = tc.Get("d")
	if found {
		t.Error("Found d when it should have been automatically deleted (later than the default)")
	}
}

func TestNewFrom(t *testing.T) {
	m := map[string]Item{
		"a": Item{
			Object:     1,
			Expiration: 0,
		},
		"b": Item{
			Object:     2,
			Expiration: 0,
		},
	}
	tc := NewFromCache(DefaultExpiration, 0, m)
	a, found := tc.Get("a")
	if !found {
		t.Fatal("Did not find a")
	}
	if a.(int) != 1 {
		t.Fatal("a is not 1")
	}
	b, found := tc.Get("b")
	if !found {
		t.Fatal("Did not find b")
	}
	if b.(int) != 2 {
		t.Fatal("b is not 2")
	}
}

func TestAdd(t *testing.T) {
	tc := NewCache(DefaultExpiration, 0)
	err := tc.Add("foo", "bar", DefaultExpiration)
	if err != nil {
		t.Error("Couldn't add foo even though it shouldn't exist")
	}
	err = tc.Add("foo", "baz", DefaultExpiration)
	if err == nil {
		t.Error("Successfully added another foo when it should have returned an error")
	}
}

func TestReplace(t *testing.T) {
	tc := NewCache(DefaultExpiration, 0)
	err := tc.Replace("foo", "bar", DefaultExpiration)
	if err == nil {
		t.Error("Replaced foo when it shouldn't exist")
	}
	tc.Set("foo", "bar", DefaultExpiration)
	err = tc.Replace("foo", "bar", DefaultExpiration)
	if err != nil {
		t.Error("Couldn't replace existing key foo")
	}
}

func TestDelete(t *testing.T) {
	tc := NewCache(DefaultExpiration, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	tc.Delete("foo")
	x, found := tc.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
}

func TestItemCount(t *testing.T) {
	tc := NewCache(DefaultExpiration, 0)
	tc.Set("foo", "1", DefaultExpiration)
	tc.Set("bar", "2", DefaultExpiration)
	tc.Set("baz", "3", DefaultExpiration)
	if n := tc.ItemCount(); n != 3 {
		t.Errorf("Item count is not 3: %d", n)
	}
}

func TestFlush(t *testing.T) {
	tc := NewCache(DefaultExpiration, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	tc.Set("baz", "yes", DefaultExpiration)
	tc.Flush()
	x, found := tc.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
	x, found = tc.Get("baz")
	if found {
		t.Error("baz was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
}

func TestIncrementWithInt64(t *testing.T) {
	tc := NewCache(DefaultExpiration, 0)
	tc.Set("tint64", int64(1), DefaultExpiration)
	err := tc.Increment("tint64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint64")
	if !found {
		t.Error("tint64 was not found")
	}
	if x.(int64) != 3 {
		t.Error("tint64 is not 3:", x)
	}
}
