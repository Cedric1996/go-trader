/*
 * @Author: cedric.jia
 * @Date: 2021-08-18 19:21:28
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-18 20:20:48
 */

package factor

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
)

const (
	NewHigh     int64 = 0
	NewLow      int64 = 1
	NormalPrice int64 = 2
)

type highLowIndexFactor struct {
	name      string
	calDate   string
	timestamp int64
	isInit    bool
}

func NewHighLowIndexFactor(name, calDate string, isInit bool) *highLowIndexFactor {
	return &highLowIndexFactor{
		name:      name,
		calDate:   calDate,
		timestamp: util.ParseDate(calDate).Unix(),
		isInit:    isInit,
	}
}

func (f *highLowIndexFactor) Run() error {
	if f.isInit {
		return f.init()
	}
	return f.execute()
}

func (f *highLowIndexFactor) execute() error {
	date, err := models.GetTradeDay(true, 1, f.timestamp)
	if err != nil {
		return err
	}
	if len(date) == 0 || date[0].Timestamp != f.timestamp {
		return fmt.Errorf("error: highest factor task date: %s", f.calDate)
	}
	queue, err := queue.NewQueue("high_low_index", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		code := data.(string)
		high, err := models.GetHighest(code, f.timestamp-24*3600, 1)
		if err != nil {
			return nil, err
		}
		low, err := models.GetLowest(code, f.timestamp-24*3600, 1)
		if err != nil {
			return nil, err
		}

		prices, err := models.GetStockPriceList(models.SearchOption{
			Code:      code,
			Timestamp: f.timestamp,
			Limit:     1,
		})
		if err != nil || len(prices) == 0 {
			return nil, err
		}
		switch {
		case high.Price < prices[0].Close:
			return NewHigh, nil
		case low.Price > prices[0].Close:
			return NewLow, nil
		default:
			return NormalPrice, nil
		}
	}, func(data []interface{}) error {
		if err := models.InsertHighLowIndex(data); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	for code, _ := range service.SecuritySet {
		queue.Push(code)
	}
	queue.Close()
	return nil
}

func (f *highLowIndexFactor) init() error {
	queue, err := queue.NewQueue("init new_high_new_low index", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		code := data.(string)
		high, err := models.GetHighest(code, f.timestamp-24*3600, 1)
		if err != nil {
			return nil, err
		}
		low, err := models.GetLowest(code, f.timestamp-24*3600, 1)
		if err != nil {
			return nil, err
		}

		prices, err := models.GetStockPriceList(models.SearchOption{
			Code:      code,
			Timestamp: f.timestamp,
			Limit:     1,
		})
		if err != nil || len(prices) == 0 {
			return nil, err
		}
		switch {
		case high.Price < prices[0].Close:
			return NewHigh, nil
		case low.Price > prices[0].Close:
			return NewLow, nil
		default:
			return NormalPrice, nil
		}
	}, func(data []interface{}) error {
		if err := models.InsertHighLowIndex(data); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	for code, _ := range service.SecuritySet {
		queue.Push(code)
	}
	queue.Close()
	return nil
}
