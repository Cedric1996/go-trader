/*
 * @Author: cedric.jia
 * @Date: 2021-08-12 11:19:31
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-27 21:54:36
 */

package factor

import (
	"errors"
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
	timestamp int64
}

type highestDatum struct {
	High models.Highest
	Low  models.Highest
}

func NewHighestFactor(name string, calDate string, period int64) *highestFactor {
	return &highestFactor{
		name:      name,
		calDate:   calDate,
		period:    period,
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
		// Reversed: true,
		BeginAt: util.ParseDate("2019-03-01").Unix(),
	})
	if err != nil {
		return err
	}
	period := int(f.period)
	calMaxAndMin := func(prices []*models.StockPriceDay) (max, min float64) {
		min = math.Inf(1)
		max = 0.0
		for _, p := range prices {
			// min = math.Min(p.Close, min)
			max = math.Max(p.Close, max)
		}
		return max, min
	}
	highest := []interface{}{}
	// lowest := []interface{}{}
	if len(prices) <= 120 {
		return nil
	}
	for i := 60; i < len(prices)-period; i++ {
		max, _ := calMaxAndMin(prices[i : i+period])
		highest = append(highest, models.Highest{
			Code:      code,
			Price:     max,
			Timestamp: prices[i-period].Timestamp,
		})
		// lowest = append(lowest, models.Highest{
		// 	Code:      code,
		// 	Price:     min,
		// 	Timestamp: prices[i+period-1].Timestamp,
		// })
	}
	if err := models.InsertHighest(highest, "highest_60_120"); err != nil {
		return err
	}
	// if err := models.InsertHighest(lowest, fmt.Sprintf("lowest_120",f.period)); err != nil {
	// 	return err
	// }
	return nil
}

func (f *highestFactor) Clean() error {
	return models.RemoveHighest(f.timestamp, f.period)
}

func (f *highestFactor) execute() error {
	day, err := models.GetTradeDay(true, 1, f.timestamp)
	if err != nil {
		return err
	}
	if len(day) == 0 || day[0].Timestamp != f.timestamp {
		return fmt.Errorf("error: highest factor task date: %s", f.calDate)
	}
	queue, err := queue.NewQueue(f.name, f.calDate, 100, 200, func(data interface{}) (interface{}, error) {
		code := data.(string)
		prices, err := models.FindHighest(models.SearchOption{
			Code:      code,
			Limit:     f.period,
			Timestamp: f.timestamp,
		})
		if err != nil || len(prices) < int(f.period) {
			return nil, errors.New("")
		}

		calMaxAndMin := func() (float64, float64) {
			min := math.Inf(1)
			max := 0.0
			for _, p := range prices {
				min = math.Min(p.Close, min)
				max = math.Max(p.Close, max)
			}
			return max, min
		}
		max, min := calMaxAndMin()
		return highestDatum{
			High: models.Highest{
				Code:      prices[0].Code,
				Price:     max,
				Timestamp: prices[0].Timestamp},
			Low: models.Highest{
				Code:      prices[0].Code,
				Price:     min,
				Timestamp: prices[0].Timestamp},
		}, nil
	}, func(data []interface{}) error {
		highs := make([]interface{}, len(data))
		lows := make([]interface{}, len(data))
		for i, d := range data {
			datum := d.(highestDatum)
			highs[i] = datum.High
			lows[i] = datum.Low
		}
		if err := models.InsertHighest(highs, fmt.Sprintf("highest_%d",f.period)); err != nil {
			return err
		}
		if err := models.InsertHighest(lows, fmt.Sprintf("lowest_%d",f.period)); err != nil {
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

func (f *highestFactor) initByCode() error {
	queue, _ := queue.NewQueue("init highest data by code", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		code := data.(string)
		highs, err := models.GetHighest(code, "120", f.timestamp, 0)
		if err != nil {
			return nil, err
		}
		lows, err := models.GetLowest(code, "120", f.timestamp, 0)
		if err != nil {
			return nil, err
		}
		if err != nil || len(highs) != len(lows) || len(highs) == 0 || len(lows) == 0 {
			if err := models.RemoveHighestByCode(code); err != nil {
				return nil, err
			}
			if err := f.Init(code); err != nil {
				return nil, err
			}
			return code, err
		}
		return nil, nil
	}, func(data []interface{}) error {
		fmt.Printf("init highest by code: %d\n", len(data))
		return nil
	})
	for code, _ := range service.SecuritySet {
		queue.Push(code)
	}
	queue.Close()
	return nil
}
