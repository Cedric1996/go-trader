/*
 * @Author: cedric.jia
 * @Date: 2021-04-24 17:54:29
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-18 12:45:58
 */

package service

import (
	"fmt"
	"sync"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/util"
)

var (
	securitySetInit      sync.Once
	SecuritySet          map[string]string
	DefaultDailyBarCount int
)

func Init() {
	securitySetInit.Do(func() {
		securities, err := models.GetAllSecurities()
		if err != nil {
			fmt.Printf("error: Init securities set error: %s\n", err)
			return
		}
		SecuritySet = make(map[string]string)
		for _, security := range securities {
			SecuritySet[security.Code] = security.StartDate
		}
	})
}

func GetAllSecurities() error {
	c := &ctx.Context{}
	if err := fetcher.GetAllSecurities(c, util.Today()); err != nil {
		fmt.Printf("error: GetAllSecurities error: %s\n", err)
		return err
	}
	securities, err := parseStockInfo(c)
	if err != nil {
		return nil
	}
	if err := models.InsertStockInfo(securities); err != nil {
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
			Code:        val[0],
			DisplayName: val[1],
			Name:        val[2],
			StartDate:   val[3],
			EndDate:     val[4],
		}
		res = append(res, stock)
	}
	return res, nil
}
