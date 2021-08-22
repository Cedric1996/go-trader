/*
 * @Author: cedric.jia
 * @Date: 2021-08-18 19:21:28
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-22 14:19:32
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
		return f.initByCode()
	}
	return f.execute()
}

func (f *highLowIndexFactor) Clean() error {
	return models.RemoveHighLowIndex(f.timestamp)
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
		case high[0].Price < prices[0].Close:
			return NewHigh, nil
		case low[0].Price > prices[0].Close:
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
		case high[0].Price < prices[0].Close:
			return NewHigh, nil
		case low[0].Price > prices[0].Close:
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

func (f *highLowIndexFactor) initByCode() error {
	queue, _ := queue.NewQueue("init new_high_new_low index by code", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		code := data.(string)
		highs, err := models.GetHighest(code, f.timestamp, 0)
		if err != nil {
			return nil, err
		}
		lows, err := models.GetLowest(code, f.timestamp, 0)
		if err != nil {
			return nil, err
		}
		if err != nil || len(highs) != len(lows) || len(highs) == 0 {
			// verifyDate(highs, lows, prices)
			// fmt.Printf("verify date, code: %v\n", code)
			if err := models.RemoveHighestByCode(code); err != nil {
				return nil, err
			}
			f := NewHighestFactor("highest", f.calDate, 120, true)
			if err := f.Init(code); err != nil {
				return nil, err
			}
			return code, err
		}
		return nil, nil
	}, func(data []interface{}) error {
		fmt.Printf("initByCode: %d\n", len(data))
		return nil
	})
	for code, _ := range service.SecuritySet {
		queue.Push(code)
	}
	queue.Close()
	return nil
}

func verifyDate(high, low []*models.Highest, prices []*models.StockPriceDay) {
	fmt.Printf("verify date, code: %v\n", prices[0].Code)
	h_index := 0
	l_index := 0
	for i := 0; i < len(prices); i++ {
		t := prices[i].Timestamp
		if high[h_index].Timestamp == t {
			h_index++
		} else {
			fmt.Printf("highest lost data, date: %s\n", util.ToDate(t))
		}
		if low[l_index].Timestamp == t {
			l_index++
		} else {
			fmt.Printf("lowest lost data, date: %s\n", util.ToDate(t))
		}
	}
}
