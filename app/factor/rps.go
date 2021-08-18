/*
 * @Author: cedric.jia
 * @Date: 2021-08-05 14:10:14
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-18 20:20:57
 */

package factor

import (
	"fmt"
	"math"
	"sync"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
)

type rpsFactor struct {
	name     string
	period   int
	percent  int
	calDate  string
	priceMap map[string]rpsPrice
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
		name:     name,
		period:   period,
		percent:  percent,
		calDate:  calDate,
		priceMap: make(map[string]rpsPrice),
	}
}

func (r *rpsFactor) Run() error {
	if err := r.get(); err != nil {
		return err
	}
	if err := r.run(); err != nil {
		return err
	}
	if err := r.calculate(); err != nil {
		return err
	}
	return nil
}

func (f *rpsFactor) get() error {
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

	priceMapMutex := sync.RWMutex{}
	periodSync := sync.WaitGroup{}
	closePriceChan := make(chan closePriceDatum, 10)
	for key, period := range periods {
		periodSync.Add(1)
		go func(key string, period int64) error {
			datas, err := models.GetStockPriceList(models.SearchOption{
				Timestamp: period,
			})
			if err != nil {
				return err
			}
			for _, datum := range datas {
				priceMapMutex.RLock()
				_, ok := f.priceMap[datum.Code]
				priceMapMutex.RUnlock()
				if !ok {
					priceMapMutex.Lock()
					f.priceMap[datum.Code] = rpsPrice{key: datum.Close}
					priceMapMutex.Unlock()
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
			priceMapMutex.RLock()
			f.priceMap[datum.code][datum.period] = datum.price
			priceMapMutex.RUnlock()
		}
	}()
	periodSync.Wait()
	close(closePriceChan)
	return nil
}

func (f *rpsFactor) run() error {
	if f.priceMap == nil {
		return fmt.Errorf("rps factor priceMap is nil, please check")
	}
	queue, err := queue.NewQueue("rps_increase", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
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
		if err := models.InsertRps(datas, "rps_increase"); err != nil {
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
func (f *rpsFactor) calculate() error {
	p := []int64{5, 10, 20, 120}
	timestamp := util.ParseDate(f.calDate).Unix()
	rpsMap := make(map[string]*models.Rps)
	for code, _ := range service.SecuritySet {
		rpsMap[code] = &models.Rps{
			RpsBase: models.RpsBase{Code: code, Date: f.calDate, Timestamp: timestamp},
		}
	}
	rpsMapMutex := sync.RWMutex{}
	rpsSync := sync.WaitGroup{}
	for _, val := range p {
		rpsSync.Add(1)
		go func(val int64) error {
			rpsIncreaseDatas, err := models.GetRpsIncrease(models.SearchOption{
				Timestamp: timestamp,
				SortBy:    fmt.Sprintf("increase_%d", val),
			})
			if err != nil {
				return err
			}
			total := len(rpsIncreaseDatas)
			end := total * 16 / 100
			for i := 0; i < end; i++ {
				datum := rpsIncreaseDatas[i]
				rpsMapMutex.Lock()
				rps := rpsMap[datum.RpsBase.Code]
				switch val {
				case 5:
					rps.Rps_5 = int64((total - i) * 100 / total)
				case 10:
					rps.Rps_10 = int64((total - i) * 100 / total)
				case 20:
					rps.Rps_20 = int64((total - i) * 100 / total)
				case 120:
					rps.Rps_120 = int64((total - i) * 100 / total)
				}
				rpsMapMutex.Unlock()
			}
			rpsSync.Done()
			return nil
		}(val)
	}
	rpsSync.Wait()
	rpsToInsert := make([]interface{}, 0)
	for _, val := range rpsMap {
		if val.Rps_120 != 0 || val.Rps_20 != 0 || val.Rps_10 != 0 || val.Rps_5 != 0 {
			rpsToInsert = append(rpsToInsert, val)
		}
	}

	if err := models.InsertRps(rpsToInsert, "rps"); err != nil {
		return err
	}
	return nil
}
