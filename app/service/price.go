/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 16:36:57
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-06 14:55:40
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

// Count should not be greater than 5000.
func GetPricesByDay(code string, count int) error {
	c := &ctx.Context{}
	if err := fetcher.GetPrice(c, code, util.Today(), fetcher.Day, count); err != nil {
		fmt.Printf("ERROR: GetPricesByDay error: %s\n", err)
		return err
	}
	if err := models.UpdateStockPriceDay(c); err != nil {
		return err
	}
	return nil
}

func initStockPriceByDay(code string, count int) error {
	c := &ctx.Context{}
	if err := fetcher.GetPrice(c, code, "2018-07-09", fetcher.Day, count); err != nil {
		fmt.Printf("ERROR: GetPricesByDay error: %s\n", err)
		return err
	}
	if err := models.InsertStockPriceDay(c); err != nil {
		return err
	}
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
	tradeDays, err := models.GetTradeDay(false)
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

func FetchStockPriceDayDaily() error {
	if err := fetchLatestTradeDay(&ctx.Context{}); err != nil {
		return err
	}

	tradeDays, err := models.GetTradeDay(true)
	if err != nil {
		return err
	}
	fetchStockPriceQueue, err := queue.NewQueue("fetch_stock_price", 50, 10, func(data interface{}) (interface{}, error) {
		day := data.(string)
		if err := fetchStockPriceByDay(day); err != nil {
			return nil, err
		}
		return nil, nil
	}, func(datas []interface{}) error {
		return nil
	})
	if err != nil {
		return err
	}
	for _, day := range tradeDays {
		fetchStockPriceQueue.Push(day.Date)
	}
	fetchStockPriceQueue.Close()
	return nil
}

func fetchStockPriceByDay(day string) error {
	for code, _ := range SecuritySet {
		c := &ctx.Context{}
		if err := fetcher.GetPrice(c, code, day, fetcher.Day, 1); err != nil {
			fmt.Printf("error: GetPricesByDay error: %s\n", err)
			return err
		}
		if err := models.InsertStockPriceDay(c); err != nil {
			return err
		}
	}
	return nil
}

/**
 *	find stock price with code from Stock table, return a slice of
 *  stock prices
 */
func GetStockPriceByCode(code string) ([]*models.StockPriceDay, error) {
	beginAt := util.ToTimeStamp("2021-08-06")
	endAt := util.ToTimeStamp("2021-08-07")

	stocks, err := models.GetStockPriceList(models.SearchPriceOption{
		// Code:    code,
		BeginAt: beginAt,
		EndAt:   endAt,
	})
	if err != nil {
		return stocks, err
	}
	return stocks, nil
}
