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
	queue, _ := queue.NewQueue("init true range data by code", f.calDate, 100, 1000, func(data interface{}) (interface{}, error) {
		code := data.(string)
		prices, err := models.GetStockPriceList(models.SearchOption{
			Code:  code,
			EndAt: f.timestamp,
			Limit: 2,
		})
		if err != nil {
			return nil, err
		}
		if len(prices) != 2 {
			return nil, fmt.Errorf("prices count is not valid")
		}
		trueRanges, err := models.GetTruesRange(models.SearchOption{
			Code:  code,
			EndAt: f.timestamp,
			Limit: 12,
		})
		if err != nil {
			return nil, err
		}
		if len(trueRanges) != 12 {
			return nil, fmt.Errorf("trueRanges count is not valid")
		}
		totalTR := 0.0
		max := prices[0].High - prices[0].Low
		max = math.Max(max, math.Abs(prices[0].High-prices[1].Close))
		max = math.Max(max, math.Abs(prices[0].Low-prices[1].Close))
		totalTR += max
		for _, v := range trueRanges {
			totalTR += v.TR
		}
		return models.TrueRange{
			Code:      code,
			Date:      f.calDate,
			Timestamp: f.timestamp,
			TR:        max,
			ATR:       totalTR / float64(f.period),
		}, nil
	}, func(datas []interface{}) error {
		if err := models.InsertTrueRange(datas); err != nil {
			return err
		}
		return nil
	})
	for code, _ := range service.SecuritySet {
		queue.Push(code)
	}
	queue.Close()
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
		for i := 1; i < len(prices); i++ {
			max := prices[i].High - prices[i].Low
			max = math.Max(max, math.Abs(prices[i].High-prices[i-1].Close))
			max = math.Max(max, math.Abs(prices[i].Low-prices[i-1].Close))
			totalTR += max
			datum := models.TrueRange{
				Code:      prices[i].Code,
				Date:      prices[i].Day,
				Timestamp: prices[i].Timestamp,
				TR:        max,
			}
			if i >= f.period {
				datum.ATR = totalTR / float64(f.period)
				totalTR -= trueRange[i-f.period+1].(models.TrueRange).TR
			}
			trueRange[i-1] = datum
		}
		if err := models.InsertTrueRange(trueRange); err != nil {
			return err
		}
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
