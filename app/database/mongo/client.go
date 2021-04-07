/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 12:26:16
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-07 22:25:38
 */
package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func client() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/trader",
		viper.GetString("mongo.username"),
		viper.GetString("mongo.password"),
		viper.GetString("mongo.hostname"),
		viper.GetString("mongo.port"))
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil
	}

	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		return nil
	}
	return mongoClient
}

func database() *mongo.Database {
	return client().Database(viper.GetString("mongo.database"))
}

func GetCollectionByName(name string) *mongo.Collection {
	return database().Collection(name)
}
