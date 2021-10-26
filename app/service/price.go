/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 16:36:57
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-26 12:35:13
 */
package service

import (
	"errors"
	"fmt"
	"time"

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
func GetPricesByWeek(code, date string, count int) ([]*models.Price, error) {
	c := &ctx.Context{}
	if err := fetcher.GetPrice(c, code, date, fetcher.Week, count); err != nil {
		fmt.Printf("ERROR: GetPricesByWeek error: %s\n", err)
		return nil, err
	}
	prices := models.ParsePriceHourInfo(c)
	return prices, nil
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

// Count should not be greater than 5000.
func GetPricesByPeriod(code, begin, end string) ([]*models.Price, error) {
	c := &ctx.Context{}
	if err := fetcher.GetPriceWithPeriod(c, code, fetcher.Day, begin, end); err != nil {
		fmt.Printf("ERROR: GetPriceWithPeriod error: %s\n", err)
		return nil, err
	}
	prices := models.ParsePriceInfo(c)
	return prices, nil
}

// Count should not be greater than 5000.
func GetPricesByHour(code, date string, count int) ([]*models.Price, error) {
	// date = date + " 10:00:00"
	c := &ctx.Context{}
	if err := fetcher.GetPrice(c, code, date, fetcher.ThirtyMinute, count); err != nil {
		fmt.Printf("ERROR: GetPricesByDay error: %s\n", err)
		return nil, err
	}
	prices := models.ParsePriceHourInfo(c)
	return prices, nil
}

/**
 * Init stock price date from 2018-01-01 and update
 * Stock table
 */
func InitStockPriceByDay(dates []string) error {
	initStockQueue, err := queue.NewQueue("init", dates[0], 50, 1000, func(data interface{}) (interface{}, error) {
		code := data.(string)
		stocks, err := GetPricesByPeriod(code, dates[0], dates[len(dates)-1])
		if err != nil || len(stocks) == 0 {
			return nil, errors.New("")
		}
		datas := make([]interface{}, len(stocks))
		for i, data := range stocks {
			datas[i] = models.StockPriceDay{
				Code:  code,
				Price: *data,
			}
		}
		return datas, nil
	}, func(data []interface{}) error {
		datas := make([]interface{}, 0)
		for _, v := range data {
			datas = append(datas, v.([]interface{})...)
		}
		if len(datas) == 0 {
			return nil
		}
		if err := models.InsertStockPriceDay(datas); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	for i, _ := range SecuritySet {
		initStockQueue.Push(i)
	}
	initStockQueue.Close()
	tradeDayToInsert := []interface{}{}
	for _, date := range dates {
		tradeDayToInsert = append(tradeDayToInsert, models.TradeDay{
			Date:      date,
			Timestamp: util.ToTimeStamp(date),
			IsInit:    true,
		})

	}
	if err := models.InsertTradeDay(tradeDayToInsert); err != nil {
		return fmt.Errorf("update lastest trade day error: %s", err)
	}
	return nil
}

func InitStockSecurity() error {
	stocks, err := GetNewSecurities()
	if err != nil {
		return err
	}
	initStockQueue, err := queue.NewQueue("init", "", 10, 100, func(data interface{}) (interface{}, error) {
		code := data.(string)
		stocks, err := GetPricesByDay(code, util.Today(), 1000)
		if err != nil || len(stocks) == 0 {
			return nil, errors.New("")
		}
		datas := make([]interface{}, len(stocks))
		for i, data := range stocks {
			datas[i] = models.StockPriceDay{
				Code:  code,
				Price: *data,
			}
		}
		return datas, nil
	}, func(data []interface{}) error {
		datas := make([]interface{}, 0)
		for _, v := range data {
			datas = append(datas, v.([]interface{})...)
		}
		if len(datas) == 0 {
			return nil
		}
		if err := models.InsertStockPriceDay(datas); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, stock := range stocks {
		initStockQueue.Push(stock.(*models.Stock).Code)
	}
	initStockQueue.Close()
	return nil
}

func VerifyStockPriceDay() error {
	date := "2021-08-20"
	queue, _ := queue.NewQueue("init", date, 20, 1000, func(data interface{}) (interface{}, error) {
		code := data.(string)
		// security, _ := models.GetSecurityByCode(code)
		// t := util.ParseDate(security.StartDate).Unix()
		// if t > 1505142000 {
		// 	return models.ReinitStock{
		// 		Code:      code,
		// 		Timestamp: t,
		// 		IsInit:    false,
		// 	}, nil
		// }
		time.Sleep(10 * time.Second)
		prices, err := GetPricesByDay(code, date, 960)
		fmt.Printf("init stock price count: %v, code: %v\n", len(prices), code)
		if err != nil {
			fmt.Printf("fetch stock price error: %v\n", err)
			return nil, err
		}
		res := make([]interface{}, 0)
		for _, price := range prices {
			res = append(res, models.StockPriceDay{
				Code:  code,
				Price: *price,
			})
		}
		fmt.Printf("stock price day to insert, code: %v, count: %d\n", code, len(res))
		if len(res) == 0 {
			return nil, nil
		}
		if err := models.DeleteStockPriceDayByCode(code); err != nil {
			return nil, err
		}
		if err := models.InsertStockPriceDay(res); err != nil {
			return nil, err
		}
		if err := models.RemoveHighestByCode(code); err != nil {
			return nil, err
		}
		if err := models.DeleteReinitStock(code); err != nil {
			return nil, err
		}
		return true, nil
	}, func(datas []interface{}) error {
		// if err := models.InsertReinitStockInfo(datas); err != nil {
		// 	return err
		// }
		return nil
	})
	stocks, err := models.GetReinitStock()
	if err != nil {
		return err
	}
	fmt.Printf("reinit stock count: %d\n", len(stocks))
	for _, stock := range stocks {
		queue.Push(stock.Code)
	}
	// for code, _ := range SecuritySet {
	// 	queue.Push(code)
	// }
	queue.Close()
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
	tradeDayToInsert := make([]string, 0)
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
			tradeDayToInsert = append(tradeDayToInsert, day)
		}
	}
	c.Params["tradeDayToInsert"] = tradeDayToInsert
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
	c := &ctx.Context{}
	if err := fetchLatestTradeDay(c); err != nil {
		return nil, err
	}

	tradeDays, ok := c.Params["tradeDayToInsert"].([]string)
	tradeDate := []string{}
	if !ok {
		return nil, fmt.Errorf("fetch trade dates err")
	}
	for _, day := range tradeDays {
		tradeDate = append(tradeDate, day)
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
