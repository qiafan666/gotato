package timermgr

import (
	"github.com/qiafan666/gotato/commons/gcommon/sval"
)

type Timer struct {
	TimerID   int64 `bson:"_id"`
	TimerType int32
	StartTs   int64
	EndTs     int64
	TimerData sval.M
	IsTicker  bool
}

func (t *Timer) update(endTs int64) {
	t.EndTs = endTs
}

// TimerList 为了Timer能使用sort.Sort排序
type TimerList []*Timer

// Len len
func (s TimerList) Len() int {
	return len(s)
}

// Swap swap
func (s TimerList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less less
func (s TimerList) Less(i, j int) bool {
	return s[i].EndTs < s[j].EndTs
}
