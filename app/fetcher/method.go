/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 21:49:41
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 12:10:14
 */

package fetcher

import (
	"fmt"
	"strconv"
	"strings"

	"github.cedric1996.com/go-trader/app/context"
)

var (
	queryCount = 0
)

// GetQueryCount return remain count of query daily.
func GetQueryCount(ctx *context.Ctx) int64 {
	params := map[string]interface{}{
		"method": "get_query_count",
		"token":  Token(),
	}
	ctx.Params = params
	t, err := Request(ctx)
	if err != nil {
		fmt.Errorf("get query count error: %v", err)
		return 0
	}
	n, err := strconv.ParseInt(string(t), 10, 64)
	if err != nil {
		fmt.Errorf("Parse int64 error: %v", err)
		return 0
	}
	return n
}

func GetAllSecurities(ctx *context.Ctx, securityType SecurityType, t string) string {
	params := map[string]interface{}{
		"method": "get_all_securities",
		"token":  Token(),
		"code":   securityType,
		"date":   t,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get all securities error: %s", err).Error()
	}
	return string(res)
}

func GetSecurityInfo(ctx *context.Ctx, code string) string {
	params := map[string]interface{}{
		"method": "get_security_info",
		"token":  Token(),
		"code":   code,
	}
	ctx.Params = params
	t, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get security info error: %s", err).Error()
	}
	return string(t)
}

func GetPrice(ctx *context.Ctx, code string, t TimeScope, count int64) error {
	params := map[string]interface{}{
		"method": "get_price",
		"token":  Token(),
		"code":   code,
		"unit":   t,
		"count":  count,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get price error: %s", err)
	}
	if err := ParseResponse(ctx, res); err != nil {
		return err
	}
	return nil
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

func GetPriceWithPeriod(ctx *context.Ctx, code string, t TimeScope, begin string, end string) string {
	params := map[string]interface{}{
		"method":   "get_price_period",
		"token":    Token(),
		"code":     code,
		"unit":     t,
		"date":     begin,
		"end_date": end,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get price with period error: %s", err).Error()
	}
	return string(res)
}

func GetCurrentPrice(ctx *context.Ctx, code string) string {
	params := map[string]interface{}{
		"method": "get_current_price",
		"token":  Token(),
		"code":   code,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get current price error: %s", err).Error()
	}
	return string(res)
}

func GetCallAuction(ctx *context.Ctx, code string, begin string, end string) string {
	params := map[string]interface{}{
		"method":   "get_call_auction",
		"token":    Token(),
		"code":     code,
		"date":     begin,
		"end_date": end,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get call auction with period error: %s", err).Error()
	}
	return string(res)
}

func GetCurrentTick(ctx *context.Ctx, code string) string {
	params := map[string]interface{}{
		"method": "get_current_tick",
		"token":  Token(),
		"code":   code,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get current tick error: %s", err).Error()
	}
	return string(res)
}

func GetCurrentTicks(ctx *context.Ctx, codes string) string {
	params := map[string]interface{}{
		"method": "get_current_ticks",
		"token":  Token(),
		"code":   codes,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get current tick error: %s", err).Error()
	}
	return string(res)
}

func GetFundInfo(ctx *context.Ctx, code string, date string) string {
	params := map[string]interface{}{
		"method": "get_fund_info",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get fundI info error: %s", err).Error()
	}
	return string(res)
}

func GetIndexStocks(ctx *context.Ctx, code string, date string) []string {
	params := map[string]interface{}{
		"method": "get_index_stocks",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return []string{fmt.Errorf("get Index Stock error: %s", err).Error()}
	}
	stocks := strings.Split(string(res), "\n")
	return stocks
}

func GetIndexWeights(ctx *context.Ctx, code string, date string) string {
	params := map[string]interface{}{
		"method": "get_index_weights",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get Index Weights error: %s", err).Error()
	}
	return string(res)
}

func GetIndustry(ctx *context.Ctx, code string, date string) string {
	params := map[string]interface{}{
		"method": "get_industry",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get Industry error: %s", err).Error()
	}
	return string(res)
}

// Query 1000 data once by default.
func GetFundamentals(ctx *context.Ctx, table FinTable, code, date string) error {
	params := map[string]interface{}{
		"method": "get_fundamentals",
		"token":  Token(),
		"table":  table,
		"code":   code,
		"date":   date,
	}
	ctx.Params = params
	res, err := Request(ctx)
	if err != nil {
		return fmt.Errorf("get Industry error: %s", err)
	}
	if err := ParseResponse(ctx, res); err != nil {
		return err
	}
	return nil
}
