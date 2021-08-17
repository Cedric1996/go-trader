/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 16:36:57
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-17 16:41:32
 */
package service

import (
	"fmt"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/util"
)

type fetchStockDailyDatum struct {
	code     string
	tradeDay *models.TradeDay
}

// Count should not be greater than 5000.
func GetPricesByDay(code, date string, count int) ([]*models.Price, error) {
	c := &ctx.Context{}
	if err := fetcher.GetPrice(c, code, date, fetcher.Day, count); err != nil {
		fmt.Printf("ERROR: GetPricesByDay error: %s\n", err)
		return nil, err
	}
	prices := models.ParsePriceInfo(c)
	return prices, nil
}

func initStockPriceByDay(code string, count int) error {
	c := &ctx.Context{}
	if err := fetcher.GetPrice(c, code, "2018-07-09", fetcher.Day, count); err != nil {
		fmt.Printf("ERROR: GetPricesByDay error: %s\n", err)
		return err
	}
	// if err := models.InsertStockPriceDay(c); err != nil {
	// 	return err
	// }
	return nil
}

/**
 * Init stock price date from 2018-01-01 and update
 * Stock table
 */
func InitStockPrice() error {
	initStockQueue, err := queue.NewQueue("init", 50, 10, func(data interface{}) (interface{}, error) {
		code := data.(string)
		if err := initStockPriceByDay(code, 200); err != nil {
			return nil, err
		}
		return nil, nil
	}, func(datas []interface{}) error {
		return nil
	})
	if err != nil {
		return err
	}
	for i, _ := range SecuritySet {
		initStockQueue.Push(i)
	}
	initStockQueue.Close()
	return nil
}

/**
 * fetch all stockPriceDay with specified day
s*/
func fetchLatestTradeDay(c *ctx.Context) error {
	if err := fetchTradeDay(c, "", ""); err != nil {
		return err
	}
	tradeDayMap := make(map[string]interface{})
	tradeDayToInsert := make([]interface{}, 0)
	tradeDays, err := models.GetTradeDay(true, 0, 0)
	if err != nil {
		return err
	}
	for _, day := range tradeDays {
		tradeDayMap[day.Date] = day
	}
	tradeDayRes := c.Params["trade_day"].([]string)
	for _, day := range tradeDayRes {
		if _, ok := tradeDayMap[day]; !ok {
			tradeDayToInsert = append(tradeDayToInsert, &models.TradeDay{
				Date:      day,
				Timestamp: util.ToTimeStamp(day),
				IsInit:    false,
			})
		}
	}
	if err := models.InsertTradeDay(tradeDayToInsert); err != nil {
		return fmt.Errorf("update lastest trade day error: %s", err)
	}
	return nil
}

/**
 * fetch trade days between begin/end date
 */
func fetchTradeDay(c *ctx.Context, beginDate, endDate string) error {
	if len(beginDate) == 0 {
		beginDate = util.DefaultBeginDate()
	}
	if len(endDate) == 0 {
		endDate = util.Today()
	}
	if err := fetcher.GetTradeDates(c, beginDate, endDate); err != nil {
		fmt.Printf("error: fetch Latest Trade Day error: %s\n", err)
		return err
	}
	c.Params["trade_day"] = c.ResBody.GetNoKeyVals()
	return nil
}

func FetchStockPriceDayDaily() ([]string, error) {
	if err := fetchLatestTradeDay(&ctx.Context{}); err != nil {
		return nil, err
	}

	tradeDays, err := models.GetTradeDay(false, 0, 0)
	tradeDate := []string{}
	if err != nil {
		return nil, err
	}
	queue, err := queue.NewQueue("fetch_stock_daily", 50, 200, func(data interface{}) (interface{}, error) {
		c := &ctx.Context{}
		datum := data.(fetchStockDailyDatum)
		day := datum.tradeDay
		if err := fetcher.GetPrice(c, datum.code, day.Date, fetcher.Day, 1); err != nil {
			fmt.Printf("error: GetPricesByDay error: %s\n", err)
			return nil, err
		}
		if prices := models.ParsePriceInfo(c); prices != nil {
			if len(prices) > 0 && prices[0].Timestamp == day.Timestamp {
				return models.StockPriceDay{
					Code:  datum.code,
					Price: *prices[0],
				}, nil
			}
		}
		return nil, nil
	}, func(datas []interface{}) error {
		if err := models.InsertStockPriceDay(datas); err != nil {
			return err
		}
		return nil
	})
	for _, day := range tradeDays {
		for code, _ := range SecuritySet {
			queue.Push(fetchStockDailyDatum{
				code:     code,
				tradeDay: day,
			})
		}
		if err := models.UpdateTradeDay([]int64{day.Timestamp}); err != nil {
			return nil, err
		}
	}
	queue.Close()
	for _, day := range tradeDays {
		tradeDate = append(tradeDate, day.Date)
	}
	return tradeDate, nil
}

/**
 *	find stock price with code from Stock table, return a slice of
 *  stock prices
 */
func GetStockPriceByCode(code string) ([]*models.StockPriceDay, error) {
	stocks, err := models.GetStockPriceList(models.SearchOption{
		Code:    code,
		BeginAt: util.ToTimeStamp("2021-05-01"),
		EndAt:   util.ToTimeStamp("2021-05-12"),
	})
	if err != nil {
		return stocks, err
	}
	return stocks, nil
}

/**
 * verify ref date check stock price is ref correctly
 */
func VerifyRefDate(code string) error {
	vals, err := models.GetStockPriceList(models.SearchOption{
		Reversed: true,
		Limit:    1,
		Code:     code,
	})
	if err != nil {
		return err
	}
	if len(vals) == 0 {
		return updateStockPriceDayByCode(code)
	}
	price := vals[0]
	date := util.ToDate(price.Timestamp)
	prices, err := GetPricesByDay(code, date, 1)
	if err != nil {
		return err
	}
	if len(prices) == 0 {
		return fmt.Errorf("error: fetch price day fail %s, %s", code, date)
	}
	if prices[0].Open != price.Open || prices[0].Close != price.Close || prices[0].High != price.High || prices[0].Low != price.Low {
		return updateStockPriceDayByCode(code)
	}
	return nil
}

func updateStockPriceDayByCode(code string) error {
	if err := models.DeleteStockPriceDayByCode(code); err != nil {
		return err
	}
	tradeDays, err := models.GetTradeDay(true, 0, util.TodayUnix())
	if err != nil {
		return err
	}
	if len(tradeDays) == 0 {
		return fmt.Errorf("error: get trade days")
	}
	stocks, err := GetPricesByDay(code, tradeDays[0].Date, len(tradeDays))
	if err != nil {
		return err
	}
	datas := make([]interface{}, 0)
	for _, datum := range stocks {
		datas = append(datas, models.StockPriceDay{
			Code:  code,
			Price: *datum,
		})
	}
	if err := models.InsertStockPriceDay(datas); err != nil {
		return err
	}
	fmt.Printf("verify ref date update: code %s\n", code)
	return nil
}
