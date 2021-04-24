/*
 * @Author: cedric.jia
 * @Date: 2021-04-24 12:26:15
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 18:03:45
 */

package models

import (
	"fmt"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/database"
)

type Stock struct {
	// ID        string  `bson:"_id, omitemptys"`
	Code      string  `bson:"code, omitempty"`
	Name      string  `bson:"name, omitempty"`
	StartDate string  `bson:"start_date, omitempty"`
	EndDate   string  `bson:"end_date, omitempty"`
	Price     []Price `bson:"price, omitempty"`
}

func InsertStockInfo(c *ctx.Context) error {
	stocks, err := parseStockInfo(c)
	if err != nil {
		return err
	}

	if len(stocks) == 1 {
		err = database.InsertOne(stocks[0])
	} else {
		err = database.InsertMany(stocks)
	}

	if err != nil {
		return fmt.Errorf("insert stocks error: %v", err)
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
		stock := Stock{
			Code:      val[0],
			Name:      val[1],
			StartDate: val[3],
			EndDate:   val[4],
		}
		res = append(res, stock)
	}
	return res, nil
}
