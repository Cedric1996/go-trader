/*
 * @Author: cedric.jia
 * @Date: 2021-08-17 14:13:23
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-17 18:25:09
 */

package factor

import (
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
)

const MAX_MA_RANGE = 30

type movingAverageFactor struct {
	calDate   string
	timestamp int64
	count     int64
}

func NewMovingAverageFactor(calDate string, count int64) *movingAverageFactor {
	return &movingAverageFactor{
		calDate:   calDate,
		timestamp: util.ParseDate(calDate).Unix(),
		count:     count,
	}
}

func (f *movingAverageFactor) Run() error {
	if err := f.run(); err != nil {
		return err
	}
	return nil
}

func (f *movingAverageFactor) run() error {
	queue, _ := queue.NewQueue("moving average", 50, 1, func(data interface{}) (interface{}, error) {
		code := data.(string)
		prices, err := models.GetStockPriceList(models.SearchOption{
			Code:  code,
			EndAt: f.timestamp,
			Limit: f.count + MAX_MA_RANGE,
		})
		if err != nil || int64(len(prices)) != f.count+MAX_MA_RANGE {
			return nil, nil
		}
		results := make([]interface{}, f.count)
		closeArr := make([]float64, len(prices))
		for i, p := range prices {
			closeArr[i] = p.Close
		}
		ema_5 := calEma(closeArr, 5)
		ema_10 := calEma(closeArr, 10)
		ema_20 := calEma(closeArr, 20)
		ema_30 := calEma(closeArr, 30)
		var i int64
		for i = 0; i < f.count; i++ {
			results[i] = models.MovingAverage{
				Code:      prices[i].Code,
				Date:      util.ToDate(prices[i].Timestamp),
				Timestamp: prices[i].Timestamp,
				MA_5:      ema_5[i],
				MA_10:     ema_10[i],
				MA_20:     ema_20[i],
				MA_30:     ema_30[i],
			}
		}
		return results, nil
	}, func(data []interface{}) error {
		datas := make([]interface{}, 0)
		for _, v := range data {
			datas = append(datas, v.([]interface{})...)
		}
		if err := models.InsertMovingAverage(datas); err != nil {
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

func calEma(arr []float64, period int) []float64 {
	size := len(arr) - period
	data := make([]float64, size)
	for i := 0; i < size; i++ {
		data[i] = ema(arr[i : i+period])
	}
	return data
}

func ema(arr []float64) float64 {
	n := len(arr)
	index := float64(2) / float64(n+1)
	ema := make([]float64, n)
	ema[0] = arr[n-1]
	for i := 1; i < n; i++ {
		ema[i] = arr[n-i-1]*index + (1-index)*ema[i-1]
	}
	return ema[n-1]
}
