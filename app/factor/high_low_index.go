/*
 * @Author: cedric.jia
 * @Date: 2021-08-18 19:21:28
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-23 23:43:45
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
}

func NewHighLowIndexFactor(name, calDate string) *highLowIndexFactor {
	return &highLowIndexFactor{
		name:      name,
		calDate:   calDate,
		timestamp: util.ParseDate(calDate).Unix(),
	}
}

func (f *highLowIndexFactor) Run() error {
	return f.execute()
}

func (f *highLowIndexFactor) Clean() error {
	return models.RemoveHighLowIndex(f.timestamp)
}

func (f *highLowIndexFactor) execute() error {
	dates, err := models.GetTradeDay(true, 2, f.timestamp)
	if err != nil {
		return err
	}
	queue, _ := queue.NewQueue("init new_high_new_low index", f.calDate, 50, 200, func(data interface{}) (interface{}, error) {
		dates := data.([]*models.TradeDay)
		t1 := util.ParseDate(dates[0].Date).Unix()
		t2 := util.ParseDate(dates[1].Date).Unix()
		highs, err := models.GetHighestList(models.SearchOption{
			BeginAt: t2,
			EndAt:   t1,
			SortBy:  "code",
		}, "highest")
		if err != nil {
			return nil, err
		}
		lows, err := models.GetHighestList(models.SearchOption{
			BeginAt: t2,
			EndAt:   t1,
			SortBy:  "code",
		}, "lowest")
		if err != nil {
			return nil, err
		}
		if len(lows) <= 1 || len(highs) <= 1 {
			return nil, nil
		}
		if len(lows) != len(highs) {
			fmt.Printf("cal nh_nl error, code: %s, high: %d, low: %d\n", lows[0].Code, len(highs), len(lows))
			return nil, err
		}

		newHigh, newLow := 0, 0
		for i := 0; i < len(highs)-1; i++ {
			if highs[i].Code == highs[i+1].Code {
				if highs[i].Price > highs[i+1].Price {
					newHigh++
					i++
					continue
				}
			} else {
				continue
			}
			if lows[i].Code == lows[i+1].Code {
				if lows[i].Price < lows[i+1].Price {
					newLow++
					i++
					continue
				}
			} else {
				continue
			}
			i++
		}
		return models.HighLowIndex{
			Date:      dates[0].Date,
			Timestamp: t1,
			High:      newHigh,
			Low:       newLow,
			Index:     newHigh - newLow,
		}, nil
	}, func(data []interface{}) error {
		if err := models.InsertHighLowIndex(data); err != nil {
			return err
		}
		return nil
	})
	for i := 0; i < len(dates)-1; i++ {
		queue.Push([]*models.TradeDay{dates[i], dates[i+1]})
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
		if err != nil || len(highs) != len(lows) || len(highs) == 0 || len(lows) == 0 {
			// verifyDate(highs, lows, prices)
			// fmt.Printf("verify date, code: %v\n", code)
			if err := models.RemoveHighestByCode(code); err != nil {
				return nil, err
			}
			f := NewHighestFactor("highest", f.calDate, 120)
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
