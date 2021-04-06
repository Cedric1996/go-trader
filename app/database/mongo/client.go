/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 12:26:16
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-05 21:52:21
 */
package mongo

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Client() *mongo.Client {
	client, err := mongoClient()
	if err != nil {
		fmt.Printf("get mongo client failed: %v\n", err)
		return nil
	}
	return client
}

func mongoClient() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// uri := "mongodb://localhost:27017"
	uri := fmt.Sprintf("mongodb://%s:%s@%s/%s",
		os.Getenv("DB_MONGO_USERNAME"),
		os.Getenv("DB_MONGO_PASSWD"),
		os.Getenv("DB_MONGO_HOST"),
		os.Getenv("DB_MONGO_DB"))
	// clientOptions := options.Client().ApplyURI(uri)
	newClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := newClient.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	return newClient, nil
}

func Insert() error {
	ctx := context.Background()
	client, err := mongoClient()
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)
	opts := options.Collection()
	db := client.Database("trader")
	collection := db.Collection("stock", opts)
	var doc = bson.M{"_id": primitive.NewObjectID(), "hometown": "Atlanta"}
	if _, err := collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("insert error: %v", err)
	}
	return nil
}

func CreateCollection(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongoClient()
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)
	if client.Database("trader").CreateCollection(ctx, name, options.CreateCollection()); err != nil {
		return fmt.Errorf("create collection error: %v", err)
	}
	return nil
}
