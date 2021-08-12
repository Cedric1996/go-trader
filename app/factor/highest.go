/*
 * @Author: cedric.jia
 * @Date: 2021-08-12 11:19:31
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-12 22:19:47
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
	name    string
	calDate string
	period  int64
}

func NewHighestFactor(name string, calDate string, period int64) *highestFactor {
	return &highestFactor{
		name:    name,
		calDate: calDate,
		period:  period,
	}
}

func (f *highestFactor) Run() error {
	if err := f.execute(); err != nil {
		return err
	}
	return nil
}

func (f *highestFactor) execute() error {
	t := util.ParseDate(f.calDate).Unix()
	day, err := models.GetTradeDay(true, 1, t)
	if err != nil {
		return err
	}
	if len(day) == 0 || day[0].Timestamp != t {
		return fmt.Errorf("error: highest factor task date: %s", f.calDate)
	}
	queue, err := queue.NewQueue("highest", 50, 1000, func(data interface{}) (interface{}, error) {
		code := data.(string)
		prices, err := models.FindHighest(models.SearchPriceOption{
			Code:      code,
			Limit:     f.period,
			Timestamp: t,
		})
		if err != nil || len(prices) < int(f.period) {
			return nil, err
		}
		max := 0.0
		for _, p := range prices {
			max = math.Max(p.High, max)
		}
		return models.Highest{
			Code:      prices[0].Code,
			Price:     max,
			Timestamp: prices[0].Timestamp,
		}, nil
	}, func(data []interface{}) error {
		if err := models.InsertHighest(data); err != nil {
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
