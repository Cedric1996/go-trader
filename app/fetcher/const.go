/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 22:27:16
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-03-14 22:32:37
 */
package fetcher

type TimeScope string

const (
	OneMinute     TimeScope = "1m"
	FiveMinute    TimeScope = "5m"
	FifteenMinute TimeScope = "15m"
	ThirtyMinute  TimeScope = "30m"
	OneHour       TimeScope = "60m"
	TwoHour       TimeScope = "120m"
	Day           TimeScope = "1d"
	Week          TimeScope = "1w"
	Month         TimeScope = "1M"
)

const (
	STOCK SecurityType = "stock"
)
