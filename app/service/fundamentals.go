/*
 * @Author: cedric.jia
 * @Date: 2021-04-03 16:36:43
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-03 17:03:20
 */
package service

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/handler"
)

func GetFundamentalsData(table fetcher.FinTable, code, date string) map[string]string {
	fetchRes, err := fetcher.GetFundamentals(table, code, date)
	if err != nil {
		fmt.Printf("ERROR: GetFundamentalsData error: %s\n", err)
		return nil
	}
	data := handler.ParseFundamentals(fetchRes)
	return data
}
