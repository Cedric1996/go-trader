/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 17:25:36
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-30 16:16:18
 */

package models

import (
	"fmt"
	"strconv"
	"time"

	ctx "github.cedric1996.com/go-trader/app/context"
)

// Price represents basic stock price info.
type Price struct {
	Timestamp int64   `bson:"timestamp, omitempty"`
	Day       string  `bson:"day, omitempty"`
	Open      float64 `bson:"open, omitempty"`
	Close     float64 `bson:"close,omitempty"`
	High      float64 `bson:"high, omitempty"`
	Low       float64 `bson:"low, omitempty"`
	Volume    int64   `bson:"volume, omitempty"`
	Money     int64   `bson:"money, omitempty"`
	Paused    int64   `bson:"paused, omitempty"`
	HighLimit float64 `bson:"highLimit, omitempty"`
	LowLimit  float64 `bson:"lowLimit, omitempty"`
	Avg       float64 `bson:"avg, omitempty"`
	PreClose  float64 `bson:"preClose, omitempty"`
}

func ParsePriceInfo(c *ctx.Context) []*Price {
	resBody := c.ResBody
	code := c.Params["code"]
	if code == "" {
		fmt.Errorf("Parse price info with error.")
		return nil
	}
	vals := resBody.GetVals()
	prices := make([]*Price, 0)
	for _, val := range vals {
		if len(val) < 12 {
			continue
		}
		t, _ := time.Parse(time.RFC3339, val[0]+"T15:00:00Z")
		price := &Price{}
		price.Timestamp = t.Unix()
		price.Day = t.String()
		price.Open, _ = strconv.ParseFloat(val[1], 10)
		price.Close, _ = strconv.ParseFloat(val[2], 10)
		price.High, _ = strconv.ParseFloat(val[3], 10)
		price.Low, _ = strconv.ParseFloat(val[4], 10)
		price.Volume, _ = strconv.ParseInt(val[5], 10, 64)
		price.Money, _ = strconv.ParseInt(val[6], 10, 64)
		price.Paused, _ = strconv.ParseInt(val[7], 10, 64)
		price.HighLimit, _ = strconv.ParseFloat(val[8], 10)
		price.LowLimit, _ = strconv.ParseFloat(val[9], 10)
		price.Avg, _ = strconv.ParseFloat(val[10], 10)
		price.PreClose, _ = strconv.ParseFloat(val[11], 10)
		prices = append(prices, price)
	}
	return prices
}

func ParseCurrentPrice(c *ctx.Context) float64 {
	resBody := c.ResBody
	code := c.Params["code"]
	if code == "" {
		fmt.Errorf("Parse price info with error.")
		return 0
	}
	val := resBody.GetVals()
	if len(val) == 0 {
		fmt.Errorf("Parse price info with error: %v", code)
	}
	price, _ := strconv.ParseFloat(val[0][2], 64)
	return price
}
