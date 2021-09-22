/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 16:24:22
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-23 10:22:17
 */

package models

import (
	"context"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HighestRps struct {
	RpsBase RpsBase `bson:",inline"`
	Rps_20  int64   `bson:"rps_20, omitempty"`
	Rps_10  int64   `bson:"rps_10, omitempty"`
	Rps_5   int64   `bson:"rps_5, omitempty"`
}

type HighestRpsTr struct {
	Code   string  `bson:"code"`
	Start  int64   `bson:"start"`
	End    int64   `bson:"end"`
	Period int64   `bson:"period"`
	Net    float64 `bson:"net"`
}

func InsertHighestRps(datas []interface{}) error {
	return InsertMany(datas, "highest_rps")
}

func InitHighestRpsTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_20", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_10", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_5", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	})
	_, err := database.Collection("highest_rps").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func RemoveHighestRps(t int64) error {
	return RemoveMany(t, "highest_rps")
}

func InitStrategyIndexes(name string) error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"start", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"end", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"net", -1}},
	},mongo.IndexModel{
		Keys: bson.D{{"max", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"period", -1}},
	})
	_, err := database.Collection(name).Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func DropHighestRps(name string) error {
	return database.Collection(name).Drop(context.Background())
}

func GetHighestRps(opt SearchOption) ([]*HighestRps, error) {
	var results []*HighestRps
	cur, err := GetCursor(opt, "highest_rps")
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem HighestRps
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
