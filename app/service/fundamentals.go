/*
 * @Author: cedric.jia
 * @Date: 2021-04-03 16:36:43
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-18 20:21:06
 */
package service

import (
	"errors"
	"fmt"
	"strconv"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/util"
)

func GetCurrentToken() (string, error) {
	c := &ctx.Context{}
	if err := fetcher.GetCurrentToken(c); err != nil {
		fmt.Printf("ERROR: GetCurrentToken error: %s\n", err)
		return "", err
	}
	return c.ResBody.GetNoKeyVals()[0], nil
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
	resBody := c.ResBody
	code := c.Params["code"]
	res := make([]interface{}, 0)
	if code == "" {
		return nil, fmt.Errorf("parse stock info with error")
	}
	vals := resBody.GetVals()
	for _, val := range vals {
		if len(val) < 17 {
			fmt.Println(val)
			continue
		}
		data := models.Valuation{
			Code: val[1],
			Date: val[5],
		}
		data.Timestamp = util.ParseDate(val[5]).Unix()
		data.Capitalization, _ = strconv.ParseFloat(val[14], 10)
		data.CirculatingCap, _ = strconv.ParseFloat(val[16], 10)
		data.MarketCap, _ = strconv.ParseFloat(val[15], 10)
		data.CirculatingMarketCap, _ = strconv.ParseFloat(val[17], 10)
		data.TurnoverRatio, _ = strconv.ParseFloat(val[10], 10)
		res = append(res, data)
	}
	return res, nil
}

func InitFundamental(code, date string, count int) ([]interface{}, error) {
	c := &ctx.Context{}
	if err := fetcher.GetFundamentals(c, fetcher.Valuation, code, date, count); err != nil {
		fmt.Printf("ERROR: GetPricesByDay error: %s\n", err)
		return nil, err
	}
	datas, err := parseValuation(c)
	if err != nil {
		return nil, err
	}
	if len(datas) == 0 {
		return nil, errors.New("fetch no fundamental data")
	}
	return datas, nil
}
