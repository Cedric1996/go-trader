/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 12:26:16
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-03-14 13:03:48
 */
package database

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientInit sync.Once

// var client *DBClient

func Init() {
	clientInit.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

		defer func() {
			if err = client.Disconnect(ctx); err != nil {
				panic(err)
			}
		}()
	})
	// return err
}
