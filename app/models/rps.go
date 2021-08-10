/*
 * @Author: cedric.jia
 * @Date: 2021-08-06 13:51:37
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-06 14:29:23
 */

package models

import (
	"context"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RpsBase struct {
	Code      string `bson:"code, omitempty"`
	Timestamp int64  `bson:"timestamp, omitempty"`
	Date      string `bson:"date, omitempty"`
}

type Rps struct {
	RpsBase RpsBase `bson:",inline"`
	Rps_120 float64 `bson:"rps_120, omitempty"`
	Rps_20  float64 `bson:"rps_20, omitempty"`
	Rps_10  float64 `bson:"rps_10, omitempty"`
	Rps_5   float64 `bson:"rps_5, omitempty"`
}

type RpsIncrease struct {
	RpsBase      RpsBase `bson:",inline"`
	Increase_120 float64 `bson:"increase_120, omitempty"`
	Increase_20  float64 `bson:"increase_20, omitempty"`
	Increase_10  float64 `bson:"increase_10, omitempty"`
	Increase_5   float64 `bson:"increase_5, omitempty"`
}

func InitRpsTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_120", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_20", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_10", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_5", -1}},
	})
	_, err := database.Collection("rps").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func InitRpsIncreaseTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"increase_120", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"increase_20", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"increase_10", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"increase_5", -1}},
	})
	_, err := database.Collection("rps_increase").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func InsertRpsIncrease(datas []interface{}) error {
	opts := options.InsertMany()
	_, err := database.Collection("rps_increase").InsertMany(context.TODO(), datas, opts)
	if err != nil {
		return err
	}
	return nil
}
