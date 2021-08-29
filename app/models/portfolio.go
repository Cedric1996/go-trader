/*
 * @Author: cedric.jia
 * @Date: 2021-08-25 10:47:48
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-29 21:26:36
 */

package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.cedric1996.com/go-trader/app/database"
	"github.cedric1996.com/go-trader/app/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type ExportOption struct {
	Format     string
	ExportPath string
	FileName   string
}

type Portfolio struct {
	Timestamp int64   `bson:"timestamp, omitempty"`
	Risk      float64 `bson:"risk, omitempty"`
	Inventory float64 `bson:"inventory, omitempty"`
	Available float64 `bson:"available, omitempty"`
	IsCurrent bool    `bson:"is_current, omitempty"`
	Positions []*Position
}

type Position struct {
	Code        string  `bson:"code, omitempty"`
	BeginAt     int64   `bson:"begin_at, omitempty"`
	EndAt       int64   `bson:"end_at, omitempty"`
	Volume      int64   `bson:"volume, omitempty"`
	Percent     float64 `bson:"percent, omitempty"`
	Risk        float64 `bson:"risk, omitempty"`
	DealPrice   float64 `bson:"deal_price, omitempty"`
	SellPrice   float64 `bson:"sell_price, omitempty"`
	ProfitPrice float64 `bson:"profit_price, omitempty"`
	LossPrice   float64 `bson:"loss_price, omitempty"`
	Price       float64
}

func NewPositions(datas []map[string]interface{}) error {
	portfolio, err := GetPortfolio(1)
	portfolio.IsCurrent = false
	t := time.Now().Unix()
	portfolio.Timestamp = t
	positions := []interface{}{}
	if err != nil {
		fmt.Errorf("Get portfolio from db error: %s", err)
		return err
	}
	for _, data := range datas {
		code := data["code"].(string)
		volume := data["volume"].(int64)
		price := data["deal_price"].(float64)
		profitPrice := data["profit_price"].(float64)
		lossPrice := data["loss_price"].(float64)
		position := &Position{
			Code:        code,
			Volume:      volume,
			BeginAt:     t,
			EndAt:       util.MaxInt(),
			DealPrice:   price,
			ProfitPrice: profitPrice,
			LossPrice:   lossPrice,
			SellPrice:   0,
		}
		positions = append(positions, position)
		portfolio.Available -= position.DealPrice * float64(position.Volume)
	}
	return insertPosition(positions, portfolio)
}

func insertPosition(positions []interface{}, portfolio interface{}) error {
	return database.Transaction(func(sctx mongo.SessionContext) error {
		err := sctx.StartTransaction(options.Transaction().
			SetReadConcern(readconcern.Snapshot()).
			SetWriteConcern(writeconcern.New(writeconcern.WMajority())),
		)
		if err != nil {
			return err
		}
		_, err = database.Collection("position").InsertMany(sctx, positions)
		if err != nil {
			sctx.AbortTransaction(sctx)
			log.Println("caught exception during transaction, aborting.")
			return err
		}
		_, err = database.Collection("portfolio").InsertOne(sctx, portfolio)
		if err != nil {
			sctx.AbortTransaction(sctx)
			log.Println("caught exception during transaction, aborting.")
			return err
		}
		if err := sctx.CommitTransaction(sctx); err != nil {
			return err
		}
		return nil
	})
}

func ExportPosition(opt ExportOption) error {
	return fmt.Errorf("TODO: export position")
}

func GetHoldPosition() ([]*Position, error) {
	t := time.Now().Unix()
	queryBson := bson.D{{"end_at", bson.D{{"gte", t}}}}
	cur, err := database.Collection("position").Find(context.TODO(), queryBson)
	if err != nil {
		return nil, err
	}
	var results []*Position
	for cur.Next(context.TODO()) {
		var elem Position
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func GetPosition(opt SearchOption) ([]*Position, error) {
	queryBson := bson.D{}
	sortBy := -1
	if opt.Reversed {
		sortBy = 1
	}
	findOptions := options.Find().SetSort(bson.D{{"begin_at", sortBy}}).SetLimit(opt.Limit)
	var results []*Position

	if len(opt.Code) > 0 {
		queryBson = append(queryBson, bson.E{"code", opt.Code})
	}
	if opt.EndAt > 0 {
		queryBson = append(queryBson, bson.E{"timestamp", bson.D{{"$gte", opt.BeginAt}, {"$lte", opt.EndAt}}})
	}
	if opt.Timestamp > 0 {
		queryBson = append(queryBson, bson.E{"timestamp", opt.Timestamp})
	}
	cur, err := database.Collection("position").Find(context.TODO(), queryBson, findOptions)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem Position
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func GetPortfolio(limit int64) (*Portfolio, error) {
	findOptions := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetLimit(limit)
	cur, err := database.Collection("portfolio").Find(context.TODO(), findOptions)
	if err != nil {
		return nil, err
	}
	var elem *Portfolio
	if err := cur.Decode(elem); err != nil {
		return nil, err
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}
	// data, err := json.Marshal(portfolio)
	// if err != nil {
	// 	return nil, err
	// }
	return elem, nil
}

func InitPortfolioIndex() error {
	indexModel := make([]mongo.IndexModel, 0)
	portfolioModel := make([]mongo.IndexModel, 0)

	indexModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"code", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"begin_at", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"end_at", -1}},
	})
	_, err := database.Collection("position").Indexes().CreateMany(context.Background(), indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	portfolioModel = append(indexModel, mongo.IndexModel{
		Keys: bson.D{{"timestamp", -1}},
	}, mongo.IndexModel{
		Keys: bson.D{{"is_current", -1}},
	})
	_, err = database.Collection("portfolio").Indexes().CreateMany(context.Background(), portfolioModel, &options.CreateIndexesOptions{})
	if err != nil {
		return err
	}
	return nil
}
