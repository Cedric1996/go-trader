/*
 * @Author: cedric.jia
 * @Date: 2021-04-07 22:10:50
 * @Last Modified by:   cedric.jia
 * @Last Modified time: 2021-04-07 22:10:50
 */
package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type Stock struct{}

func collection() *mongo.Collection {
	return GetCollectionByName("stock")
}

func (s *Stock) Insert() error {
	ctx := context.Background()
	var doc = bson.M{"_id": primitive.NewObjectID(), "hometown": "Atlanta"}
	if _, err := collection().InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("insert error: %v", err)
	}
	return nil
}
