/*
 * @Author: cedric.jia
 * @Date: 2021-08-12 11:19:31
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-23 11:00:21
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
			max = math.Max(p.Close, max)
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
	if err := models.InsertHighest(highest, fmt.Sprintf("highest_%d",f.period)); err != nil {
		return err
	}
	if err := models.InsertHighest(lowest, fmt.Sprintf("lowest_%d",f.period)); err != nil {
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
		if err := models.InsertHighest(highs, "highest"); err != nil {
			return err
		}
		if err := models.InsertHighest(lows, "lowest"); err != nil {
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
		highs, err := models.GetHighest(code, 120, f.timestamp, 0)
		if err != nil {
			return nil, err
		}
		lows, err := models.GetLowest(code, f.timestamp, 0)
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
