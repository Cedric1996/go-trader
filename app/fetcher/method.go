/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 21:49:41
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-03-15 22:31:07
 */

package fetcher

import (
	"fmt"
	"strconv"
)

type SecurityType string

var (
	queryCount = 0
)

// GetQueryCount return remain count of query daily.
func GetQueryCount() int64 {
	params := map[string]interface{}{
		"method": "get_query_count",
		"token":  Token(),
	}
	t, err := Request(params)
	if err != nil {
		fmt.Errorf("Get query count error: %v", err)
		return 0
	}
	n, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		fmt.Errorf("Parse int64 error: %v", err)
		return 0
	}
	return n
}

func GetAllSecurities(securityType SecurityType, t string) string {
	params := map[string]interface{}{
		"method": "get_all_securities",
		"token":  Token(),
		"code":   securityType,
		"date":   t,
	}
	t, err := Request(params)
	if err != nil {
		return fmt.Errorf("Get all securities error.", err).Error()
	}
	return t
}

func GetSecurityInfo(code string) string {
	params := map[string]interface{}{
		"method": "get_security_info",
		"token":  Token(),
		"code":   code,
	}
	t, err := Request(params)
	if err != nil {
		return fmt.Errorf("Get security info error.", err).Error()
	}
	return t
}

func GetPrice(code string, t TimeScope, count int64) string {
	params := map[string]interface{}{
		"method": "get_price",
		"token":  Token(),
		"code":   code,
		"unit":   t,
		"count":  count,
	}
	res, err := Request(params)
	if err != nil {
		return fmt.Errorf("Get price error.", err).Error()
	}
	return res
}

/* Sample Request
{
    "method": "get_price_period",
    "token": "5b6a9ba7b0f572bb6c287e280ed",
    "code": "600000.XSHG",
    "unit": "30m",
    "date": "2018-12-04 09:45:00",
    "end_date": "2018-12-04 10:40:00",
    "fq_ref_date": "2018-12-18"
}
*/
func GetPriceWithPeriod(code string, t TimeScope, begin string, end string) string {
	params := map[string]interface{}{
		"method":   "get_price_period",
		"token":    Token(),
		"code":     code,
		"unit":     t,
		"date":     begin,
		"end_date": end,
	}
	res, err := Request(params)
	if err != nil {
		return fmt.Errorf("Get price with period error.", err).Error()
	}
	return res
}

func GetCurrentPrice(code string) string {
	params := map[string]interface{}{
		"method": "get_current_price",
		"token":  Token(),
		"code":   code,
	}
	res, err := Request(params)
	if err != nil {
		return fmt.Errorf("Get current price error.", err).Error()
	}
	return res
}

func GetCallAuction(code string, begin string, end string) string {
	params := map[string]interface{}{
		"method":   "get_call_auction",
		"token":    Token(),
		"code":     code,
		"date":     begin,
		"end_date": end,
	}
	res, err := Request(params)
	if err != nil {
		return fmt.Errorf("Get call auction with period error.", err).Error()
	}
	return res
}

func GetCurrentTick(code string) string {
	params := map[string]interface{}{
		"method": "get_current_tick",
		"token":  Token(),
		"code":   code,
	}
	res, err := Request(params)
	if err != nil {
		return fmt.Errorf("Get current tick error.", err).Error()
	}
	return res
}

func GetCurrentTicks(codes string) string {
	params := map[string]interface{}{
		"method": "get_current_ticks",
		"token":  Token(),
		"code":   codes,
	}
	res, err := Request(params)
	if err != nil {
		return fmt.Errorf("Get current tick error.", err).Error()
	}
	return res
}

func GetFundInfo(code string, date string) string {
	params := map[string]interface{}{
		"method": "get_fund_info",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	res, err := Request(params)
	if err != nil {
		return fmt.Errorf("Get fundI info error.", err).Error()
	}
	return res
}
