/*
 * @Author: cedric.jia
 * @Date: 2021-04-24 17:54:29
 * @Last Modified by:   cedric.jia
 * @Last Modified time: 2021-04-24 17:54:29
 */

package service

import (
	"fmt"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
)

// Count should not be greater than 5000.
func GetStockInfo(code string) error {
	c := &ctx.Context{}
	if err := fetcher.GetSecurityInfo(c, code); err != nil {
		fmt.Printf("ERROR: GetSecurityInfo error: %s\n", err)
		return err
	}
	stocks, err := parseStockInfo(c)
	if err != nil {
		return nil
	}

	if err := models.InsertStockInfo(stocks); err != nil {
		return err
	}
	return nil
}

func parseStockInfo(c *ctx.Context) ([]interface{}, error) {
	resBody := c.ResBody
	code := c.Params["code"]
	res := make([]interface{}, 0)
	if code == "" {
		return nil, fmt.Errorf("parse stock info with error")
	}
	vals := resBody.GetVals()
	for _, val := range vals {
		stock := models.Stock{
			Code:      val[0],
			Name:      val[1],
			StartDate: val[3],
			EndDate:   val[4],
		}
		res = append(res, stock)
	}
	return res, nil
}
