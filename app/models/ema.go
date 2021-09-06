/*
 * @Author: cedric.jia
 * @Date: 2021-08-17 15:51:51
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-06 09:58:43
 */

package models

import (
	"context"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Ema struct {
	Code      string  `bson:"code, omitempty"`
	Date      string  `bson:"date"`
	Timestamp int64   `bson:"timestamp, omitempty"`
	MA_6      float64 `bson:"ma_6, omitempty"`
	MA_12     float64 `bson:"ma_12, omitempty"`
	MA_26     float64 `bson:"ma_26, omitempty"`
	MA_60     float64 `bson:"ma_60, omitempty"`
}

func InsertEma(datas []interface{}) error {
	return InsertMany(datas, "ema")
}

func InitEmaTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	})
	_, err := database.Collection("ema").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func RemoveEma(t int64) error {
	return RemoveMany(t, "ema")
}

func GetEma(opt SearchOption) ([]*Ema, error) {
	var results []*Ema
	cur, err := GetCursor(opt, "ema")
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem Ema
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
