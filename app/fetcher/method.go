/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 21:49:41
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-30 16:08:45
 */

package fetcher

import (
	"os"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/util"
)

var (
	queryCount = 0
)

func GetCurrentToken(c *ctx.Context) error {
	Mob := os.Getenv("TRADER_MOB")
	Pwd := os.Getenv("TRADER_PWD")
	// Mob := viper.GetString("env.mob")
	// Pwd := viper.GetString("env.pwd")
	params := map[string]interface{}{
		"method": "get_token",
		"mob":    Mob,
		"pwd":    Pwd,
	}
	c.Params = params
	return fetchData(c, "get current token")
}

// GetQueryCount return remain count of query daily.
// 获取当日剩余查询条数
func GetQueryCount(c *ctx.Context) error {
	params := map[string]interface{}{
		"method": "get_query_count",
		"token":  Token(),
	}
	c.Params = params
	return fetchData(c, "get query count")
}

// 获取所有标的信息
func GetAllSecurities(c *ctx.Context, date string) error {
	params := map[string]interface{}{
		"method": "get_all_securities",
		"token":  Token(),
		// By default, go-trader fetch stock infos
		"code": "stock",
		"date": date,
	}
	c.Params = params
	return fetchData(c, "get all securities")
}

// 获取单个标的信息
func GetSecurityInfo(c *ctx.Context, code string) error {
	params := map[string]interface{}{
		"method": "get_security_info",
		"token":  Token(),
		"code":   code,
	}
	c.Params = params
	return fetchData(c, "get security info")
}

// 获取指定时间周期的行情
func GetPrice(c *ctx.Context, code, date string, t TimeScope, count int) error {
	params := map[string]interface{}{
		"method":      "get_price",
		"token":       Token(),
		"code":        code,
		"unit":        t,
		"count":       count,
		"end_date":    date,
		"fq_ref_date": PostRefDate(),
	}
	c.Params = params
	return fetchData(c, "get price")
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

// 获取指定时间段的行情数据
func GetPriceWithPeriod(c *ctx.Context, code string, t TimeScope, begin string, end string) error {
	params := map[string]interface{}{
		"method":      "get_price_period",
		"token":       Token(),
		"code":        code,
		"unit":        t,
		"date":        begin,
		"end_date":    end,
		"fq_ref_date": PostRefDate(),
	}
	c.Params = params
	return fetchData(c, "get price with period")
}

// 获取标的当前价格
func GetCurrentPrice(c *ctx.Context, code string) error {
	params := map[string]interface{}{
		"method": "get_current_price",
		"token":  Token(),
		"code":   code,
	}
	c.Params = params
	return fetchData(c, "get current price")
}

// 获取集合竞价 tick 数据
func GetCallAuction(c *ctx.Context, code string, begin string, end string) error {
	params := map[string]interface{}{
		"method":   "get_call_auction",
		"token":    Token(),
		"code":     code,
		"date":     begin,
		"end_date": end,
	}
	c.Params = params
	return fetchData(c, "get call auction with period ")
}

// 获取最新 tick 数据
func GetCurrentTick(c *ctx.Context, code string) error {
	params := map[string]interface{}{
		"method": "get_current_tick",
		"token":  Token(),
		"code":   code,
	}
	c.Params = params
	return fetchData(c, "get current tick")
}

func GetCurrentTicks(c *ctx.Context, code string, limit int64) error {
	params := map[string]interface{}{
		"method":   "get_ticks",
		"token":    Token(),
		"code":     code,
		"count":    limit,
		"end_date": util.Today(),
	}
	c.Params = params
	return fetchData(c, "get current tick")
}

func GetFundInfo(c *ctx.Context, code string, date string) error {
	params := map[string]interface{}{
		"method": "get_fund_info",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	c.Params = params
	return fetchData(c, "get fund info")
}

func GetIndexStocks(c *ctx.Context, code string, date string) error {
	params := map[string]interface{}{
		"method": "get_index_stocks",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	c.Params = params
	return fetchData(c, "get Index Stock")
}

func GetIndexWeights(c *ctx.Context, code string, date string) error {
	params := map[string]interface{}{
		"method": "get_index_weights",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	c.Params = params
	return fetchData(c, "get Index Weights")
}

func GetIndustryList(c *ctx.Context, code string) error {
	params := map[string]interface{}{
		"method": "get_industries",
		"token":  Token(),
		"code":   code,
	}
	c.Params = params
	return fetchData(c, "get industry list")
}

func GetIndustry(c *ctx.Context, code string, date string) error {
	params := map[string]interface{}{
		"method": "get_industry",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	c.Params = params
	return fetchData(c, "get industry")
}

func GetIndustryStock(c *ctx.Context, code, date string) error {
	params := map[string]interface{}{
		"method": "get_industry_stocks",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	c.Params = params
	return fetchData(c, "get industry stocks")
}

// Query 1000 data once by default.
func GetFundamentals(c *ctx.Context, table FinTable, code, date string, count int) error {
	params := map[string]interface{}{
		"method": "get_fundamentals",
		"token":  Token(),
		"table":  table,
		"code":   code,
		"date":   date,
		"count":  count,
	}
	c.Params = params
	return fetchData(c, "get Fundamentals")
}

func GetConcepts(c *ctx.Context) error {
	params := map[string]interface{}{
		"method": "get_concepts",
		"token":  Token(),
	}
	c.Params = params
	return fetchData(c, "get concepts")
}

func GetConceptStock(c *ctx.Context, code, date string) error {
	params := map[string]interface{}{
		"method": "get_concept_stocks",
		"token":  Token(),
		"code":   code,
		"date":   date,
	}
	c.Params = params
	return fetchData(c, "get concept stocks")
}

func GetTradeDates(c *ctx.Context, beginDate, endDate string) error {
	params := map[string]interface{}{
		"method":   "get_trade_days",
		"token":    Token(),
		"date":     beginDate,
		"end_date": endDate,
	}
	c.Params = params
	return fetchData(c, "fetch trade dates")
}

// func RunQuery(c *ctx.Context) error {

// }
