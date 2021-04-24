/*
 * @Author: cedric.jia
 * @Date: 2021-04-03 16:36:43
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 11:59:52
 */
package service

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
)

func GetFundamentalsData(table fetcher.FinTable, code, date string) error {
	ctx := &context.Ctx{}
	err := fetcher.GetFundamentals(ctx, table, code, date)
	if err != nil {
		fmt.Printf("ERROR: GetFundamentalsData error: %s\n", err)
		return nil
	}
	return err
}
