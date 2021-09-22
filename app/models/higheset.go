/*
 * @Author: cedric.jia
 * @Date: 2021-08-12 16:55:08
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-24 09:59:24
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

type Highest struct {
	Code      string  `bson:"code, omitempty"`
	Price     float64 `bson:"price, omitempty"`
	Timestamp int64   `bson:"timestamp, omitempty"`
}

func InitHighestTableIndexes(period int64) error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"price", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	})
	_, err := database.Collection( fmt.Sprintf("highest_%d",period)).Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	_, err = database.Collection( fmt.Sprintf("highest_%d",period)).Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func InsertHighest(datas []interface{}, name string) error {
	return InsertMany(datas, name)
}

func RemoveHighest(t int64) (err error) {
	err = RemoveMany(t, "highest")
	if err == nil {
		err = RemoveMany(t, "lowest")
	}
	return err
}

func RemoveHighestByCode(code string) error {
	filter := bson.M{"code": code}
	h_count, err := database.Collection("highest").DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	l_count, err := database.Collection("lowest").DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	fmt.Printf("delete highest data code: %s, highest: %d, lowest: %d\n", code, h_count.DeletedCount, l_count.DeletedCount)
	return nil
}

func FindHighest(opt SearchOption) ([]*StockPriceDay, error) {
	queryBson := bson.D{{"code", opt.Code}, {"timestamp", bson.D{{"$lte", opt.Timestamp}}}}
	findOptions := options.Find().SetLimit(opt.Limit).SetSort(bson.D{{"timestamp", -1}})
	var results []*StockPriceDay
	cur, err := database.Collection("stock").Find(context.TODO(), queryBson, findOptions)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem StockPriceDay
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

func GetHighest(code string, period, t, count int64) ([]*Highest, error) {
	datas, err := GetHighestList(SearchOption{Code: code, EndAt: t, Limit: count}, fmt.Sprintf("highest_%d", period))
	if err != nil || len(datas) == 0 {
		return nil, err
	}
	return datas, nil
}

func GetLowest(code string, t, count int64) ([]*Highest, error) {
	datas, err := GetHighestList(SearchOption{Code: code, EndAt: t, Limit: count}, "lowest")
	if err != nil || len(datas) == 0 {
		return nil, err
	}
	return datas, nil
}

func GetHighestList(opt SearchOption, name string) ([]*Highest, error) {
	reversed := -1
	sort := bson.D{}
	if opt.Reversed {
		reversed = 1
	}
	if len(opt.SortBy) != 0 {
		sort = append(sort, bson.E{opt.SortBy, -1})
	}
	sort = append(sort, bson.E{"timestamp", reversed})
	findOptions := options.Find().SetSort(sort).SetLimit(opt.Limit)
	var results []*Highest
	queryBson := bson.D{}
	if len(opt.Code) > 0 {
		queryBson = append(queryBson, bson.E{"code", opt.Code})
	}
	if opt.EndAt > 0 || opt.BeginAt > 0 {
		scope := bson.D{}
		if opt.BeginAt > 0 {
			scope = append(scope, bson.E{"$gte", opt.BeginAt})
		}
		if opt.EndAt > 0 {
			scope = append(scope, bson.E{"$lte", opt.EndAt})
		}
		queryBson = append(queryBson, bson.E{"timestamp", scope})
	}
	if opt.Timestamp > 0 {
		queryBson = append(queryBson, bson.E{"timestamp", opt.Timestamp})
	}
	cur, err := database.Collection(name).Find(context.TODO(), queryBson, findOptions)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem Highest
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

func (s *StockPriceDay) CheckApproachHighest(code string, period, t int64, ratio float64) (bool, error) {
	// filter tradeDay close price goes beyond highest too much
	highest, err := GetHighest(code, period, t-24*3600, 1)
	if err != nil || highest == nil {
		return false, err
	}
	priceRatio := s.Close / highest[0].Price
	return priceRatio <= 1 && priceRatio >= ratio, nil
}

func (s *StockPriceDay) CheckBreakHighest(code string, t int64) (bool, error) {
	// filter tradeDay close price goes beyond highest too much
	highest, err := GetHighest(code, 120, t-24*3600, 1)
	if err != nil || highest == nil {
		return false, err
	}
	isBreak := s.Close > highest[0].Price
	return isBreak, nil
}
