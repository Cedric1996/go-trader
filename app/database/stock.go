/*
 * @Author: cedric.jia
 * @Date: 2021-04-07 22:10:50
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-07-27 23:06:22
 */
package database

import (
	"context"
	"fmt"
	"log"

	"github.cedric1996.com/go-trader/app/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Stock() *mongo.Collection {
	return mongodb.GetCollectionByName("stock")
}

func Basic() *mongo.Collection {
	return mongodb.GetCollectionByName("stock_info")
}


func InsertOne(data interface{}) error {
	ctx := context.Background()
	_, err := Basic().InsertOne(ctx, data)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func InsertMany(data []interface{}) error {
	ctx := context.Background()
	_, err := Basic().InsertMany(ctx, data)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func Update(filter bson.M, update bson.D) error {
	opt := &options.UpdateOptions{}
	opt.SetUpsert(true)

	updateResult, err := Stock().UpdateOne(context.TODO(), filter, update, opt)
	if err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Printf("Matched %v documents and updated %v documents.\n", updateResult.UpsertedID, updateResult.UpsertedCount)
	return nil
}

func RemoveStockInfo() error {
	ctx := context.Background()

	if err:= Basic().Drop(ctx);err != nil {
		return fmt.Errorf("drop collection: stock info with error %v", err)
	}
	return nil
}