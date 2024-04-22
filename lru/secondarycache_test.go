package lru

import (
	"strconv"
	"testing"
	"time"
)

func Test_SecondaryCache_GetsANonExistantValue(t *testing.T) {
	cache := newLayered[string]().GetOrCreateSecondaryCache("foo")
	Equal(t, cache == nil, false)
}

func Test_SecondaryCache_SetANewValue(t *testing.T) {
	cache := newLayered[string]()
	cache.Set("spice", "flow", "a value", time.Minute)
	sCache := cache.GetOrCreateSecondaryCache("spice")
	Equal(t, sCache.Get("flow").Value(), "a value")
	Equal(t, sCache.Get("stop"), nil)
}

func Test_SecondaryCache_ValueCanBeSeenInBothCaches1(t *testing.T) {
	cache := newLayered[string]()
	cache.Set("spice", "flow", "a value", time.Minute)
	sCache := cache.GetOrCreateSecondaryCache("spice")
	sCache.Set("orinoco", "another value", time.Minute)
	Equal(t, sCache.Get("orinoco").Value(), "another value")
	Equal(t, cache.Get("spice", "orinoco").Value(), "another value")
}

func Test_SecondaryCache_ValueCanBeSeenInBothCaches2(t *testing.T) {
	cache := newLayered[string]()
	sCache := cache.GetOrCreateSecondaryCache("spice")
	sCache.Set("flow", "a value", time.Minute)
	Equal(t, sCache.Get("flow").Value(), "a value")
	Equal(t, cache.Get("spice", "flow").Value(), "a value")
}

func Test_SecondaryCache_DeletesAreReflectedInBothCaches(t *testing.T) {
	cache := newLayered[string]()
	cache.Set("spice", "flow", "a value", time.Minute)
	cache.Set("spice", "sister", "ghanima", time.Minute)
	sCache := cache.GetOrCreateSecondaryCache("spice")

	cache.Delete("spice", "flow")
	Equal(t, cache.Get("spice", "flow"), nil)
	Equal(t, sCache.Get("flow"), nil)

	sCache.Delete("sister")
	Equal(t, cache.Get("spice", "sister"), nil)
	Equal(t, sCache.Get("sister"), nil)
}

func Test_SecondaryCache_ReplaceDoesNothingIfKeyDoesNotExist(t *testing.T) {
	cache := newLayered[string]()
	sCache := cache.GetOrCreateSecondaryCache("spice")
	Equal(t, sCache.Replace("flow", "value-a"), false)
	Equal(t, cache.Get("spice", "flow"), nil)
}

func Test_SecondaryCache_ReplaceUpdatesTheValue(t *testing.T) {
	cache := newLayered[string]()
	cache.Set("spice", "flow", "value-a", time.Minute)
	sCache := cache.GetOrCreateSecondaryCache("spice")
	Equal(t, sCache.Replace("flow", "value-b"), true)
	Equal(t, cache.Get("spice", "flow").Value(), "value-b")
}

func Test_SecondaryCache_FetchReturnsAnExistingValue(t *testing.T) {
	cache := newLayered[string]()
	cache.Set("spice", "flow", "value-a", time.Minute)
	sCache := cache.GetOrCreateSecondaryCache("spice")
	val, _ := sCache.Fetch("flow", time.Minute, func() (string, error) { return "a fetched value", nil })
	Equal(t, val.Value(), "value-a")
}

func Test_SecondaryCache_FetchReturnsANewValue(t *testing.T) {
	cache := newLayered[string]()
	sCache := cache.GetOrCreateSecondaryCache("spice")
	val, _ := sCache.Fetch("flow", time.Minute, func() (string, error) { return "a fetched value", nil })
	Equal(t, val.Value(), "a fetched value")
}

func Test_SecondaryCache_TrackerDoesNotCleanupHeldInstance(t *testing.T) {
	cache := Layered(Configure[int]().ItemsToPrune(10).Track())
	for i := 0; i < 10; i++ {
		cache.Set(strconv.Itoa(i), "a", i, time.Minute)
	}
	sCache := cache.GetOrCreateSecondaryCache("0")
	item := sCache.TrackingGet("a")
	time.Sleep(time.Millisecond * 10)
	cache.GC()
	Equal(t, cache.Get("0", "a").Value(), 0)
	Equal(t, cache.Get("1", "a"), nil)
	item.Release()
	cache.GC()
	Equal(t, cache.Get("0", "a"), nil)
}
