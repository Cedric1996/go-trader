/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 12:26:16
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-04 18:19:42
 */
package mongo

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientInit sync.Once
var mongoClient *mongo.Client

func Client() *mongo.Client {
	clientInit.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
		defer func() {
			if err = mongoClient.Disconnect(ctx); err != nil {
				panic(err)
			}
		}()
	})
	return mongoClient
}
