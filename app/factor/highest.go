/*
 * @Author: cedric.jia
 * @Date: 2021-08-12 11:19:31
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-21 23:41:39
 */

package factor

import (
	"fmt"
	"math"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
)

type highestFactor struct {
	name      string
	calDate   string
	period    int64
	isLowest  bool
	timestamp int64
}

func NewHighestFactor(name string, calDate string, period int64, isLowest bool) *highestFactor {
	return &highestFactor{
		name:      name,
		calDate:   calDate,
		period:    period,
		isLowest:  isLowest,
		timestamp: util.ParseDate(calDate).Unix(),
	}
}

func (f *highestFactor) Run() error {
	if err := f.execute(); err != nil {
		return err
	}
	return nil
}

func (f *highestFactor) Init(code string) error {
	prices, err := models.GetStockPriceList(models.SearchOption{
		Code:     code,
		Reversed: true,
	})
	if err != nil {
		return err
	}
	period := int(f.period)
	calMaxAndMin := func(prices []*models.StockPriceDay) (max, min float64) {
		min = math.Inf(1)
		max = 0.0
		for _, p := range prices {
			min = math.Min(p.Close, min)
			max = math.Max(p.High, max)
		}
		return max, min
	}
	highest := []interface{}{}
	lowest := []interface{}{}
	for i := 0; i < len(prices)-period; i++ {
		max, min := calMaxAndMin(prices[i : i+period])
		highest = append(highest, models.Highest{
			Code:      code,
			Price:     max,
			Timestamp: prices[i+period-1].Timestamp,
		})
		lowest = append(lowest, models.Highest{
			Code:      code,
			Price:     min,
			Timestamp: prices[i+period-1].Timestamp,
		})
	}
	if err := models.InsertHighest(highest, "highest"); err != nil {
		return err
	}
	if err := models.InsertHighest(lowest, "lowest"); err != nil {
		return err
	}
	return nil
}

func (f *highestFactor) Clean() error {
	return models.RemoveHighest(f.timestamp, f.isLowest)
}

func (f *highestFactor) execute() error {
	day, err := models.GetTradeDay(true, 1, f.timestamp)
	if err != nil {
		return err
	}
	if len(day) == 0 || day[0].Timestamp != f.timestamp {
		return fmt.Errorf("error: highest factor task date: %s", f.calDate)
	}
	queue, err := queue.NewQueue(f.name, f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		code := data.(string)
		prices, err := models.FindHighest(models.SearchOption{
			Code:      code,
			Limit:     f.period,
			Timestamp: f.timestamp,
		})
		if err != nil || len(prices) < int(f.period) {
			return nil, err
		}

		calMaxOrMin := func() float64 {
			if f.isLowest {
				min := math.Inf(1)
				for _, p := range prices {
					min = math.Min(p.Close, min)
				}
				return min
			} else {
				max := 0.0
				for _, p := range prices {
					max = math.Max(p.High, max)
				}
				return max
			}
		}
		return models.Highest{
			Code:      prices[0].Code,
			Price:     calMaxOrMin(),
			Timestamp: prices[0].Timestamp,
		}, nil
	}, func(data []interface{}) error {
		name := "highest"
		if f.isLowest {
			name = "lowest"
		}
		if err := models.InsertHighest(data, name); err != nil {
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
