/*
 * @Author: cedric.jia
 * @Date: 2021-08-08 10:03:45
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-17 15:56:58
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

type TradeDay struct {
	Timestamp int64  `bson:"timestamp"`
	Date      string `bson:"date"`
	IsInit    bool   `bson:"is_init, default=false"`
}

func InsertTradeDay(days []interface{}) error {
	return InsertMany(days, "trade_day")
}

func GetTradeDay(isInit bool, limit, timestamp int64) ([]*TradeDay, error) {
	filter := bson.M{"is_init": isInit}
	findOptions := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetLimit(limit)
	results := make([]*TradeDay, 0)
	if timestamp > 0 {
		filter["timestamp"] = bson.M{"$lte": timestamp}
	}
	cur, err := database.Collection("trade_day").Find(context.TODO(), filter, findOptions)
	if err != nil {
		return results, err
	}

	for cur.Next(context.TODO()) {
		var elem TradeDay
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

func UpdateTradeDay(days []int64) error {
	update := bson.D{{"$set", bson.D{{"is_init", true}}}}
	for _, day := range days {
		filter := bson.M{"timestamp": bson.M{"$eq": day}}
		if _, err := database.Collection("trade_day").UpdateMany(context.TODO(), filter, update); err != nil {
			return err
		}
	}
	fmt.Printf("update trade day count: %d\n", len(days))
	return nil
}

func GetTradeDayByPeriod(limit int64, timestamp int64) ([]int64, error) {
	filter := bson.M{"is_init": true, "timestamp": bson.M{"$lte": timestamp}}
	findOptions := options.Find().SetLimit(limit).SetSort(bson.D{{"timestamp", -1}})
	results := make([]int64, 0)

	cur, err := database.Collection("trade_day").Find(context.TODO(), filter, findOptions)
	if err != nil {
		return results, err
	}

	for cur.Next(context.TODO()) {
		var elem TradeDay
		err := cur.Decode(&elem)
		if err != nil {
			return results, err
		}
		results = append(results, elem.Timestamp)
	}

	if err := cur.Err(); err != nil {
		return results, err
	}
	return results, nil
}

func InitTradeDayTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"is_init", -1}},
	})
	_, err := database.Collection("trade_day").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}
