/*
 * @Author: cedric.jia
 * @Date: 2021-08-16 12:30:24
 * @Last Modified by:   cedric.jia
 * @Last Modified time: 2021-08-16 12:30:24
 */

package models

import (
	"context"
	"fmt"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Valuation struct {
	Code                 string  `bson:"code"`
	Date                 string  `bson:"date"`
	Timestamp            int64   `bson:"timestamp"`
	Capitalization       float64 `bson:"capitalization"`
	CirculatingCap       float64 `bson:"circulating_cap"`
	MarketCap            float64 `bson:"market_cap"`
	CirculatingMarketCap float64 `bson:"circulating_market_cap"`
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

func InitFundamentalIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	})
	_, err := database.Collection("valuation").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func DeleteFundamental(timestamp int64) error {
	filter := bson.M{"timestamp": timestamp}
	results, err := database.Collection("valuation").DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	fmt.Printf("delete rps_increase data count: %d\n", results.DeletedCount)
	return nil
}

func GetValuation(opt SearchOption) ([]*Valuation, error) {
	filter := bson.M{"timestamp": opt.Timestamp}
	findOptions := options.Find().SetSort(bson.D{{"timestamp", -1}})
	results := make([]*Valuation, 0)

	cur, err := database.Collection("valuation").Find(context.TODO(), filter, findOptions)
	if err != nil {
		return results, err
	}

	for cur.Next(context.TODO()) {
		var elem Valuation
		err := cur.Decode(&elem)
		if err != nil {
			return results, err
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		return results, err
	}
	return results, nil
}
