/*
 * @Author: cedric.jia
 * @Date: 2021-08-22 16:05:14
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-22 16:09:04
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

type trueRangeFactor struct {
	calDate   string
	timestamp int64
	code      string
	period    int
}

func NewTrueRangeFactor(calDate string, period int) *trueRangeFactor {
	return &trueRangeFactor{
		calDate:   calDate,
		period:    period,
		timestamp: util.ParseDate(calDate).Unix(),
	}
}

func (f *trueRangeFactor) Run() error {
	if err := f.execute(); err != nil {
		return err
	}
	return nil
}

func (f *trueRangeFactor) Clean() error {
	return models.RemoveTr(f.timestamp)
}

func (f *trueRangeFactor) execute() error {
	return nil
}

func (f *trueRangeFactor) InitByCode() error {
	taskQueue := queue.NewTaskQueue("init true range data by code", 100, func(data interface{}) error {
		code := data.(string)
		prices, err := models.GetStockPriceList(models.SearchOption{
			Code:     code,
			Reversed: true,
		})
		if err != nil {
			return err
		}
		if len(prices) < 2 {
			return fmt.Errorf("prices count is not valid")
		}
		trueRange := make([]interface{}, len(prices)-1)
		totalTR := 0.0
		for i, _ := range prices {
			max := prices[i].High - prices[i].Low
			max = math.Max(max, math.Abs(prices[i].High-prices[i+1].Close))
			max = math.Max(max, math.Abs(prices[i].Low-prices[i+1].Close))
			totalTR += max
			datum := models.TrueRange{
				Code:      prices[i].Code,
				Date:      prices[i].Day,
				Timestamp: prices[i].Timestamp,
				TR:        max,
			}
			if i >= f.period-1 {
				datum.ATR = totalTR / float64(f.period)
				totalTR -= trueRange[i-f.period+1].(models.TrueRange).TR
			}
			trueRange[i] = datum
		}
		if err := models.InsertTrueRange(trueRange); err != nil {
			return err
		}
		fmt.Printf("init tr count: %v, code: %s\n", len(trueRange), code)
		return nil
	}, func(dateChan *chan interface{}) {
		for code, _ := range service.SecuritySet {
			*dateChan <- code
		}
	})
	if err := taskQueue.Run(); err != nil {
		return err
	}
	return nil
}
