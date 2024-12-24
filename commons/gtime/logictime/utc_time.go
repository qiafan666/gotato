package logictime

import (
	"math"
	"sync/atomic"
	"time"
)

// 其他
const (
	Year        = 365 * Day
	Month       = 30 * Day
	Week        = 7 * Day
	Day         = 24 * Hour
	Hour        = 60 * Minute
	Minute      = 60 * Second
	Second      = 1000 * Millisecond
	Millisecond = 1

	YearsBegin    = 2000         // 从2000年开始计算周数、月数
	YearPerMonth  = 12           // 每年12个月
	WeekPerMonth  = 52           // 每年52周
	BeginMsOf2000 = 946684800000 // 2000.1.1 0点的时间戳 毫秒
)

var offset time.Duration

func SetZone(name string) error {
	location := "utc"
	if name == "Asia/Shanghai" {
		location = name
	}
	loc, err := time.LoadLocation(location)
	if err != nil {
		return err
	}
	time.Local = loc
	return nil
}

// Now UTC时间
func Now() time.Time {
	realNow := time.Now()
	offsetVal := GetTimeOffset()
	if offsetVal == 0 {
		return realNow.UTC()
	}
	return realNow.Add(offsetVal).UTC()
}

func GetTimeOffset() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(&offset)))
}

func SetTimeOffset(val time.Duration) {
	atomic.StoreInt64((*int64)(&offset), int64(val))
}

func AddTimeOffset(val time.Duration) {
	atomic.AddInt64((*int64)(&offset), int64(val))
}

func Since(t time.Time) time.Duration {
	return Now().Sub(t)
}

// NowMs 毫秒级时间戳
func NowMs() int64 {
	return Now().UnixNano() / 1e6
}

// NowFrame 十分之一秒
func NowFrame() int64 {
	return Now().UnixNano() / 1e8
}

// NowUs 微秒级时间戳
func NowUs() int64 {
	return Now().UnixNano() / 1e3
}

// NowNs 纳秒级时间戳
func NowNs() int64 {
	return Now().UnixNano()
}

// TodayMs 计算本地今天0点ms
func TodayMs(ms int64) int64 {
	t := time.Unix(ms/Second, 0)
	todayStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return todayStart.UnixNano() / 1e6
}

// NextDayMs 计算下一天0点ms
func NextDayMs(nowMs int64) int64 {
	ts := TodayMs(nowMs)
	return ts + Day
}

// WeekMs 周日0点
func WeekMs(nowMs int64) int64 {
	t := Ms2Time(nowMs)
	d := time.Date(t.Year(), t.Month(), t.Day()-int(t.Weekday()), 0, 0, 0, 0, t.Location())
	// 周一算一周的开始
	return Time2Ms(d) + Day
}

// NextHourMs 下个整点时间戳
func NextHourMs(nowMs int64) int64 {
	if nowMs%Hour == 0 {
		return nowMs + Hour
	}

	ts := nowMs/Hour*Hour + Hour
	return ts
}

// NextWeekMs 下周周一，0点
func NextWeekMs(nowMs int64) int64 {
	ws := WeekMs(nowMs)
	return ws + Day*7
}

// NextMonthMs 下个月月初，0点
func NextMonthMs(nowMs int64) int64 {
	ts := Ms2Time(nowMs)
	t := time.Date(ts.Year(), ts.Month()+1, 1, 0, 0, 0, 0, ts.Location())
	if ts.Month() == 12 {
		t = time.Date(ts.Year()+1, 1, 1, 0, 0, 0, 0, ts.Location())
	}
	return Time2Ms(t)
}

// NowSec 秒级时间戳
func NowSec() int64 {
	return Now().UnixNano() / 1e9
}

// SameDay 是否是同一天，参数毫秒
func SameDay(ms1 int64, ms2 int64) bool {
	t1 := Ms2Time(ms1)
	t2 := Ms2Time(ms2)
	return t1.Day() == t2.Day() && t1.Month() == t2.Month() && t1.Year() == t2.Year()
}

// Ms2Day 将得到的时间戳转化到天
func Ms2Day(ms int64) int64 {
	return ms / Day
}

// Ms2Sec 毫秒 -> 秒
func Ms2Sec(ms int64) int64 {
	return ms / Second
}

// Ms2Ns 毫秒 -> 纳秒
func Ms2Ns(ms int64) int64 {
	return ms * 1000 * 1000
}

// Ms2Hour 毫秒 -> 小时
func Ms2Hour(ms int64) int64 {
	return ms / Hour
}

// Sec2Ms 秒 -> 毫秒
func Sec2Ms(sec int32) int32 {
	return sec * Second
}

// Min2Ms 分钟 -> 毫秒
func Min2Ms(min int32) int32 {
	return min * Minute
}

// Hour2Ms 小时 -> 毫秒
func Hour2Ms(h int32) int32 {
	return h * Hour
}

// Time2Ms 系统时间转化为 ms时间戳
func Time2Ms(t time.Time) int64 {
	return t.UTC().UnixNano() / 1e6
}

// Ms2Time ms时间戳 转化为 本地时间
func Ms2Time(ms int64) time.Time {
	return time.Unix(ms/1000, ms%1000*1e6).Local()
}

// TrackTime 一行代码获取函数的执行时间
func TrackTime(pre time.Time) time.Duration {
	elapsed := time.Since(pre)
	return elapsed
}

// IsDayBegin time1是否是一天0点, s
func IsDayBegin(time1 int64) bool {
	ts := Ms2Time(time1 * 1000)
	return ts.Hour() == 0 && ts.Minute() == 0 && ts.Second() == 0
}

// IsHourBegin 当前时间是否整点
func IsHourBegin(timeSec int64) bool {
	t := time.Unix(timeSec/Second, 0)
	return t.Minute() == 0 && t.Second() == 0
}

// TimeHour 获取当前小时数
func TimeHour(timeMillSec int64) int64 {
	t := time.Unix(timeMillSec/Second, 0)
	return int64(t.Local().Hour())
}

// TimeHourIndex 从2000.1.1累计小时数 不是精确的
func TimeHourIndex(timeMillSec int64) int64 {
	accDay := (timeMillSec - BeginMsOf2000) / Day
	t := time.Unix(timeMillSec/Second, 0)
	return int64(t.Local().Hour()) + accDay*24
}

// NowMonthIndex 从2000.1.1累计月数 不是精确的
func NowMonthIndex() int64 {
	year, _ := Now().Local().ISOWeek()
	month := Now().Local().Month()
	return (int64(year)-YearsBegin)*YearPerMonth + int64(month)
}

// NowWeekIndex 从2000.1.1累计周数 不是精确的
func NowWeekIndex() int64 {
	year, week := Now().Local().ISOWeek()
	return int64((year-YearsBegin)*WeekPerMonth + week)
}

// NowDayIndex 从2000.1.1累计天数
func NowDayIndex() int64 {
	beginTS := int64(BeginMsOf2000)
	if time.Local.String() == "Asia/Shanghai" {
		beginTS -= 8 * Hour
	}
	accDay := (NowMs() - beginTS) / Day
	return accDay
}

// TimeDayIndex 从2000.1.1累计天数
func TimeDayIndex(timeMSec int64) int64 {
	beginTS := int64(BeginMsOf2000)
	if time.Local.String() == "Asia/Shanghai" {
		beginTS -= 8 * Hour
	}
	accDay := (timeMSec - beginTS) / Day
	return accDay
}

// DaysApart 计算两个秒时间戳所在的日期相隔了几天, 返回的是绝对值
func DaysApart(sec1, sec2 int64) int64 {
	t1 := time.Unix(sec1, 0)
	t2 := time.Unix(sec2, 0)
	date1 := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	date2 := time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())
	days := int64(math.Abs(date2.Sub(date1).Hours() / 24))
	return days
}

// CrossDays 两个时间戳跨多少天, ms
func CrossDays(t1, t2 int64) int64 {
	tt1 := TodayMs(t1)
	tt2 := TodayMs(t2)
	day := (tt1 - tt2) / Day
	if day > 0 {
		return day
	}
	return -day
}

func NowConvertTimeToEsFormat() string {
	return Now().In(time.Local).Format(time.RFC3339Nano)
}

// IsTimeOverlap 时间重叠
// startTime2, entTime2是否与startTime1, entTime1有重叠
// 返回true有重叠 false没有重叠
func IsTimeOverlap(startTime1, entTime1 int64, startTime2, entTime2 int64) bool {

	t1 := time.Unix(startTime1, 0)
	t2 := time.Unix(entTime1, 0)

	t3 := time.Unix(startTime2, 0)
	t4 := time.Unix(entTime2, 0)

	if t3.After(t2) || t4.Before(t1) {
		return false
	}
	return true
}
