/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 16:36:57
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-07-27 23:06:26
 */
package service

import (
	"fmt"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
)

// Count should not be greater than 5000.
func GetPricesByDay(code string, count int64) error {
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

/**
 * fetch all stockPriceDay with specified day
s*/
func FetchStockPriceByDay(date string) error {
	_, err := models.GetSecurities()
	if err != nil {
		return err
	}
	return nil
}

/**
 * fetch trade days between begin/end date
 */
func fetchTradeDates(beginDate, endDate string) error {
	c := &ctx.Context{}
	if len(beginDate) == 0 {
		beginDate = defaultBeginDate()
	}
	if len(endDate) == 0 {
		endDate = today()
	}
	if err := fetcher.GetTradeDates(c, beginDate, endDate); err != nil {
		fmt.Printf("ERROR: GetTradeDates error: %s\n", err)
		return err
	}
	return nil
}
