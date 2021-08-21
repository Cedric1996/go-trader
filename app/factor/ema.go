/*
 * @Author: cedric.jia
 * @Date: 2021-08-17 14:13:23
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-21 15:56:52
 */

package factor

import (
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
)

const MAX_MA_RANGE = 60

type emaFactor struct {
	calDate   string
	timestamp int64
	count     int64
}

func NewEmaFactor(date string, count int64) *emaFactor {
	return &emaFactor{
		calDate:   date,
		count:     count,
		timestamp: util.ParseDate(date).Unix(),
	}
}

func (f *emaFactor) Run() error {
	return f.run()
}

func (f *emaFactor) Clean() error {
	return models.RemoveEma(f.timestamp)
}

func (f *emaFactor) run() error {
	queue, _ := queue.NewQueue("index moving average", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
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
		ema_6 := calEma(closeArr, 6)
		ema_12 := calEma(closeArr, 12)
		ema_26 := calEma(closeArr, 26)
		ema_60 := calEma(closeArr, 60)
		var i int64
		for i = 0; i < f.count; i++ {
			results[i] = models.Ema{
				Code:      prices[i].Code,
				Date:      util.ToDate(prices[i].Timestamp),
				Timestamp: prices[i].Timestamp,
				MA_6:      ema_6[i],
				MA_12:     ema_12[i],
				MA_26:     ema_26[i],
				MA_60:     ema_60[i],
			}
		}
		return results, nil
	}, func(data []interface{}) error {
		datas := make([]interface{}, 0)
		for _, v := range data {
			datas = append(datas, v.([]interface{})...)
		}
		if len(datas) == 0 {
			return nil
		}
		if err := models.InsertEma(datas); err != nil {
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
