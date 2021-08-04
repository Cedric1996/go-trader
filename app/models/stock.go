/*
 * @Author: cedric.jia
 * @Date: 2021-04-24 12:26:15
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-07-27 23:16:53
 */

package models

import (
	"context"
	"fmt"
	"log"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
)

type Stock struct {
	// ID        string  `bson:"_id, omitemptys"`
	Code        string `bson:"code, omitempty"`
	DisplayName string `bson:"display_name"`
	Name        string `bson:"name, omitempty"`
	StartDate   string `bson:"start_date, omitempty"`
	EndDate     string `bson:"end_date, omitempty"`
	// Price     []Price `bson:"price, omitempty"`
}

func GetAllSecurities() (securities []Stock, err error) {
	securities = make([]Stock, 0)
	ctx := context.Background()
	cur, err := database.Basic().Find(ctx, bson.D{})
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

func InsertStockInfo(stocks []interface{}) error {
	var err error
	if err := database.RemoveStockInfo(); err != nil {
		return err
	}
	if len(stocks) == 1 {
		err = database.InsertOne(stocks[0])
	} else {
		err = database.InsertMany(stocks)
	}

	if err != nil {
		return fmt.Errorf("insert stocks error: %v", err)
	}
	return nil
}

func GetStockInfoByCode(code string) (*Stock, error) {
	result := &Stock{}
	filter := bson.M{"code": code}
	err := database.Stock().FindOne(context.Background(), filter).Decode(result)
	if err != nil {
		return nil, fmt.Errorf("get stock info: %s error", code)
	}
	return result, nil
}

func toM(v interface{}) (m *bson.M, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}
	err = bson.Unmarshal(data, &m)
	return
}
