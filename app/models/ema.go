/*
 * @Author: cedric.jia
 * @Date: 2021-08-17 15:51:51
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-12 21:07:28
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

type Ma struct {
	Code string  `bson:"code"`
	Timestamp int64   `bson:"timestamp"`
	MA_50    float64   `bson:"ma_50"`
	LongTrend bool `bson:"long_trend"`
}

func InsertEma(datas []interface{}) error {
	return InsertMany(datas, "ema")
}

func InsertMa(datas []interface{}) error {
	return InsertMany(datas, "ma")
}

func InsertMaIndex(datas []interface{}) error {
	return InsertMany(datas, "ma_index")
}

func InitEmaTableIndexes(name string) error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	},mongo.IndexModel{
		Keys: bson.D{{"ma_50", -1}},
	},mongo.IndexModel{
		Keys: bson.D{{"long_trend", -1}},
	})
	_, err := database.Collection(name).Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}


func InitMaIndexTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"index", -1}},
	})
	_, err := database.Collection("ma_index").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}


func RemoveEma(t int64) error {
	return RemoveMany(t, "ema")
}

func RemoveMa(t int64) error {
	return RemoveMany(t, "ma")
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

func GetMa(opt SearchOption) ([]*Ma, error) {
	var results []*Ma
	cur, err := GetCursor(opt, "ma")
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem  Ma
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