/*
 * @Author: cedric.jia
 * @Date: 2021-08-12 11:19:31
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-20 15:04:10
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

func (f *highestFactor) Clean() error {
	return models.RemoveHighest(f.timestamp)
}

func (f *highestFactor) execute() error {
	day, err := models.GetTradeDay(true, 1, f.timestamp)
	if err != nil {
		return err
	}
	if len(day) == 0 || day[0].Timestamp != f.timestamp {
		return fmt.Errorf("error: highest factor task date: %s", f.calDate)
	}
	queue, err := queue.NewQueue("highest", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
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
