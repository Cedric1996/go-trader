package models

import (
	"context"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StockPriceDay struct {
	Price `bson:",inline"`
	Code  string `bson:"code, omitempty"`
}

type SearchOption struct {
	Code      string
	BeginAt   int64
	EndAt     int64
	Timestamp int64
	Limit     int64
	Reversed  bool
	SortBy    string
	Skip      int64
	Opts []bson.E
}

func UpdateStockPriceDay(c *ctx.Context) error {
	priceChan := make(chan *Price, 10)
	c.Params["priceChan"] = priceChan
	go ParsePriceInfo(c)

	code := c.Params["code"]
	opts := options.FindOneAndUpdate().SetUpsert(true)
	for price := range priceChan {
		stock := &StockPriceDay{
			Code:  code.(string),
			Price: *price,
		}
		filter := bson.M{"code": code, "timestamp": price.Timestamp}
		update := bson.D{{"$set", stock}}
		err := database.Collection("stock").FindOneAndUpdate(context.TODO(), filter, update, opts).Err()
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}
	}
	return nil
}

func InsertStockPriceDay(stocks []interface{}) error {
	_, err := database.Collection("stock").InsertMany(context.TODO(), stocks)
	if err != nil {
		return err
	}
	return nil
}

func InitStockTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"high", -1}},
	})
	_, err := database.Collection("stock").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func GetStockPriceList(opt SearchOption) ([]*StockPriceDay, error) {
	queryBson := bson.D{}
	sortBy := -1
	if opt.Reversed {
		sortBy = 1
	}
	findOptions := options.Find().SetSort(bson.D{{"timestamp", sortBy}}).SetLimit(opt.Limit).SetSkip(opt.Skip)
	var results []*StockPriceDay

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

func DeleteStockPriceDayByDay(t int64) error {
	return RemoveMany(t, "stock")
}

func DeleteStockPriceDayByCode(code string) error {
	filter := bson.M{"code": code}
	_, err := database.Collection("stock").DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (s *StockPriceDay) GetVolume() float64 {
	return float64(s.Volume) * s.Avg / (1000 * 1000 * 100.0)
}
