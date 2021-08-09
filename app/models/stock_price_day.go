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

type SearchPriceOption struct {
	Code      string
	BeginAt   int64
	EndAt     int64
	Timestamp int64
	Limit     int64
}

func UpdateStockPriceDay(c *ctx.Context) error {
	priceChan := make(chan *Price, 10)
	c.Params["priceChan"] = priceChan
	go parsePriceInfo(c)

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

func InsertStockPriceDay(c *ctx.Context) error {
	priceChan := make(chan *Price, 10)
	c.Params["priceChan"] = priceChan
	go parsePriceInfo(c)

	code := c.Params["code"]
	stocks := make([]interface{}, 0)
	for price := range priceChan {
		stocks = append(stocks, StockPriceDay{
			Code:  code.(string),
			Price: *price,
		})
	}
	if len(stocks) == 0 {
		return nil
	}
	_, err := database.Collection("stock").InsertMany(context.TODO(), stocks)
	if err != nil {
		return err
	}
	return nil
}

func InitStockTableIndexes() error {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{"code", 1}, {"timestamp", -1}},
	}
	_, err := database.Collection("stock").Indexes().CreateOne(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}

func GetStockPriceList(opt SearchPriceOption) ([]*StockPriceDay, error) {
	queryBson := bson.D{}
	findOptions := options.Find().SetSort(bson.D{{"timestamp", 1}}).SetLimit(opt.Limit)
	var results []*StockPriceDay

	if len(opt.Code) > 0 {
		queryBson = append(queryBson, bson.E{"code", opt.Code})
	}
	if opt.EndAt > 0 {
		queryBson = append(queryBson, bson.E{"timestamp", bson.D{{"$gte", opt.BeginAt}, {"$lte", opt.EndAt}}})
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