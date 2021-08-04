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

func UpdateStockPriceDay(c *ctx.Context) error {
	priceChan := make(chan *Price, 10)
	c.Params["priceChan"] = priceChan
	go parsePriceInfo(c)

	code := c.Params["code"]
	opts := options.FindOneAndUpdate().SetUpsert(true)
	updateCount := 0

	for price := range priceChan {
		stock := &StockPriceDay{
			Code:  code.(string),
			Price: *price,
		}
		filter := bson.M{"code": code, "timestamp": price.Timestamp}
		update := bson.D{{"$set", stock}}
		err := database.Stock().FindOneAndUpdate(context.TODO(), filter, update, opts).Err()
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}
		updateCount++
	}

	// fmt.Printf("updated document %v.\n", updateCount)
	return nil
}

func InsertStockPriceDay(c *ctx.Context) error {
	return nil
}

func InitStockTableIndexes() error {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{"code", 1}, {"timestamp", -1}},
	}
	_, err := database.Stock().Indexes().CreateOne(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}
