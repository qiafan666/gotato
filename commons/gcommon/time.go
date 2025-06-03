package gcommon

import (
	"github.com/qiafan666/gotato/commons/gcast"
	"strings"
	"time"
)

// AlignIntervalTime 对齐时间，根据时间戳和是否是utc，返回对齐后的时间
// ts 时间戳，秒
// interval 格式：1min, 1h, 1day, 1week, 1month, 1year
func AlignIntervalTime(ts int64, interval string, isUTC bool) time.Time {
	var t time.Time
	if isUTC {
		t = time.Unix(ts, 0).UTC()
	} else {
		t = time.Unix(ts, 0).In(time.FixedZone("CST", 8*3600)) // 默认北京时间
	}

	switch {
	case strings.HasSuffix(interval, "min"):
		num := gcast.ToInt64(strings.TrimSuffix(interval, "min"))
		step := time.Duration(num) * time.Minute
		return t.Truncate(step)

	case strings.HasSuffix(interval, "h"):
		num := gcast.ToInt64(strings.TrimSuffix(interval, "h"))
		step := time.Duration(num) * time.Hour
		return t.Truncate(step)

	case strings.HasSuffix(interval, "day"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "day"))
		base := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		daysSinceEpoch := int(base.Unix() / 86400)
		alignedDays := daysSinceEpoch / num * num
		return time.Unix(int64(alignedDays)*86400, 0).In(t.Location())

	case strings.HasSuffix(interval, "week"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "week"))
		// 当前时间是第几周
		// 先对齐到最近一个周一
		offset := int(t.Weekday())
		if offset == 0 {
			offset = 7
		}
		monday := t.AddDate(0, 0, -offset+1)
		// 距离 Unix 起始时间的“第几个星期一”
		weeksSinceEpoch := int(monday.Unix() / 86400 / 7)
		alignedWeeks := weeksSinceEpoch / num * num
		return time.Unix(int64(alignedWeeks*7*86400), 0).In(t.Location())

	case strings.HasSuffix(interval, "month"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "month"))
		// 当前时间所在的月份编号
		monthIndex := t.Year()*12 + int(t.Month()) - 1
		alignedMonthIndex := monthIndex / num * num
		year := alignedMonthIndex / 12
		month := time.Month(alignedMonthIndex%12 + 1)
		return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())

	case strings.HasSuffix(interval, "year"):
		num := gcast.ToInt(strings.TrimSuffix(interval, "year"))
		year := t.Year()
		alignedYear := year / num * num
		return time.Date(alignedYear, 1, 1, 0, 0, 0, 0, t.Location())

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
