/*
 * @Author: cedric.jia
 * @Date: 2021-04-03 16:36:43
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-17 17:21:39
 */
package service

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/fetcher"
)

func GetFundamentalsData(table fetcher.FinTable, code, date string) error {
	_, err := fetcher.GetFundamentals(table, code, date)
	if err != nil {
		fmt.Printf("ERROR: GetFundamentalsData error: %s\n", err)
		return nil
	}
	return err
}
