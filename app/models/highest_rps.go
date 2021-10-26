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
	RpsBase         RpsBase `bson:",inline"`
	Rps_250         int64   `bson:"rps_250, omitempty"`
	Rps_120         int64   `bson:"rps_120, omitempty"`
	Rps_60          int64   `bson:"rps_60, omitempty"`
	RpsIncrease_250 float64 `bson:"rps_increase_250, omitempty"`
	RpsIncrease_120 float64 `bson:"rps_increase_120, omitempty"`
	RpsIncrease_60  float64 `bson:"rps_increase_60, omitempty"`
	Net             float64 `bson:"net, omitempty"`
}

type RpsWeek struct {
	RpsBase   RpsBase `bson:",inline"`
	Rps_250   int64   `bson:"rps_250, omitempty"`
	Rps_120   int64   `bson:"rps_120, omitempty"`
	Rps_60    int64   `bson:"rps_60, omitempty"`
	Net       float64 `bson:"net, omitempty"`
	MarketCap float64 `bson:"market_cap, omitempty"`
}

type LowestRps struct {
	RpsBase  RpsBase `bson:",inline"`
	Highest  float64 `bson:"highest, omitempty"`
	Lowest   float64 `bson:"lowest, omitempty"`
	Price    float64 `bson:"price, omitempty"`
	Net      float64 `bson:"net, omitempty"`
	DrawBack float64 `bson:"drawback, omitempty"`
}

type HighestApproach struct {
	RpsBase RpsBase `bson:",inline"`
	Highest float64 `bson:"highest, omitempty"`
	Lowest  float64 `bson:"lowest, omitempty"`
	Range   float64 `bson:"range, omitempty"`
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

func InsertLowestRps(datas []interface{}) error {
	return InsertMany(datas, "lowest_rps")
}

func InsertRpsWeek(datas []interface{}) error {
	return InsertMany(datas, "rps_week")
}

func InsertHighestApproach(datas []interface{}) error {
	return InsertMany(datas, "highest_approach")
}

func InitHighestRpsTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_250", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_120", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_60", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"net", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	})
	_, err := database.Collection("highest_rps").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func InittHighestApproachTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"range", -1}},
	})
	_, err := database.Collection("highest_approach").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func InitRpsWeekTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_250", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_120", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_60", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"net", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"market_cap", -1}},
	})
	_, err := database.Collection("rps_week").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func InitLowestRpsTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"net", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"drawback", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"net", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	})
	_, err := database.Collection("lowest_rps").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
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
	}, mongo.IndexModel{
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

func DropCollection(name string) error {
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

func GetHighestApproach(opt SearchOption) ([]*HighestApproach, error) {
	var results []*HighestApproach
	cur, err := GetCursor(opt, "highest_approach")
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem HighestApproach
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
