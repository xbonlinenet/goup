package util

import (
	"fmt"
	"time"
)

//time format
const (
	TIME_YYYY_MM_DD_HH_MM_SS_M6TZ = "2006-01-02T15:04:05.000000Z"
	TIME_YYYY_MM_DD_HH_MM_SS_M9   = "2006-01-02 15:04:05.000000000"
	TIME_YYYY_MM_DD_HH_MM_SS_M6   = "2006-01-02 15:04:05.000000"
	TIME_YYYY_MM_DD_HH_MM_SS_M3   = "2006-01-02 15:04:05.000"
	TIME_YYYY_MM_DD_HH_MM_SS      = "2006-01-02 15:04:05"
	TIME_YYYY_MM_DD               = "2006-01-02"
	TIME_YYYYMMDDHHMMSSM9         = "20060102150405000000000"
	TIME_YYYYMMDDHHMMSSM6         = "20060102150405000000"
	TIME_YYYYMMDDHHMMSSM3         = "20060102150405000"
	TIME_YYYYMMDDHHMMSS           = "20060102150405"
	TIME_YYYYMMDDHHMM             = "200601021504"
	TIME_YYYYMMDD                 = "20060102"
)

var (
	// VNLoc 越南时区
	VNLoc, _ = time.LoadLocation("Asia/Vientiane")
)

// GetTodayString 获取当天的字符串
func GetTodayString() string {
	str := time.Now().In(VNLoc).Format("20060102")
	return str
}

// ToVnDateString 转化成越南的日期
func ToVnDateString(t time.Time) string {
	str := t.In(VNLoc).Format("20060102")
	return str
}

// GetDayOfMonth 获取当前在当月是第几天的字符串
func GetDayOfMonth() string {
	str := fmt.Sprintf("%d", time.Now().Day())
	return str
}

// GetHourOfDay 获取当前是几点
func GetHourOfDay() int {
	return time.Now().In(VNLoc).Hour()
}

// GetSecondOfDay 获取当前是当天的第几秒
func GetSecondOfDay() int64 {
	zero := GetTodayZeroUTC()
	return time.Now().Unix() - zero
}

// GetLast24hUTC 获取24小时前的 UTC 时间戳
func GetLast24hUTC() int64 {
	return time.Now().Add(-24 * time.Hour).Unix()
}

// GetTodayZeroUTC 获取当天最早的时间戳
func GetTodayZeroUTC() int64 {
	str := time.Now().In(VNLoc).Format("20060102")
	t, _ := time.ParseInLocation("20060102", str, VNLoc)
	return t.Unix()
}

// GetTodayZero 获取当天起点时间
func GetTodayZero() time.Time {
	str := time.Now().In(VNLoc).Format("20060102")
	t, _ := time.ParseInLocation("20060102", str, VNLoc)
	return t
}

// GetVNDateFromUTC 根据 UTC 时间戳获取当天越南日期的字符串
func GetVNDateFromUTC(unixSeconds int64) string {
	return time.Unix(unixSeconds, 0).In(VNLoc).Format("20060102")
}

// GetLastNDayStr 获取最近 N 天，以字符串表示，如: 20060102, 20060101
func GetLastNDayStr(n int, today bool) []string {

	strList := make([]string, 0, n)
	if today {
		str := time.Now().In(VNLoc).Format("20060102")
		strList = append(strList, str)
	}

	before := 1
loop:
	for {
		if len(strList) >= n {
			break loop
		}
		str := time.Now().Add(-1 * time.Duration(before*24) * time.Hour).In(VNLoc).Format("20060102")
		strList = append(strList, str)
		before++
	}

	return strList
}

// GetDayBeginBySpecialTime 获取给定时间当天最早的时间点 VNLoc
func GetDayBeginBySpecialTime(t time.Time) time.Time {
	str := t.In(VNLoc).Format("20060102")
	begin, _ := time.ParseInLocation("20060102", str, VNLoc)
	return begin
}

// GetDayEndBySpecialTime 获取给定时间当天最早的时间点 VNLoc
func GetDayEndBySpecialTime(t time.Time) time.Time {
	str := t.In(VNLoc).Format("20060102")
	begin, _ := time.ParseInLocation("20060102", str, VNLoc)
	return begin.Add((24*3600 - 1) * time.Second)
}

// FormatForDB 转换成保存到数据库的时间字符串， UTC
func FormatForDB(t time.Time) string {
	return t.In(time.UTC).Format("2006-01-02 15:04:05")
}
