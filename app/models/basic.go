/*
 * @Author: cedric.jia
 * @Date: 2021-08-17 15:55:00
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-20 14:36:53
 */

package models

import (
	"context"
	"fmt"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
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
