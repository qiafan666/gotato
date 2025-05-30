package gcommon

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// AlignIntervalTime 对齐时间，根据时间戳和是否是utc，返回对齐后的时间
// ts 时间戳，秒
// interval 格式：1min, 1h, 1day, 1week, 1month, 1year
func AlignIntervalTime(ts int64, interval string, isUTC bool) (time.Time, error) {
	t := time.Unix(ts, 0)
	if isUTC {
		t = t.UTC()
	} else {
		t = t.In(time.FixedZone("CST", 8*3600)) // 默认北京时间
	}

	step, err := ParseInterval(interval)
	if err != nil {
		return time.Time{}, err
	}

	return t.Truncate(step), nil
}

// NextIntervalTime 获取下一个时间点，根据时间戳和时间间隔，返回下一个时间点
// interval 格式：1min, 1h, 1day, 1week, 1month, 1year
func NextIntervalTime(t time.Time, interval string) (time.Time, error) {
	step, err := ParseInterval(interval)
	if err != nil {
		return time.Time{}, err
	}

	return t.Add(step), nil
}

// ParseInterval 解析时间间隔字符串，返回对应的 time.Duration
func ParseInterval(interval string) (time.Duration, error) {
	switch {
	case strings.HasSuffix(interval, "min"):
		num, err := strconv.Atoi(strings.TrimSuffix(interval, "min"))
		if err != nil {
			return 0, fmt.Errorf("invalid interval: %s", interval)
		}
		return time.Duration(num) * time.Minute, nil

	case strings.HasSuffix(interval, "h"):
		num, err := strconv.Atoi(strings.TrimSuffix(interval, "h"))
		if err != nil {
			return 0, fmt.Errorf("invalid interval: %s", interval)
		}
		return time.Duration(num) * time.Hour, nil

	case strings.HasSuffix(interval, "day"):
		num, err := strconv.Atoi(strings.TrimSuffix(interval, "day"))
		if err != nil {
			return 0, fmt.Errorf("invalid interval: %s", interval)
		}
		days := 24 * time.Hour
		return time.Duration(num) * days, nil

	case strings.HasSuffix(interval, "week"):
		num, err := strconv.Atoi(strings.TrimSuffix(interval, "week"))
		if err != nil {
			return 0, fmt.Errorf("invalid interval: %s", interval)
		}
		weeks := 7 * 24 * time.Hour
		return time.Duration(num) * weeks, nil

	case strings.HasSuffix(interval, "month"):
		num, err := strconv.Atoi(strings.TrimSuffix(interval, "month"))
		if err != nil {
			return 0, fmt.Errorf("invalid interval: %s", interval)
		}
		return time.Duration(num) * 30 * 24 * time.Hour, nil // 简化处理，假设每月30天

	case strings.HasSuffix(interval, "year"):
		num, err := strconv.Atoi(strings.TrimSuffix(interval, "year"))
		if err != nil {
			return 0, fmt.Errorf("invalid interval: %s", interval)
		}
		return time.Duration(num) * 365 * 24 * time.Hour, nil // 简化处理，假设每年365天

	default:
		return 0, fmt.Errorf("unsupported interval: %s", interval)
	}
}
