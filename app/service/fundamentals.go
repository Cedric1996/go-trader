/*
 * @Author: cedric.jia
 * @Date: 2021-04-03 16:36:43
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-07-26 20:35:47
 */
package service

import (
	"fmt"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
)

func GetFundamentalsData(table fetcher.FinTable, code, date string) error {
	c := &ctx.Context{}
	if len(date) == 0 {
		date = today()
	}
	err := fetcher.GetFundamentals(c, table, code, date, 10)
	if err != nil {
		fmt.Printf("ERROR: GetFundamentalsData error: %s\n", err)
		return nil
	}
	return err
}
