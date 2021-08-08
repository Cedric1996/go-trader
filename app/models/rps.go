/*
 * @Author: cedric.jia
 * @Date: 2021-08-06 13:51:37
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-06 14:29:23
 */

package models

import (
	"context"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Rps struct {
	Code      string  `bson:"code, omitempty"`
	timestamp int64   `bson:"timestamp, omitempty"`
	Date      string  `bson:"date, omitempty"`
	rps_120   float64 `bson:"rps_120, omitempty"`
	rps_20    float64 `bson:"rps_20, omitempty"`
	rps_10    float64 `bson:"rps_10, omitempty"`
	rps_5     float64 `bson:"rps_5, omitempty"`
}

func InitRpsTableIndexes() error {
	indexModel := make([]mongo.IndexModel, 0)
	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_120", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_20", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_10", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"rps_5", -1}},
	})
	_, err := database.Collection("rps").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}
