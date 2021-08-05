package models

import (
	"context"
	"fmt"

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
	res, err := database.Stock().InsertMany(context.TODO(), stocks)
	if err != nil {
		return err
	}
	fmt.Printf("code: %s, insert %d docs\n", code, len(res.InsertedIDs))
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

func GetStockPriceByCode(code string) ([]*StockPriceDay, error) {
	// Pass these options to the Find method
	findOptions := options.Find()

	// Here's an array in which you can store the decoded documents
	var results []*StockPriceDay

	// Passing bson.D{{}} as the filter matches all documents in the collection
	cur, err := database.Stock().Find(context.TODO(), bson.D{{"code", code}}, findOptions)
	if err != nil {
		return nil, err
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
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
	// Close the cursor once finished
	return results, nil
}
