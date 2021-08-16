/*
 * @Author: cedric.jia
 * @Date: 2021-08-16 12:30:24
 * @Last Modified by:   cedric.jia
 * @Last Modified time: 2021-08-16 12:30:24
 */

package models

import (
	"context"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Valuation struct {
	Code                 string  `bson:"code"`
	Date                 string  `bson:"date"`
	Timestamp            int64   `bson:"timestamp"`
	Capitalization       string  `bson:"capitalization"`
	CirculatingCap       float64 `bson:"circulating"`
	MarketCap            float64 `bson:"market"`
	CirculatingMarketCap float64 `bson:"circulating"`
	TurnoverRatio        float64 `bson:"turnover"`
}

func InsertFundamental(datas []interface{}, table string) error {
	opts := options.InsertMany()
	_, err := database.Collection(table).InsertMany(context.TODO(), datas, opts)
	if err != nil {
		return err
	}
	return nil
}
