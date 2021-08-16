/*
 * @Author: cedric.jia
 * @Date: 2021-08-12 16:55:08
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-16 12:15:27
 */

package models

import (
	"context"

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

func InitHighestTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"price", -1}},
	})
	_, err := database.Collection("highest").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func InsertHighest(datas []interface{}) error {
	opts := options.InsertMany()
	_, err := database.Collection("highest").InsertMany(context.TODO(), datas, opts)
	if err != nil {
		return err
	}
	return nil
}

func FindHighest(opt SearchPriceOption) ([]*StockPriceDay, error) {
	queryBson := bson.D{{"code", opt.Code}, {"timestamp", bson.D{{"$lte", opt.Timestamp}}}}
	findOptions := options.Find().SetLimit(opt.Limit)
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

func GetHighest(code string, t int64) (*Highest, error) {
	queryBson := bson.D{{"code", code}, {"timestamp", t}}
	findOptions := options.FindOne()
	res := database.Collection("highest").FindOne(context.TODO(), queryBson, findOptions)
	if res.Err() != nil {
		return nil, res.Err()
	}
	var elem Highest
	err := res.Decode(&elem)
	if err != nil {
		return nil, err
	}
	return &elem, nil
}

func (s *StockPriceDay) CheckApproachHighest(code string, t int64, ratio float64) (bool, error) {
	// filter tradeDay close price goes beyond highest too much
	highest, err := GetHighest(code, t-24*3600)
	if err != nil || highest == nil {
		return false, err
	}
	priceRatio := s.Close / highest.Price
	return priceRatio <= (2-ratio) && priceRatio >= ratio, nil
}
