/*
 * @Author: cedric.jia
 * @Date: 2021-09-11 15:59:20
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-12 21:08:57
 */

package factor

import (
	"errors"

	chart "github.cedric1996.com/go-trader/app/charts"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const MA_RANGE = 49

type maFactor struct {
	calDate   string
	timestamp int64
	count     int
}

type maIndex struct {
	data string
	index float64
}

func NewMaFactor(date string, count int) *maFactor {
	return &maFactor{
		calDate:   date,
		count:     count,
		timestamp: util.ParseDate(date).Unix(),
	}
}

func (f *maFactor) Run() error {
	return f.run()
}

func (f *maFactor) Clean() error {
	return models.RemoveMa(f.timestamp)
}

func (f *maFactor) run() error {
	queue, _ := queue.NewQueue("moving average", f.calDate, 100, 10, func(data interface{}) (interface{}, error) {
		code := data.(string)
		prices, err := models.GetStockPriceList(models.SearchOption{
			Code:  code,
			BeginAt: f.timestamp,
			Reversed: true,
		})
		if err != nil || len(prices) <  MA_RANGE {
			return nil, errors.New("")
		}
		count := len(prices) - MA_RANGE
		total_50,ma_50 := 0.0, 0.0
		for i:=0;i<MA_RANGE;i++ {
			total_50 += prices[i].Close
		}
		results := make([]interface{}, count)
		for i := 0; i < count; i++ {
			price := prices[i+MA_RANGE]
			total_50 += price.Close
			ma_50 = total_50/ 50.0
			results[i] =  models.Ma{
				Code:      price.Code,
				Timestamp: price.Timestamp,
				MA_50:   ma_50,  
				LongTrend: price.Close > ma_50,
			}
			total_50 -= prices[i].Close
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
		if err := models.InsertMa(datas); err != nil {
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

func (f *maFactor) Output() error {
	chart := chart.NewChart("ma_50 index")
	chart.BarPage(f.maIndex())
	return nil
}

func (f *maFactor) maIndex() *charts.Line {
	// dates ,_ := models.GetTradeDays(models.SearchOption{
	// 	BeginAt: util.ParseDate("2019-03-19").Unix(),
	// 	Reversed: true,
	// })
	xAxis := []interface{}{}
	index :=  []opts.LineData{}

	start := util.ParseDate("2019-04-06")
	for i := 0; i < 29; i++ {
		dates ,_ := models.GetTradeDays(models.SearchOption{
			BeginAt: start.Unix(),
			Reversed: true,
			Limit: 1,
		})
		for _,date := range dates {
			mas, err := models.GetMa(models.SearchOption{
				Timestamp: date.Timestamp,
			})
			if err != nil || len(mas) == 0{
				continue
			}
			count:= 0
			for _,ma := range mas {
				if ma.LongTrend {
					count++
				}
			}
			xAxis = append(xAxis,date.Date)
			index = append(index, opts.LineData{Value:count*100/len(mas)})
		}
		start = start.AddDate(0, 1, 0)
	}
	
	// for _,date := range dates {
	// 	mas, err := models.GetMa(models.SearchOption{
	// 		Timestamp: date.Timestamp,
	// 	})
	// 	if err != nil || len(mas) == 0{
	// 		continue
	// 	}
	// 	count:= 0
	// 	for _,ma := range mas {
	// 		if ma.LongTrend {
	// 			count++
	// 		}
	// 	}
	// 	xAxis = append(xAxis,date.Date)
	// 	index = append(index, opts.LineData{Value:count*100/len(mas)})
	// }
	line:= chart.LineChart(xAxis, index)
	return line
}