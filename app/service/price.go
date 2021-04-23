/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 16:36:57
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-23 23:01:26
 */
package service

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
)

// Count should not be greater than 5000.
func GetPricesByDay(code string, count int64) error {
	ctx := &context.Ctx{}
	resBody, err := fetcher.GetPrice(ctx, code, fetcher.Day, count)
	if err != nil {
		fmt.Printf("ERROR: GetPricesByDay error: %s\n", err)
		return nil
	}
	err = models.UpdatePricesByDay(code, resBody)
	return nil
}
