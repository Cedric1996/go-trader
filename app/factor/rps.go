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
	"github.cedric1996.com/go-trader/app/util"
)

type rpsFactor struct {
	name     string
	period   int
	percent  int
	calDate  string
	priceMap map[string]rpsPrice
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
	periods := map[string]int64{
		"period_120": util.ParseDate(f.calDate).AddDate(0, 0, -120).Unix(),
		"period_20":  util.ParseDate(f.calDate).AddDate(0, 0, -20).Unix(),
		"period_10":  util.ParseDate(f.calDate).AddDate(0, 0, -10).Unix(),
		"period_5":   util.ParseDate(f.calDate).AddDate(0, 0, -5).Unix(),
		"period_0":   util.ParseDate(f.calDate).Unix(),
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
	return nil
}

/**
 * Rps can be specified by period and trade_date
 */
func calculate() error {
	return nil
}
