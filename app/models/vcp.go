/*
 * @Author: cedric.jia
 * @Date: 2021-08-13 14:37:24
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-14 10:13:33
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

type Vcp struct {
	RpsBase      RpsBase `bson:",inline"`
	Period       int64   `bson:"period, omitempty"`
	HighestRatio float64 `bson:"highest_ratio, omitempty"`
	VcpRatio     float64 `bson:"vcp_ratio, omitempty"`
	Rps_120      int64   `bson:"rps_120, omitempty"`
}

func GetVcpRange(code string, timestamp, period int64) (float64, error) {
	beginAt := timestamp - period*24*3600
	highPriceDay, err := getClosePriceByPeriod(SearchPriceOption{Code: code, EndAt: timestamp, BeginAt: beginAt})
	if err != nil {
		return 0, err
	}
	beginAt = highPriceDay.Timestamp
	lowPriceDay, err := getClosePriceByPeriod(SearchPriceOption{Code: code, EndAt: timestamp, BeginAt: beginAt, Reversed: true})
	if err != nil {
		return 0, err
	}
	return 1 - lowPriceDay.Close/highPriceDay.Close, nil
}

func getClosePriceByPeriod(opt SearchPriceOption) (*StockPriceDay, error) {
	sortBy := -1
	if opt.Reversed {
		sortBy = 1
	}
	queryBson := bson.D{{"code", opt.Code}, {"timestamp", bson.D{{"$gte", opt.BeginAt}, {"$lte", opt.EndAt}}}}
	findOptions := options.FindOne().SetSort(bson.D{{"close", sortBy}})
	res := database.Collection("stock").FindOne(context.TODO(), queryBson, findOptions)
	if res.Err() != nil {
		return nil, res.Err()
	}
	var elem StockPriceDay
	err := res.Decode(&elem)
	if err != nil {
		return nil, err
	}
	return &elem, nil
}

func InsertVcp(datas []interface{}) error {
	opts := options.InsertMany()
	_, err := database.Collection("vcp").InsertMany(context.TODO(), datas, opts)
	if err != nil {
		return err
	}
	return nil
}

func GetVcpByDate(t int64) ([]*Vcp, error) {
	ctx := context.Background()
	queryBson := bson.D{{"timestamp", t}}
	findOptions := options.Find()
	cur, err := database.Collection("vcp").Find(ctx, queryBson, findOptions)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var results []*Vcp
	for cur.Next(context.TODO()) {
		var elem Vcp
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

func GetNewVcpByDate(t int64) ([]*Vcp, error) {
	oldVcps, err := GetVcpByDate(t - 3600*24)
	if err != nil {
		return nil, err
	}
	vcpMap := make(map[string]*Vcp)
	for _, v := range oldVcps {
		vcpMap[v.RpsBase.Code] = v
	}
	newVcps, err := GetVcpByDate(t)
	if err != nil {
		return nil, err
	}
	results := make([]*Vcp, 0)
	for _, v := range newVcps {
		if _, ok := vcpMap[v.RpsBase.Code]; !ok {
			results = append(results, v)
		}
	}
	return results, nil
}

func (v *Vcp) String() string {
	return fmt.Sprintf("Code: %s", v.RpsBase.Code)
}

func InitVcpTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"highest_ratio", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"vcp_ratio", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_120", -1}},
	})
	_, err := database.Collection("highest").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}
