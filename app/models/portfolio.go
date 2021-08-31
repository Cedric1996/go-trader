/*
 * @Author: cedric.jia
 * @Date: 2021-08-25 10:47:48
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-30 23:29:20
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
	Timestamp int64       `bson:"timestamp, omitempty"`
	Risk      float64     `bson:"risk, omitempty"`
	Inventory float64     `bson:"inventory, omitempty"`
	Available float64     `bson:"available, omitempty"`
	IsCurrent bool        `bson:"is_current, omitempty"`
	Positions []*Position `bson:"-"`
}

type Position struct {
	Code        string  `bson:"code"`
	BeginAt     int64   `bson:"begin_at"`
	Volume      int64   `bson:"volume"`
	DealPrice   float64 `bson:"deal_price"`
	LossPrice   float64 `bson:"loss_price"`
	EndAt       int64   `bson:"end_at, omitempty"`
	SellPrice   float64 `bson:"sell_price, omitempty"`
	ProfitPrice float64 `bson:"profit_price, omitempty"`
	Price       float64 `bson:"-"`
	Percent     float64 `bson:"-"`
	Risk        float64 `bson:"-"`
}

func OpenPositions(datas []interface{}) error {
	data, _ := GetPortfolio(1)
	portfolio := data[0]
	portfolio.IsCurrent = false
	t := time.Now().Unix()
	portfolio.Timestamp = t
	positions := []interface{}{}

	for _, data := range datas {
		pos := data.(map[string]interface{})
		code := pos["code"].(string)
		price := pos["price"].(float64)
		profitPrice := pos["profit_price"].(float64)
		lossPrice := pos["loss_price"].(float64)
		volume := pos["volume"].(float64)
		position := &Position{
			Code:        code,
			Volume:      int64(volume),
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

func ClosePositions(datas []interface{}) error {
	t := time.Now().Unix()
	data, _ := GetPortfolio(1)
	portfolio := data[0]
	portfolio.IsCurrent = false
	portfolio.Timestamp = t
	positions := []interface{}{}
	updatePositions := []interface{}{}

	holdPositions, err := GetHoldPosition()
	if err != nil {
		return err
	}
	holdPositionMap := map[string]*Position{}
	for _, datum := range holdPositions {
		holdPositionMap[datum.Code] = datum
	}

	for _, data := range datas {
		pos := data.(map[string]interface{})
		code := pos["code"].(string)
		sellPrice := pos["sell_price"].(float64)
		volume := int64(pos["volume"].(float64))
		position := &Position{
			Code: code,
		}
		portfolio.Available += position.SellPrice * float64(position.Volume)
		hold := holdPositionMap[code]
		hold.EndAt = t
		hold.SellPrice = sellPrice
		if hold.Volume == volume {
			updatePositions = append(updatePositions, hold)
		} else {
			position.Volume = hold.Volume - volume
			position.DealPrice = (hold.DealPrice*float64(hold.Volume) - sellPrice*float64(volume)) / float64(position.Volume)
			position.BeginAt = t
			position.EndAt = util.MaxInt()
			position.ProfitPrice = hold.ProfitPrice
			position.LossPrice = hold.LossPrice
			positions = append(positions, position)
		}
	}
	return deletePosition(positions, updatePositions, portfolio)
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

func deletePosition(positions, updatePositions []interface{}, portfolio interface{}) error {
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
	queryBson := bson.D{{"end_at", bson.D{{"$gte", t}}}}
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

func GetPortfolio(limit int64) ([]*Portfolio, error) {
	queryBson := bson.D{}
	findOptions := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetLimit(limit)
	var results []*Portfolio
	cur, err := database.Collection("portfolio").Find(context.TODO(), queryBson, findOptions)
	if err != nil {
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var elem Portfolio
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}
		results = append(results, &elem)
	}
	return results, nil
}

func (portfolio *Portfolio) CalPortfolio() error {
	portfolio.Risk = 0.0
	portfolio.Inventory = 0.0
	portfolio.IsCurrent = true
	portfolio.Timestamp = time.Now().Unix()
	for _, position := range portfolio.Positions {
		portfolio.Inventory += position.Price * float64(position.Volume)
	}
	total := portfolio.Inventory + portfolio.Available
	for _, position := range portfolio.Positions {
		position.Percent = position.Price * float64(position.Volume) * 100 / total
		position.Risk = (position.DealPrice - position.LossPrice) / position.DealPrice * position.Percent
		portfolio.Risk += position.Risk
	}
	if err := InsertPortfolio(portfolio); err != nil {
		return err
	}
	return nil
}

func InsertPortfolio(data interface{}) error {
	if _, err := database.Collection("portfolio").InsertOne(context.TODO(), data); err != nil {
		return err
	}
	return nil
}

func InsertPositions(data []interface{}) error {
	return InsertMany(data, "position")
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
