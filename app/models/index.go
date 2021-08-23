/*
 * @Author: cedric.jia
 * @Date: 2021-08-18 19:18:28
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-23 22:44:02
 */

package models

import (
	"context"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HighLowIndex struct {
	Date      string `bson:"date,omitempty"`
	Timestamp int64  `bson:"timestamp, omitempty"`
	High      int    `bson:"high,omitempty"`
	Low       int    `bson:"low,omitempty"`
	Index     int    `bson:"index,omitempty"`
}

func InsertHighLowIndex(datas []interface{}) error {
	return InsertMany(datas, "high_low_index")
}

func RemoveHighLowIndex(t int64) error {
	return RemoveMany(t, "high_low_index")
}

func InitHighLowTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"high", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"low", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"index", -1}},
	})
	_, err := database.Collection("high_low_index").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}
