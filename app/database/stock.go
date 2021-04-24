/*
 * @Author: cedric.jia
 * @Date: 2021-04-07 22:10:50
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-24 18:06:40
 */
package database

import (
	"context"
	"fmt"
	"log"

	"github.cedric1996.com/go-trader/app/database/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gopkg.in/mgo.v2/bson"
)

func Collection() *mongo.Collection {
	return mongodb.GetCollectionByName("stock")
}

func InsertOne(data interface{}) error {
	ctx := context.Background()
	_, err := Collection().InsertOne(ctx, data)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func InsertMany(data []interface{}) error {
	ctx := context.Background()
	_, err := Collection().InsertMany(ctx, data)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func Update(filter bson.M, update bson.D) error {
	opt := &options.UpdateOptions{}
	opt.SetUpsert(true)

	updateResult, err := Collection().UpdateOne(context.TODO(), filter, update, opt)
	if err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Printf("Matched %v documents and updated %v documents.\n", updateResult.UpsertedID, updateResult.UpsertedCount)
	return nil
}
