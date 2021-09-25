/*
 * @Author: cedric.jia
 * @Date: 2021-08-12 16:55:08
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-28 16:34:44
 */

package models

import (
	"context"
	"fmt"
	"sort"

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

func InitHighestTableIndexes(period string) error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"price", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	})
	_, err := database.Collection( fmt.Sprintf("highest_%s",period)).Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	_, err = database.Collection( fmt.Sprintf("highest_%s",period)).Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func InsertHighest(datas []interface{}, name string) error {
	return InsertMany(datas, name)
}

func RemoveHighest(t, period int64) (err error) {
	err = RemoveMany(t, fmt.Sprintf("lowest_%d", period))
	if err == nil {
		err = RemoveMany(t, fmt.Sprintf("highest_%d", period))
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

func GetHighest(code ,period string, t, count int64) ([]*Highest, error) {
	datas, err := GetHighestList(SearchOption{Code: code, EndAt: t, Limit: count}, fmt.Sprintf("highest_%s", period))
	if err != nil || len(datas) == 0 {
		return nil, err
	}
	return datas, nil
}

func GetLowest(code ,period string, t, count int64) ([]*Highest, error) {
	datas, err := GetHighestList(SearchOption{Code: code, EndAt: t, Limit: count},  fmt.Sprintf("highest_%s", period))
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

func (s *StockPriceDay) CheckApproachHighest(code string, period int64, ratio float64) (bool, float64, error) {
	// filter tradeDay close price goes beyond highest too much
	t := s.Timestamp
	highest60, err := GetHighest(code, "60", t-24*3600, 1)
	if err != nil || highest60 == nil {
		return false, 0, err
	}

	highest60_120, err := GetHighest(code, "60_120", t-24*3600, 1)
	if err != nil || highest60_120 == nil {
		return false, 0, err
	}
	// maxClose := math.Max(highest60[0].Price, highest60_120[0].Price)
	// minClose := math.Min(highest60[0].Price, highest60_120[0].Price)

	highRatio := highest60[0].Price / highest60_120[0].Price
	// return  s.Close >= maxClose*ratio && s.Close >= minClose && s.Close <= maxClose && highRatio < 1.20, nil
	high := s.Close / highest60[0].Price
	return  high >= ratio && high <= 1 && highRatio > 0.90, highRatio, nil
}

func (s *StockPriceDay) CheckBreakHighest(code , period string, t int64) (bool, float64, error) {
	// filter tradeDay close price goes beyond highest too much
	highest, err := GetHighest(code, period, t-1, 120)
	if err != nil || len(highest) < 1 {
		return false, 0.0, err
	}
	sort.Slice(highest, func(i, j int) bool{
		return highest[i].Price > highest[j].Price
	})
	isBreak := s.Close > highest[0].Price && s.PreClose < highest[0].Price
	return isBreak, highest[0].Price, nil
}
