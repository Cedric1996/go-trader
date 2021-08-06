/*
 * @Author: cedric.jia
 * @Date: 2021-07-26 20:33:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-04 21:54:19
 */

package service

import (
	"strconv"
	"strings"
	"time"
)

func today() string {
	t := strings.Split(time.Now().Format(time.RFC3339), "T")[0]
	return t
}

func defaultBeginDate() string {
	return "2018-01-01"
}

func toTimeStamp(t string) int64 {
	parts := strings.Split(t, "-")
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Unix()
}
