/*
 * @Author: cedric.jia
 * @Date: 2021-08-22 17:12:10
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-22 17:20:13
 */

package models

import (
	"context"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TrueRange struct {
	Code      string  `bson:"code, omitempty"`
	Date      string  `bson:"date"`
	Timestamp int64   `bson:"timestamp, omitempty"`
	TR        float64 `bson:"tr, omitempty"`
	ATR       float64 `bson:"atr, omitempty"`
}

func RemoveTr(t int64) error {
	return RemoveMany(t, "tr")
}

func InsertTrueRange(data []interface{}) error {
	return InsertMany(data, "tr")
}

func InitTrueRangeTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"ATR", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"TR", -1}},
	})
	_, err := database.Collection("tr").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func GetTruesRange(opt SearchOption) ([]*TrueRange, error) {
	queryBson := bson.D{}
	sortBy := -1
	if opt.Reversed {
		sortBy = 1
	}
	findOptions := options.Find().SetSort(bson.D{{"timestamp", sortBy}}).SetLimit(opt.Limit)
	var results []*TrueRange
	if len(opt.Code) > 0 {
		queryBson = append(queryBson, bson.E{"code", opt.Code})
	}
	if opt.EndAt > 0 {
		queryBson = append(queryBson, bson.E{"timestamp", bson.D{{"$gte", opt.BeginAt}, {"$lte", opt.EndAt}}})
	}
	cur, err := database.Collection("tr").Find(context.TODO(), queryBson, findOptions)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem TrueRange
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
