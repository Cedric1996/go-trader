/*
 * @Author: cedric.jia
 * @Date: 2021-08-13 14:37:24
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-28 16:40:33
 */

package models

import (
	"context"
	"fmt"
	"math"
	"sort"

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
	DealPrice 	 float64 `bson:"deal_price, omitempty"`
}

type VcpContinue struct {
	RpsBase      RpsBase `bson:",inline"`
	Period       int64   `bson:"period, omitempty"`
}
type TradeResult struct {
	Code   string  `bson:"code"`
	Start  int64   `bson:"start"`
	End    int64   `bson:"end"`
	StartDate  string   `bson:"start_date"`
	EndDate    string   `bson:"end_date"`
	Period int64   `bson:"period"`
	Net    float64 `bson:"net"`
}

func GetVcpRange(code string, endAt, period int64) (float64, error) {
	// dayTime := 24 * 3600
	highPriceDay, err := GetStockPriceList(SearchOption{Code: code, EndAt: endAt, Limit: period})
	if err != nil || len(highPriceDay) <1 {
		return 0, err
	}
	sort.Slice(highPriceDay, func(i, j int) bool{
		return highPriceDay[i].Close < highPriceDay[j].Close
	})
	beginAt := highPriceDay[0].Timestamp
	lowPriceDay, err := GetStockPriceList(SearchOption{Code: code, EndAt: endAt,  BeginAt: beginAt})
	if err != nil || len(lowPriceDay) <1{
		return 0, err
	}
	sort.Slice(highPriceDay, func(i, j int) bool{
		return highPriceDay[i].Close > highPriceDay[j].Close
	})
	return lowPriceDay[0].Close/highPriceDay[0].Close, nil
}

func GetVcpRanges(code string, endAt, beginAt int64) (float64, error) {
	priceDays, err := GetStockPriceList(SearchOption{Code: code, EndAt: endAt, BeginAt: beginAt})
	if err != nil || len(priceDays) <1 {
		return 0, err
	}
	sort.Slice(priceDays, func(i, j int) bool{
		return priceDays[i].Timestamp < priceDays[j].Timestamp
	})
	maxClose := priceDays[0].Close
	drawBack := 1.0
	for i:=1;i<len(priceDays);i++ {
		if priceDays[i].Close < maxClose {
			drawBack = math.Min(drawBack, priceDays[i].Close/maxClose)
		} else {
			maxClose = math.Max(maxClose, priceDays[i].Close)
		}
	}
	return drawBack, nil
}

func InsertVcp(datas []interface{}) error {
	return InsertMany(datas, "vcp")
}

func InsertVcpNew(datas []interface{}) error {
	return InsertMany(datas, "vcp_new")
}


func InsertVcpContinue(datas []interface{}) error {
	return InsertMany(datas, "vcp_continue")
}

func RemoveVcp(t int64) (err error) {
	return RemoveMany(t, "vcp")
}

func RemoveVcpNew(t int64) (err error) {
	return RemoveMany(t, "vcp_new")
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

func GetVcpNew(opt SearchOption) ([]*Vcp, error) {
	var results []*Vcp
	cur, err := GetCursor(opt, "vcp_new")
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

func GetVcpContinue(opt SearchOption) ([]*VcpContinue, error) {
	var results []*VcpContinue
	cur, err := GetCursor(opt, "vcp_continue")
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem VcpContinue
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
func GetTradeResult(opt SearchOption, name string) ([]*TradeResult, error) {
	var results []*TradeResult
	cur, err := GetCursor(opt, name)
	if err != nil {
		return nil, err
	}
	for cur.Next(context.TODO()) {
		var elem TradeResult
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

func GetTradeResultByDay(t int64, name string) ([]*TradeResult, error) {
	// time.Unix(t, 0)
	queryBson := bson.D{{"start", t}}
	// queryBson := bson.D{{"start", bson.D{{"$eq", time.Unix(t, 0)}}}}

	findOptions := options.Find()
	cur, err := database.Collection(name).Find(context.TODO(), queryBson, findOptions)
	if err != nil {
		return nil, err
	}
	var results []*TradeResult
	for cur.Next(context.TODO()) {
		var elem TradeResult
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

func InitVcpTableIndexes(name string) error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	},  mongo.IndexModel{
		Keys: bson.D{{"period", -1}},
	})
	_, err := database.Collection(name).Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
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
	_, err := database.Collection("vcp_ema_strategy").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}
