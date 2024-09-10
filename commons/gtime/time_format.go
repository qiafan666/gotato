package gtime

import (
	"github.com/qiafan666/gotato/commons/gerr"
	"strconv"
	"time"
)

const (
	TimeOffset = 8 * 3600  //8 hour offset
	HalfOffset = 12 * 3600 //Half-day hourly offset
)

// GetCurrentTimestampBySecond 获取当前时间戳（秒）
func GetCurrentTimestampBySecond() int64 {
	return time.Now().Unix()
}

// UnixSecondToTime 转换unix时间戳（秒）为time.Time类型
func UnixSecondToTime(second int64) time.Time {
	return time.Unix(second, 0)
}

// UnixNanoSecondToTime 转换unix时间戳（纳秒）为time.Time类型
func UnixNanoSecondToTime(nanoSecond int64) time.Time {
	return time.Unix(0, nanoSecond)
}

// UnixMillSecondToTime 转换unix时间戳（毫秒）为time.Time类型
func UnixMillSecondToTime(millSecond int64) time.Time {
	return time.Unix(0, millSecond*1e6)
}

// GetCurrentTimestampByNano 获取当前时间戳（纳秒）
func GetCurrentTimestampByNano() int64 {
	return time.Now().UnixNano()
}

// GetCurrentTimestampByMill 获取当前时间戳（毫秒）
func GetCurrentTimestampByMill() int64 {
	return time.Now().UnixNano() / 1e6
}

// GetCurDayZeroTimestamp 获取当前0点时间戳
func GetCurDayZeroTimestamp() int64 {
	timeStr := time.Now().Format("2006-01-02")
	// parse 默认为utc时间，所以需要-8小时
	t, _ := time.Parse("2006-01-02", timeStr)
	return t.Unix() - TimeOffset
}

// GetCurDayHalfTimestamp 获取当日12点时间戳
func GetCurDayHalfTimestamp() int64 {
	return GetCurDayZeroTimestamp() + HalfOffset

}

// GetCurDayZeroTimeFormat 获取当日0点时间格式化字符串 "2006-01-02_15:04:05"
func GetCurDayZeroTimeFormat() string {
	return time.Unix(GetCurDayZeroTimestamp(), 0).Format("2006-01-02_15:04:05")
}

// GetCurDayHalfTimeFormat 获取当日12点时间格式化字符串 "2006-01-02_15-04-05"
func GetCurDayHalfTimeFormat() string {
	return time.Unix(GetCurDayZeroTimestamp()+HalfOffset, 0).Format("2006-01-02_15-04-05")
}

// GetTimeStampByFormat 转换时间格式字符串为时间戳字符串
func GetTimeStampByFormat(datetime string) string {
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	tmp, _ := time.ParseInLocation(timeLayout, datetime, loc)
	timestamp := tmp.Unix()
	return strconv.FormatInt(timestamp, 10)
}

// TimeStringFormatTimeUnix 转换时间格式字符串为时间戳
func TimeStringFormatTimeUnix(timeFormat string, timeSrc string) int64 {
	tm, _ := time.Parse(timeFormat, timeSrc)
	return tm.Unix()
}

// TimeStringToTime 转换时间格式字符串为time.Time类型
func TimeStringToTime(timeString string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", timeString)
	return t, gerr.WrapMsg(err, "error parsing time string")
}

// TimeToString 转换time.Time类型为字符串
func TimeToString(t time.Time) string {
	return t.Format("2006-01-02")
}

// GetCurrentTimeFormatted 获取当前时间格式化字符串 "2006-01-02 15:04:05"
func GetCurrentTimeFormatted() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetTimestampByTimezone 获取指定时区时间戳
func GetTimestampByTimezone(timezone string) (int64, error) {
	// set time zone
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return 0, gerr.WrapMsg(err, "error loading location")
	}
	// get current time
	currentTime := time.Now().In(location)
	// get timestamp
	timestamp := currentTime.Unix()
	return timestamp, nil
}

// DaysBetweenTimestamps 计算某个时区的时间戳到当前时间戳差值（天）
func DaysBetweenTimestamps(timezone string, timestamp int64) (int, error) {
	// set time zone
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return 0, gerr.WrapMsg(err, "error loading location")
	}
	// get current time
	now := time.Now().In(location)
	// timestamp to time
	givenTime := time.Unix(timestamp, 0)
	// calculate duration
	duration := now.Sub(givenTime)
	// change to days
	days := int(duration.Hours() / 24)
	return days, nil
}

// IsSameWeekday 判断某个时区的时间戳是否是同一星期内
func IsSameWeekday(timezone string, timestamp int64) (bool, error) {
	// set time zone
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return false, gerr.WrapMsg(err, "error loading location")
	}
	// get current weekday
	currentWeekday := time.Now().In(location).Weekday()
	// change timestamp to weekday
	givenTime := time.Unix(timestamp, 0)
	givenWeekday := givenTime.Weekday()
	// compare two days
	return currentWeekday == givenWeekday, nil
}

// IsSameDayOfMonth 判断某个时区的时间戳是否是同一月内
func IsSameDayOfMonth(timezone string, timestamp int64) (bool, error) {
	// set time zone
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return false, gerr.WrapMsg(err, "error loading location")
	}
	// Get the current day of the month
	currentDay := time.Now().In(location).Day()
	// Convert the timestamp to time and get the day of the month
	givenDay := time.Unix(timestamp, 0).Day()
	// Compare the days
	return currentDay == givenDay, nil
}

// IsWeekday 判断时间戳是否是工作日
func IsWeekday(timestamp int64) bool {
	// Convert the timestamp to time
	givenTime := time.Unix(timestamp, 0)
	// Get the day of the week
	weekday := givenTime.Weekday()
	// Check if the day is between Monday (1) and Friday (5)
	return weekday >= time.Monday && weekday <= time.Friday
}

// IsNthDayCycle 检查当前日期是否是N日周期的第n个周期
func IsNthDayCycle(timezone string, startTimestamp int64, n int) (bool, error) {
	// set time zone
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return false, gerr.WrapMsg(err, "error loading location")
	}
	// Parse the start date
	startTime := time.Unix(startTimestamp, 0)
	if err != nil {
		return false, gerr.WrapMsg(err, "invalid start timestamp format")
	}
	// Get the current time
	now := time.Now().In(location)
	// Calculate the difference in days between the current time and the start time
	diff := now.Sub(startTime).Hours() / 24
	// Check if the difference in days is a multiple of n
	return int(diff)%n == 0, nil
}

// IsNthWeekCycle 检查当前日期是否是N周周期的第n个周期
func IsNthWeekCycle(timezone string, startTimestamp int64, n int) (bool, error) {
	// set time zone
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return false, gerr.WrapMsg(err, "error loading location")
	}

	// Get the current time
	now := time.Now().In(location)

	// Parse the start timestamp
	startTime := time.Unix(startTimestamp, 0)
	if err != nil {
		return false, gerr.WrapMsg(err, "invalid start timestamp format")
	}

	// Calculate the difference in days between the current time and the start time
	diff := now.Sub(startTime).Hours() / 24

	// Convert days to weeks
	weeks := int(diff) / 7

	// Check if the difference in weeks is a multiple of n
	return weeks%n == 0, nil
}

// IsNthMonthCycle 检查当前日期是否是N月周期的第n个周期
func IsNthMonthCycle(timezone string, startTimestamp int64, n int) (bool, error) {
	// set time zone
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return false, gerr.WrapMsg(err, "error loading location")
	}

	// Get the current date
	now := time.Now().In(location)

	// Parse the start timestamp
	startTime := time.Unix(startTimestamp, 0)
	if err != nil {
		return false, gerr.WrapMsg(err, "invalid start timestamp format")
	}

	// Calculate the difference in months between the current time and the start time
	yearsDiff := now.Year() - startTime.Year()
	monthsDiff := int(now.Month()) - int(startTime.Month())
	totalMonths := yearsDiff*12 + monthsDiff

	// Check if the difference in months is a multiple of n
	return totalMonths%n == 0, nil
}
