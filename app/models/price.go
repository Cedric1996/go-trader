/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 17:25:36
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 12:11:30
 */

package models

import (
	"strconv"
	"time"

	"github.cedric1996.com/go-trader/app/context"
)

// Price represents basic stock price info.
type Price struct {
	open      float64
	close     float64
	high      float64
	low       float64
	volume    int64
	money     int64
	paused    int64
	highLimit float64
	lowLimit  float64
	avg       float64
	preClose  float64
}

func UpdatePricesByDay(ctx *context.Ctx) error {
	resBody := ctx.ResBody
	vals := resBody.GetVals()
	const shortForm = "2020-01-02T15:04:05Z"
	res := make(map[time.Time]*Price)

	for _, val := range vals {
		t, _ := time.Parse(shortForm, val[0]+"T15:00:00Z")
		price := &Price{}
		price.open, _ = strconv.ParseFloat(val[1], 10)
		price.close, _ = strconv.ParseFloat(val[2], 10)
		price.high, _ = strconv.ParseFloat(val[3], 10)
		price.low, _ = strconv.ParseFloat(val[4], 10)
		price.volume, _ = strconv.ParseInt(val[5], 10, 64)
		price.money, _ = strconv.ParseInt(val[6], 10, 64)
		price.paused, _ = strconv.ParseInt(val[7], 10, 64)
		price.highLimit, _ = strconv.ParseFloat(val[8], 10)
		price.lowLimit, _ = strconv.ParseFloat(val[9], 10)
		price.avg, _ = strconv.ParseFloat(val[10], 10)
		price.preClose, _ = strconv.ParseFloat(val[11], 10)
		res[t] = price
	}
	return nil
}
