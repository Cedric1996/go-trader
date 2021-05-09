/*
 * @Author: cedric.jia
 * @Date: 2021-04-24 12:26:15
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-25 00:22:53
 */

package models

import (
	"context"
	"fmt"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
)

type Stock struct {
	// ID        string  `bson:"_id, omitemptys"`
	Code      string  `bson:"code, omitempty"`
	Name      string  `bson:"name, omitempty"`
	StartDate string  `bson:"start_date, omitempty"`
	EndDate   string  `bson:"end_date, omitempty"`
	Price     []Price `bson:"price, omitempty"`
}

func InsertStockInfo(stocks []interface{}) error {
	var err error
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
	err := database.Collection().FindOne(context.Background(), filter).Decode(result)
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
