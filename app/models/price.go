/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 17:25:36
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-25 00:27:39
 */

package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Price represents basic stock price info.
type Price struct {
	Timestamp uint32  `bson:"time, omitempty"`
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

func UpdatePricesByDay(c *ctx.Context) error {
	code := c.Params["code"]
	prices, err := parsePriceInfo(c)
	if err != nil {
		return err
	}

	opts := options.FindOneAndUpdate().SetUpsert(false)

	filter := bson.M{"code": code}
	update := bson.D{{"$set", bson.D{{"price", prices}}}}

	stock := &Stock{}
	err = database.Collection().FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(stock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return err
		}
	}
	fmt.Printf("updated document %v", stock)
	return nil
}

func parsePriceInfo(c *ctx.Context) ([]interface{}, error) {
	resBody := c.ResBody
	code := c.Params["code"]
	if code == "" {
		return nil, fmt.Errorf("Parse price info with error.")
	}
	vals := resBody.GetVals()
	const shortForm = "2020-01-02T15:04:05Z"
	prices := make([]interface{}, 0)
	for _, val := range vals {
		t, _ := time.Parse(time.RFC3339, val[0]+"T15:00:00Z")
		price := &Price{}
		price.Timestamp = uint32(t.Unix())
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
		// m, err := toM(price)
		// if err != nil {
		// 	return nil, err
		// }
		prices = append(prices, price)
	}
	return prices, nil
}
