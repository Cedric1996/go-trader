/*
 * @Author: cedric.jia
 * @Date: 2021-08-17 15:55:00
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-06 10:39:28
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

func InsertMany(datas []interface{}, name string) error {
	opts := options.InsertMany()
	_, err := database.Collection(name).InsertMany(context.TODO(), datas, opts)
	if err != nil {
		return err
	}
	return nil
}

func RemoveMany(t int64, name string) error {
	filter := bson.M{"timestamp": t}
	count, err := database.Collection(name).DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	fmt.Printf("delete %d documents in collection: %s\n", count.DeletedCount, name)
	return nil
}

func GetCursor(opt SearchOption, name string) (*mongo.Cursor, error) {
	queryBson := bson.D{}
	sortBy := -1
	if opt.Reversed {
		sortBy = 1
	}
	findOptions := options.Find().SetSort(bson.D{{"timestamp", sortBy}}).SetLimit(opt.Limit).SetSkip(opt.Skip)
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
	if len(opt.Opts) > 0 {
		queryBson = append(queryBson, opt.Opts...)

	}
	cur, err := database.Collection(name).Find(context.TODO(), queryBson, findOptions)
	if err != nil {
		return nil, err
	}
	return cur, nil
}
