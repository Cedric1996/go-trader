/*
 * @Author: cedric.jia
 * @Date: 2021-07-26 20:33:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-17 16:41:19
 */

package util

import (
	"strconv"
	"strings"
	"time"
)

func Today() string {
	t := strings.Split(time.Now().Format(time.RFC3339), "T")[0]
	return t
}

func TodayUnix() int64 {
	return ParseDate(Today()).Unix()
}

func DefaultBeginDate() string {
	return "2018-01-01"
}

func ToTimeStamp(t string) int64 {
	return ParseDate(t).Unix()
}

func ParseDate(t string) time.Time {
	parts := strings.Split(t, "-")
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	return time.Date(year, time.Month(month), day, 15, 0, 0, 0, time.UTC)
}

func ToDate(timestamp int64) string {
	tm := time.Unix(timestamp, 0)
	return strings.Split(tm.Format(time.RFC3339), "T")[0]
}
