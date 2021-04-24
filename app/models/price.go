/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 17:25:36
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 18:05:45
 */

package models

import (
	"fmt"
	"log"
	"strconv"
	"time"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/database"

	"gopkg.in/mgo.v2/bson"
)

// Price represents basic stock price info.
type Price struct {
	timestamp uint32  `bson:"time, omitempty"`
	day       string  `bson:"day, omitempty"`
	open      float64 `bson:"open, omitempty"`
	close     float64 `bson:"close,omitempty"`
	high      float64 `bson:"high, omitempty"`
	low       float64 `bson:"low, omitempty"`
	volume    int64   `bson:"volume, omitempty"`
	money     int64   `bson:"money, omitempty"`
	paused    int64   `bson:"paused, omitempty"`
	highLimit float64 `bson:"highLimit, omitempty"`
	lowLimit  float64 `bson:"lowLimit, omitempty"`
	avg       float64 `bson:"avg, omitempty"`
	preClose  float64 `bson:"preClose, omitempty"`
}

func UpdatePricesByDay(c *ctx.Context) error {
	code := c.Params["code"]
	prices, err := parsePriceInfo(c)
	if err != nil {
		return err
	}
	// updateInfo := bson.D(bson.EC.SubDocument("$set", doc))
	updateInfo := bson.D{
		{"$set", bson.D{{"price", prices}}},
	}
	filter := bson.M{"code": code}
	if err := database.Update(filter, updateInfo); err != nil {
		log.Fatal("Error on updating stock prices", err)
		return err
	}
	return nil
}

func parsePriceInfo(c *ctx.Context) ([]*Price, error) {
	resBody := c.ResBody
	code := c.Params["code"]
	if code == "" {
		return nil, fmt.Errorf("Parse price info with error.")
	}
	vals := resBody.GetVals()
	const shortForm = "2020-01-02T15:04:05Z"
	prices := make([]*Price, 0)

	for _, val := range vals {
		t, _ := time.Parse(time.RFC3339, val[0]+"T15:00:00Z")
		price := &Price{}
		price.timestamp = uint32(t.Unix())
		price.day = t.String()
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
		prices = append(prices, price)
	}
	return prices, nil
}
