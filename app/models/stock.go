/*
 * @Author: cedric.jia
 * @Date: 2021-04-24 12:26:15
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-22 13:14:47
 */

package models

import (
	"context"
	"log"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Stock struct {
	Code        string `bson:"code, omitempty"`
	DisplayName string `bson:"display_name"`
	Name        string `bson:"name, omitempty"`
	StartDate   string `bson:"start_date, omitempty"`
	EndDate     string `bson:"end_date, omitempty"`
}

type ReinitStock struct {
	Code      string `bson:"code, omitempty"`
	Timestamp int64  `bson:"timestamp, omitempty"`
	IsInit    bool   `bson:"init, omitempty"`
}

func GetAllSecurities() (securities []Stock, err error) {
	securities = make([]Stock, 0)
	ctx := context.Background()
	cur, err := database.Collection("stock_info").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var security Stock
		err := cur.Decode(&security)
		if err != nil {
			log.Fatal(err)
		}
		securities = append(securities, security)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return securities, nil
}

func PruneStockSecurity() error {
	ctx := context.Background()
	securities, err := GetAllSecurities()
	if err != nil {
		return err
	}
	secMap := make(map[string]string)
	for _, security := range securities {
		secMap[security.Code] = security.Code
	}
	stocks, err := database.Collection("stock").Distinct(ctx, "code", bson.D{})
	if err != nil {
		return err
	}
	for _, stock := range stocks {
		if _, ok := secMap[stock.(string)]; !ok {
			DeleteStockPriceDayByCode(stock.(string))
		}
	}
	return nil
}

func GetSecurityByCode(code string) (*Stock, error) {
	ctx := context.Background()
	stock := database.Collection("stock_info").FindOne(ctx, bson.D{{"code", code}})
	if stock.Err() != nil {
		return nil, stock.Err()
	}
	var res Stock
	stock.Decode(&res)
	return &res, nil
}

func InsertStockInfo(stocks []interface{}) error {
	return InsertMany(stocks, "stock_info")
}

func InsertReinitStockInfo(stocks []interface{}) error {
	return InsertMany(stocks, "reinit_stock")
}

func GetReinitStock() ([]*ReinitStock, error) {
	securities := make([]*ReinitStock, 0)
	ctx := context.Background()
	filter := bson.M{"init": false}
	cur, err := database.Collection("reinit_stock").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var security ReinitStock
		err := cur.Decode(&security)
		if err != nil {
			log.Fatal(err)
		}
		securities = append(securities, &security)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return securities, nil
}

func DeleteReinitStock(code string) error {
	ctx := context.Background()
	filter := bson.M{"code": code}
	_, err := database.Collection("reinit_stock").DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

func toM(v interface{}) (m *bson.M, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}
	err = bson.Unmarshal(data, &m)
	return
}

func InitStockInfoTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	})
	_, err := database.Collection("reinit_stock").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}
