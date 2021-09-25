/*
 * @Author: cedric.jia
 * @Date: 2021-07-26 14:42:38
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-26 17:14:15
 */

package models

import (
	"context"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Module struct {
	Code string `bson:"code"`
	Name string `bson:"name"`
	StartDate string `bson:"start_date"`
}

type IndustryModule struct {
	Module  `bson:",inline"`
}

type ConceptModule struct {
	Module  `bson:",inline"`
	Date string `bson:"date"`
}
type StockModule struct {
	Code string `bson:"code"`
	ModuleName string `bson:"module_name"`
	StartDate string `bson:"start_date"`
	Timestamp int64 `bson:"timestamp"`
}

func RemoveStockModule(t int64) error {
	return RemoveMany(t, "stock_module")
}

func InsertStockModule(data []interface{}) error {
	return InsertMany(data, "stock_module")
}

func InitStockModuleIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"module_name", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"start_date", -1}},
	})
	_, err := database.Collection("stock_module").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func GetStockModule(opt SearchOption) ([]*StockModule, error) {
	var results []*StockModule
	cur, err := GetCursor(opt, "stock_module")
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem StockModule
		err := cur.Decode(&elem)
		if err != nil {
			return results, err
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		return results, err
	}
	return results, nil
}
