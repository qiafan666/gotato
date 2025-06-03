package gcommon

import (
	"github.com/qiafan666/gotato/commons/gcast"
	"strings"
	"time"
)

// AlignIntervalTime 对齐时间，根据时间戳和是否是utc，返回对齐后的时间
// ts 时间戳，秒
// interval 格式：3min, 5h 只支持1day, 1week, 1month
func AlignIntervalTime(ts int64, interval string, isUTC bool) time.Time {
	var t time.Time
	if isUTC {
		t = time.Unix(ts, 0).UTC()
	} else {
		t = time.Unix(ts, 0).In(time.FixedZone("CST", 8*3600)) // 默认北京时间
	}

	switch {
	case strings.HasSuffix(interval, "min"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "min"))
		step := time.Duration(num) * time.Minute
		return t.Truncate(step)

	case strings.HasSuffix(interval, "h"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "h"))
		step := time.Duration(num) * time.Hour
		return t.Truncate(step)

	case strings.HasSuffix(interval, "day"):
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case strings.HasSuffix(interval, "week"):
		offset := int(t.Weekday())
		if offset == 0 {
			offset = 7
		}
		aligned := t.AddDate(0, 0, -offset+1)
		return time.Date(aligned.Year(), aligned.Month(), aligned.Day(), 0, 0, 0, 0, t.Location())
	case strings.HasSuffix(interval, "month"):
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}

// NextIntervalTime 获取下一个时间点，根据时间戳和时间间隔，返回下一个时间点
// interval 格式：1min, 1h, 1day, 1week, 1month, 1year
func NextIntervalTime(t time.Time, interval string) time.Time {
	switch {
	case strings.HasSuffix(interval, "min"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "min"))
		return t.Add(time.Duration(num) * time.Minute)

	case strings.HasSuffix(interval, "h"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "h"))
		return t.Add(time.Duration(num) * time.Hour)

	case strings.HasSuffix(interval, "day"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "day"))
		return t.AddDate(0, 0, num)

	case strings.HasSuffix(interval, "week"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "week"))
		return t.AddDate(0, 0, num*7)

	case strings.HasSuffix(interval, "month"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "month"))
		return t.AddDate(0, num, 0)

	case strings.HasSuffix(interval, "year"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "year"))
		return t.AddDate(num, 0, 0)

	default:
		return t // 不支持的周期，原样返回
	}
}
