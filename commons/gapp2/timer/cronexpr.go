package timer

// reference: https://github.com/robfig/cron
import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// CronExpr Cron表达式解析
// Field name   | Mandatory? | Allowed values | Allowed special characters
// ----------   | ---------- | -------------- | --------------------------
// Seconds      | No         | 0-59           | * / , -
// Minutes      | Yes        | 0-59           | * / , -
// Hours        | Yes        | 0-23           | * / , -
// Day of month | Yes        | 1-31           | * / , -
// Month        | Yes        | 1-12           | * / , -
// Day of week  | Yes        | 0-6            | * / , -
type CronExpr struct {
	sec   uint64
	min   uint64
	hour  uint64
	dom   uint64
	month uint64
	dow   uint64
}

// NewCronExpr goroutine safe
func NewCronExpr(expr string) (cronExpr *CronExpr, err error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 && len(fields) != 6 {
		err = fmt.Errorf("invalid expr %v: expected 5 or 6 fields, got %v", expr, len(fields))
		return nil, err
	}

	if len(fields) == 5 {
		fields = append([]string{"0"}, fields...)
	}

	cronExpr = new(CronExpr)
	// Seconds
	cronExpr.sec, err = parseCronField(fields[0], 0, 59)
	if err != nil {
		goto onError
	}
	// Minutes
	cronExpr.min, err = parseCronField(fields[1], 0, 59)
	if err != nil {
		goto onError
	}
	// Hours
	cronExpr.hour, err = parseCronField(fields[2], 0, 23)
	if err != nil {
		goto onError
	}
	// Day of month
	cronExpr.dom, err = parseCronField(fields[3], 1, 31)
	if err != nil {
		goto onError
	}
	// Month
	cronExpr.month, err = parseCronField(fields[4], 1, 12)
	if err != nil {
		goto onError
	}
	// Day of week
	cronExpr.dow, err = parseCronField(fields[5], 0, 6)
	if err != nil {
		goto onError
	}
	return cronExpr, nil

onError:
	return nil, fmt.Errorf("invalid expr %v: %v", expr, err)
}

// 1. *
// 2. num
// 3. num-num
// 4. */num
// 5. num/num (means num-max/num)
// 6. num-num/num
//
//nolint:gocognit
func parseCronField(field string, min, max int) (uint64, error) {
	fields := strings.Split(field, ",")
	cronField := uint64(0)
	for _, field := range fields {
		rangeAndIncr := strings.Split(field, "/")
		if len(rangeAndIncr) > 2 {
			return 0, fmt.Errorf("too many slashes: %v", field)
		}

		// range
		startAndEnd := strings.Split(rangeAndIncr[0], "-")
		if len(startAndEnd) > 2 {
			return 0, fmt.Errorf("too many hyphens: %v", rangeAndIncr[0])
		}

		var start, end int
		var err error
		if startAndEnd[0] == "*" {
			if len(startAndEnd) != 1 {
				return 0, fmt.Errorf("invalid range: %v", rangeAndIncr[0])
			}
			start = min
			end = max
		} else {
			// start
			start, err = strconv.Atoi(startAndEnd[0])
			if err != nil {
				return 0, fmt.Errorf("invalid range: %v", rangeAndIncr[0])
			}
			// end
			if len(startAndEnd) == 1 {
				if len(rangeAndIncr) == 2 {
					end = max
				} else {
					end = start
				}
			} else {
				end, err = strconv.Atoi(startAndEnd[1])
				if err != nil {
					return 0, fmt.Errorf("invalid range: %v", rangeAndIncr[0])
				}
			}
		}

		if start > end {
			return 0, fmt.Errorf("invalid range: %v", rangeAndIncr[0])
		}
		if start < min {
			return 0, fmt.Errorf("out of range [%v, %v]: %v", min, max, rangeAndIncr[0])
		}
		if end > max {
			return 0, fmt.Errorf("out of range [%v, %v]: %v", min, max, rangeAndIncr[0])
		}

		// increment
		var incr int
		if len(rangeAndIncr) == 1 {
			incr = 1
		} else {
			var err error
			incr, err = strconv.Atoi(rangeAndIncr[1])
			if err != nil {
				return 0, fmt.Errorf("invalid increment: %v", rangeAndIncr[1])
			}
			if incr <= 0 {
				return 0, fmt.Errorf("invalid increment: %v", rangeAndIncr[1])
			}
		}

		// cronField
		if incr == 1 {
			cronField |= ^(math.MaxUint64 << uint(end+1)) & (math.MaxUint64 << uint(start))
		} else {
			for i := start; i <= end; i += incr {
				cronField |= 1 << uint(i)
			}
		}
	}
	return cronField, nil
}

func (e *CronExpr) matchDay(t time.Time) bool {
	// day-of-month blank
	if e.dom == 0xfffffffe {
		return 1<<uint(t.Weekday())&e.dow != 0
	}

	// day-of-week blank
	if e.dow == 0x7f {
		return 1<<uint(t.Day())&e.dom != 0
	}

	return 1<<uint(t.Weekday())&e.dow != 0 ||
		1<<uint(t.Day())&e.dom != 0
}

// Next goroutine safe
//
//nolint:gocognit
func (e *CronExpr) Next(t0 time.Time) time.Time {
	// the upcoming second
	t := t0.Truncate(time.Second).Add(time.Second)

	year := t.Year()
	initFlag := false

retry:
	// Year
	if t.Year() > year+1 {
		return time.Time{}
	}

	// Month
	for 1<<uint(t.Month())&e.month == 0 {
		if !initFlag {
			initFlag = true
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		}

		t = t.AddDate(0, 1, 0)
		if t.Month() == time.January {
			goto retry
		}
	}

	// Day
	for !e.matchDay(t) {
		if !initFlag {
			initFlag = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}

		t = t.AddDate(0, 0, 1)
		if t.Day() == 1 {
			goto retry
		}
	}

	// Hours
	for 1<<uint(t.Hour())&e.hour == 0 {
		if !initFlag {
			initFlag = true
			t = t.Truncate(time.Hour)
		}

		t = t.Add(time.Hour)
		if t.Hour() == 0 {
			goto retry
		}
	}

	// Minutes
	for 1<<uint(t.Minute())&e.min == 0 {
		if !initFlag {
			initFlag = true
			t = t.Truncate(time.Minute)
		}

		t = t.Add(time.Minute)
		if t.Minute() == 0 {
			goto retry
		}
	}

	// Seconds
	for 1<<uint(t.Second())&e.sec == 0 {
		if !initFlag {
			initFlag = true
		}

		t = t.Add(time.Second)
		if t.Second() == 0 {
			goto retry
		}
	}

	return t
}
