/*
 * @Author: cedric.jia
 * @Date: 2021-08-05 14:10:14
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-06 15:53:38
 */

package factor

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
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

func NewRpsFactor(name string, period int, percent int, calDate string) *rpsFactor {
	return &rpsFactor{
		name:     name,
		period:   period,
		percent:  percent,
		calDate:  calDate,
		priceMap: make(map[string]rpsPrice),
	}
}

func (f *rpsFactor) Get() error {
	// p := []int64{5, 10, 20, 120}
	p := []int64{5}
	periods := make(map[string]int64)
	timestamp := util.ParseDate(f.calDate).Unix()
	periods["period_0"] = timestamp

	for _, val := range p {
		period, err := models.GetTradeDayByPeriod(val, timestamp)
		if err != nil {
			return err
		}
		periods[fmt.Sprintf("period_%d", val)] = period
	}
	for key, period := range periods {
		fmt.Printf("begin query period %s raw data\n", key)
		datas, err := models.GetStockPriceList(models.SearchPriceOption{
			Timestamp: period,
		})
		if err != nil {
			return err
		}
		for _, datum := range datas {
			rps, ok := f.priceMap[datum.Code]
			if !ok {
				f.priceMap[datum.Code] = rpsPrice{
					key: datum.Close,
				}
			} else {
				rps[key] = datum.Close
			}
		}
		fmt.Printf("end query period %s raw data\n", key)
	}
	return nil
}

func (f *rpsFactor) Run() error {
	if f.priceMap == nil {
		return fmt.Errorf("rps factor priceMap is nil, please check")
	}
	queue, err := queue.NewQueue("rps_increase", 50, 100, func(data interface{}) (interface{}, error) {
		datum := data.(rpsDatum)
		val := datum.rpsPrice
		return &models.RpsIncrease{
			RpsBase:      datum.rpsBase,
			Increase_120: (val["period_0"] - val["period_120"]) / val["period_0"],
			Increase_20:  (val["period_0"] - val["period_20"]) / val["period_0"],
			Increase_10:  (val["period_0"] - val["period_10"]) / val["period_0"],
			Increase_5:   (val["period_0"] - val["period_5"]) / val["period_0"],
		}, nil
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
