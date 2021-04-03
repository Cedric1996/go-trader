/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 22:27:16
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-03 17:05:24
 */
package fetcher

type TimeScope string
type FinTable string
type SecurityType string

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

/*
*  API Reference:
* https://www.joinquant.com/help/api/help#Stock:%E8%B4%A2%E5%8A%A1%E6%95%B0%E6%8D%AE%E5%88%97%E8%A1%A8
 */
const (
	Balance  FinTable = "balance"
	Income   FinTable = "income"
	CashFlow FinTable = "cash_flow"
)
