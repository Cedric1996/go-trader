/*
 * @Author: cedric.jia
 * @Date: 2021-08-06 13:51:37
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-23 11:07:34
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

type RpsBase struct {
	Code      string `bson:"code, omitempty"`
	Timestamp int64  `bson:"timestamp, omitempty"`
	Date      string `bson:"date, omitempty"`
}

type Rps struct {
	RpsBase RpsBase `bson:",inline"`
	Rps_120 int64   `bson:"rps_120, omitempty"`
	Rps_20  int64   `bson:"rps_20, omitempty"`
	Rps_10  int64   `bson:"rps_10, omitempty"`
	Rps_5   int64   `bson:"rps_5, omitempty"`
}

type RpsIncrease struct {
	RpsBase      RpsBase `bson:",inline"`
	Increase_120 float64 `bson:"increase_120, omitempty"`
	Increase_20  float64 `bson:"increase_20, omitempty"`
	Increase_10  float64 `bson:"increase_10, omitempty"`
	Increase_5   float64 `bson:"increase_5, omitempty"`
}

type RpsOption struct {
	Code      string
	Timestamp int64
	SortBy    string
}

func InitRpsTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
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

func InsertRps(datas []interface{}, name string) error {
	return InsertMany(datas, name)
}

func RemoveRps(t int64) (err error) {
	err = RemoveMany(t, "rps")
	if err == nil {
		err = RemoveMany(t, "rps_increase")
	}
	return err
}

func GetRps(t, period int64) ([]*Rps, error) {
	key := fmt.Sprintf("rps_%d", period)
	queryBson := bson.D{{"timestamp", t}, {key, bson.D{{"$gte", 85}}}}
	var results []*Rps
	cur, err := database.Collection("rps").Find(context.TODO(), queryBson)
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem Rps
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

func GetRpsByOpt(opt SearchOption) ([]*Rps, error) {
	var results []*Rps
	cur, err := GetCursor(opt, "rps")
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem Rps
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

func GetRpsIncrease(opt SearchOption) ([]*RpsIncrease, error) {
	queryBson := bson.D{}
	findOptions := options.Find().SetSort(bson.D{{"timestamp", 1}})
	var results []*RpsIncrease
	if opt.Limit > 0 {
		findOptions.SetLimit(opt.Limit)
	}
	if len(opt.SortBy) > 0 {
		findOptions.SetSort(bson.D{{opt.SortBy, -1}})
	}
	if len(opt.Code) > 0 {
		queryBson = append(queryBson, bson.E{"code", opt.Code})
	}
	if opt.Timestamp > 0 {
		queryBson = append(queryBson, bson.E{"timestamp", opt.Timestamp})
	}
	cur, err := database.Collection("rps_increase").Find(context.TODO(), queryBson, findOptions)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem RpsIncrease
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

func DeleteRpsIncrease(timestamp int64) error {
	filter := bson.M{"timestamp": timestamp}
	results, err := database.Collection("rps_increase").DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	fmt.Printf("delete rps_increase data count: %d\n", results.DeletedCount)
	return nil
}
