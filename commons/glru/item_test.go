package glru

import (
	"math"
	"testing"
	"time"
)

func Test_Item_Key(t *testing.T) {
	item := &Item[int]{key: "foo"}
	Equal(t, item.Key(), "foo")
}

func Test_Item_Promotability(t *testing.T) {
	item := &Item[int]{promotions: 4}
	Equal(t, item.shouldPromote(5), true)
	Equal(t, item.shouldPromote(5), false)
}

func Test_Item_Expired(t *testing.T) {
	now := time.Now().UnixNano()
	item1 := &Item[int]{expires: now + (10 * int64(time.Millisecond))}
	item2 := &Item[int]{expires: now - (10 * int64(time.Millisecond))}
	Equal(t, item1.Expired(), false)
	Equal(t, item2.Expired(), true)
}

func Test_Item_TTL(t *testing.T) {
	now := time.Now().UnixNano()
	item1 := &Item[int]{expires: now + int64(time.Second)}
	item2 := &Item[int]{expires: now - int64(time.Second)}
	Equal(t, int(math.Ceil(item1.TTL().Seconds())), 1)
	Equal(t, int(math.Ceil(item2.TTL().Seconds())), -1)
}

func Test_Item_Expires(t *testing.T) {
	now := time.Now().UnixNano()
	item := &Item[int]{expires: now + (10)}
	Equal(t, item.Expires().UnixNano(), now+10)
}

func Test_Item_Extend(t *testing.T) {
	item := &Item[int]{expires: time.Now().UnixNano() + 10}
	item.Extend(time.Minute * 2)
	Equal(t, item.Expires().Unix(), time.Now().Unix()+120)
}
