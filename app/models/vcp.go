/*
 * @Author: cedric.jia
 * @Date: 2021-08-13 14:37:24
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-06 08:58:02
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

type VcpTr struct {
	Code   string  `bson:"code"`
	Start  int64   `bson:"start"`
	End    int64   `bson:"end"`
	Period int64   `bson:"period"`
	Net    float64 `bson:"net"`
}

func GetVcpRange(code string, timestamp, period int64) (float64, error) {
	dayTime := 24 * 3600
	endAt := timestamp - int64(dayTime)
	beginAt := endAt - period*int64(dayTime)
	highPriceDay, err := getClosePriceByPeriod(SearchOption{Code: code, EndAt: endAt, BeginAt: beginAt})
	if err != nil {
		return 0, err
	}
	beginAt = highPriceDay.Timestamp
	lowPriceDay, err := getClosePriceByPeriod(SearchOption{Code: code, EndAt: endAt, BeginAt: beginAt, Reversed: true})
	if err != nil {
		return 0, err
	}
	return 1 - lowPriceDay.Close/highPriceDay.Close, nil
}

func getClosePriceByPeriod(opt SearchOption) (*StockPriceDay, error) {
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
	return InsertMany(datas, "vcp")
}

func RemoveVcp(t int64) (err error) {
	return RemoveMany(t, "vcp")
}

func GetVcp(opt SearchOption) ([]*Vcp, error) {
	var results []*Vcp
	cur, err := GetCursor(opt, "vcp")
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem Vcp
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

func GetVcpTr(opt SearchOption, name string) ([]*VcpTr, error) {
	var results []*VcpTr
	cur, err := GetCursor(opt, name)
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem VcpTr
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

func GetVcpTrByDay(t int64, name string) ([]*VcpTr, error) {
	// time.Unix(t, 0)
	queryBson := bson.D{{"start", t}}
	// queryBson := bson.D{{"start", bson.D{{"$eq", time.Unix(t, 0)}}}}

	findOptions := options.Find()
	cur, err := database.Collection(name).Find(context.TODO(), queryBson, findOptions)
	if err != nil {
		return nil, err
	}
	var results []*VcpTr
	for cur.Next(context.TODO()) {
		var elem VcpTr
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

func GetNewVcpByDate(t1, t2 int64) ([]*Vcp, error) {
	oldVcps, err := GetVcpByDate(t2)
	if err != nil {
		return nil, err
	}
	vcpMap := make(map[string]*Vcp)
	for _, v := range oldVcps {
		vcpMap[v.RpsBase.Code] = v
	}
	newVcps, err := GetVcpByDate(t1)
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
	// return newVcps, nil
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
	_, err := database.Collection("vcp").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func InitVcpTrIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"start", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"end", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"net", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"period", -1}},
	})
	_, err := database.Collection("vcp_tr_strategy_02").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}
