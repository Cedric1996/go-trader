/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 16:36:57
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-07-26 20:36:13
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
	// c := &ctx.Context{}
	// if err := fetcher.GetPrice(c, fetcher.Day, 1); err != nil {
	// 	fmt.Printf("ERROR: GetPricesByDay error: %s\n", err)
	// 	return err
	// }
	// if err := models.UpdateStockPriceDay(c); err != nil {
	// 	return err
	// }
	return nil
}