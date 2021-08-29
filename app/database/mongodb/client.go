/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 12:26:16
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-29 21:13:14
 */
package mongodb

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	databaseSync  sync.Once
	mongoDatabase *mongo.Database
	mongoClient   *mongo.Client
)

func ConnectMongoClient() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/trader",
		viper.GetString("mongo.username"),
		viper.GetString("mongo.password"),
		viper.GetString("mongo.hostname"),
		viper.GetString("mongo.port"))
	mongoClient, _ = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}
}

func client() *mongo.Client {
	return mongoClient
}

func database() *mongo.Database {
	databaseSync.Do(func() {
		mongoDatabase = client().Database(viper.GetString("mongo.database"))
	})
	return mongoDatabase
}

func GetCollectionByName(name string) *mongo.Collection {
	return database().Collection(name)
}

func Transaction(fn func(sctx mongo.SessionContext) error) error {
	ctx := context.Background()
	return client().UseSession(ctx, fn)
}
