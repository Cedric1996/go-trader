/*
 * @Author: cedric.jia
 * @Date: 2021-08-08 10:03:45
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-08 10:04:07
 */

package models

import (
	"context"
	"fmt"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TradeDay struct {
	Timestamp int64  `bson:"timestamp"`
	Date      string `bson:"date"`
	IsInit    bool   `bson:"is_init, default=false"`
}

func InsertTradeDay(days []interface{}) error {
	res, err := database.Collection("trade_day").InsertMany(context.TODO(), days)
	if err != nil {
		return err
	}
	fmt.Printf("init trade day count: %d\n", len(res.InsertedIDs))
	return nil
}

func GetTradeDay(isInit bool) ([]*TradeDay, error) {
	queryBson := bson.D{}
	findOptions := options.Find().SetSort(bson.D{{"timestamp", 1}})
	results := make([]*TradeDay, 0)
	if isInit {
		queryBson = append(queryBson, bson.E{"is_init", false})

	}
	cur, err := database.Collection("trade_day").Find(context.TODO(), queryBson, findOptions)
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
