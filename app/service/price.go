/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 16:36:57
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-05 14:06:12
 */
package service

import (
	"fmt"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
)

// Count should not be greater than 5000.
func GetPricesByDay(code string, count int) error {
	c := &ctx.Context{}
	if err := fetcher.GetPrice(c, code, today(), fetcher.Day, count); err != nil {
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
	if err := fetcher.GetPrice(c, code, "2020-10-14", fetcher.Day, count); err != nil {
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
	// if err := fetchTradeDateCount("", ""); err != nil {
	// 	return err
	// }
	initStockQueue, err := queue.NewQueue("init", 50, func(data interface{}) error {
		code := data.(string)
		if err := initStockPriceByDay(code, 200); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	initStockQueue.Run()
	for i, _ := range SecuritySet {
		initStockQueue.Push(i)
	}
	initStockQueue.Close()
	return nil
}

/**
 * fetch all stockPriceDay with specified day
s*/
func FetchStockPriceByDay(date string) error {
	return nil
}

/**
 * fetch trade days between begin/end date
 */
func fetchTradeDateCount(beginDate, endDate string) error {
	c := &ctx.Context{}
	if len(beginDate) == 0 {
		beginDate = defaultBeginDate()
	}
	if len(endDate) == 0 {
		endDate = today()
	}
	if err := fetcher.GetTradeDates(c, beginDate, endDate); err != nil {
		fmt.Printf("error: GetTradeDates error: %s\n", err)
		return err
	}
	DefaultDailyBarCount = len(c.ResBody.GetVals())
	return nil
}

/**
 *	find stock price with code from Stock table, return a slice of
 *  stock prices
 */
func GetStockPriceByCode(code string) error {
	stocks, err := models.GetStockPriceByCode(code)
	if err != nil {
		return err
	}
	fmt.Println(stocks)
	return nil
}
