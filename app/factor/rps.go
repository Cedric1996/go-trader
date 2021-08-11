/*
 * @Author: cedric.jia
 * @Date: 2021-08-05 14:10:14
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-06 15:53:38
 */

package factor

import (
	"fmt"
	"math"
	"sync"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/util"
)

type rpsFactor struct {
	name          string
	period        int
	percent       int
	calDate       string
	priceMap      map[string]rpsPrice
	priceMapMutex sync.RWMutex
}

type rpsDatum struct {
	rpsBase  models.RpsBase
	rpsPrice rpsPrice
}

type rpsPrice map[string]float64
type closePriceDatum struct {
	code   string
	period string
	price  float64
}

func NewRpsFactor(name string, period int, percent int, calDate string) *rpsFactor {
	return &rpsFactor{
		name:          name,
		period:        period,
		percent:       percent,
		calDate:       calDate,
		priceMap:      make(map[string]rpsPrice),
		priceMapMutex: sync.RWMutex{},
	}
}

func (f *rpsFactor) Get() error {
	var max int64
	max = 120
	p := []int64{5, 10, 20, max}
	periods := make(map[string]int64)
	timestamp := util.ParseDate(f.calDate).Unix()
	periods["period_0"] = timestamp

	timestamps, err := models.GetTradeDayByPeriod(max+1, timestamp)
	if err != nil {
		return err
	}
	for _, val := range p {
		periods[fmt.Sprintf("period_%d", val)] = timestamps[val]
	}

	periodSync := sync.WaitGroup{}
	closePriceChan := make(chan closePriceDatum, 10)
	for key, period := range periods {
		periodSync.Add(1)
		go func(key string, period int64) error {
			datas, err := models.GetStockPriceList(models.SearchPriceOption{
				Timestamp: period,
			})
			if err != nil {
				return err
			}
			for _, datum := range datas {
				f.priceMapMutex.RLock()
				_, ok := f.priceMap[datum.Code]
				f.priceMapMutex.RUnlock()
				if !ok {
					f.priceMapMutex.Lock()
					f.priceMap[datum.Code] = rpsPrice{key: datum.Close}
					f.priceMapMutex.Unlock()
				} else {
					closePriceChan <- closePriceDatum{
						code:   datum.Code,
						period: key,
						price:  datum.Close,
					}
				}
			}
			periodSync.Done()
			return nil
		}(key, period)
	}
	go func() {
		for datum := range closePriceChan {
			f.priceMapMutex.RLock()
			f.priceMap[datum.code][datum.period] = datum.price
			f.priceMapMutex.RUnlock()
		}
	}()
	periodSync.Wait()
	close(closePriceChan)
	return nil
}

func (f *rpsFactor) Run() error {
	if f.priceMap == nil {
		return fmt.Errorf("rps factor priceMap is nil, please check")
	}
	queue, err := queue.NewQueue("rps_increase", 50, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(rpsDatum)
		val := datum.rpsPrice
		rpsIncrease := &models.RpsIncrease{
			RpsBase: datum.rpsBase,
		}
		if math.Dim(val["period_120"], 1.0) < 0.0000001 {
			rpsIncrease.Increase_120 = -1
		} else {
			rpsIncrease.Increase_120 = (val["period_0"] - val["period_120"]) / val["period_120"]

		}
		if math.Dim(val["period_20"], 1.0) < 0.0000001 {
			rpsIncrease.Increase_20 = -1
		} else {
			rpsIncrease.Increase_20 = (val["period_0"] - val["period_20"]) / val["period_20"]

		}
		if math.Dim(val["period_10"], 1.0) < 0.0000001 {
			rpsIncrease.Increase_10 = -1
		} else {
			rpsIncrease.Increase_10 = (val["period_0"] - val["period_10"]) / val["period_10"]

		}
		if math.Dim(val["period_5"], 1.0) < 0.0000001 {
			rpsIncrease.Increase_5 = -1
		} else {
			rpsIncrease.Increase_5 = (val["period_0"] - val["period_5"]) / val["period_5"]
		}
		return rpsIncrease, nil
	}, func(datas []interface{}) error {
		if err := models.InsertRpsIncrease(datas); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	timestamp := util.ParseDate(f.calDate).Unix()
	for k, val := range f.priceMap {
		if val != nil {
			queue.Push(rpsDatum{
				rpsBase:  models.RpsBase{Code: k, Timestamp: timestamp, Date: f.calDate},
				rpsPrice: val,
			})
		}
	}
	queue.Close()
	return nil
}

/**
 * Rps can be specified by period and trade_date
 */
func calculate() error {
	return nil
}
