/*
 * @Author: cedric.jia
 * @Date: 2021-04-03 16:36:43
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-16 12:51:02
 */
package service

import (
	"fmt"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/util"
)

func GetFundamentalsData(table fetcher.FinTable, code, date string) error {
	c := &ctx.Context{}
	if len(date) == 0 {
		date = util.Today()
	}
	err := fetcher.GetFundamentals(c, table, code, date, 10)
	if err != nil {
		fmt.Printf("ERROR: GetFundamentalsData error: %s\n", err)
		return nil
	}
	return err
}

func GetQueryCount() error {
	c := &ctx.Context{}
	if err := fetcher.GetQueryCount(c); err != nil {
		fmt.Printf("ERROR: GetQueryCount error: %s\n", err)
		return err
	}
	fmt.Printf("Query count: %s\n", c.ResBody)
	return nil
}

func GetValuation(code, date string) ([]string, error) {
	c := &ctx.Context{}
	err := fetcher.GetFundamentals(c, fetcher.Valuation, code, date, 1)
	if err != nil {
		fmt.Printf("ERROR: GetFundamentalsData error: %s\n", err)
		return nil, nil
	}
	res := c.ResBody.GetKeys()
	return res, nil
}

func parseValuation(c *ctx.Context) ([]interface{}, error) {
	// resBody := c.ResBody
	// code := c.Params["code"]
	// res := make([]interface{}, 0)
	// if code == "" {
	// 	return nil, fmt.Errorf("parse stock info with error")
	// }
	// vals := resBody.GetVals()
	// for _, val := range vals {
	// 	stock := models.Valuation{
	// 		Code:        val[0],
	// 		DisplayName: val[1],
	// 		Name:        val[2],
	// 		StartDate:   val[3],
	// 		EndDate:     val[4],
	// 	}
	// 	res = append(res, stock)
	// }
	return nil, nil
}
