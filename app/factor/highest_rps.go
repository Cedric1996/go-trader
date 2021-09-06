/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 16:16:50
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-06 16:45:56
 */

package factor

import (
	"errors"
	"fmt"
	"math"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
)

type HighestRpsFactor struct {
	calDate       string  `bson:"calDate, omitempty"`
	timestamp     int64   `bson:"timestamp, omitempty"`
	highest_ratio float64 `bson:"highest_ratio, omitempty"`
	volume        float64 `bson:"volume, omitempty"`
}

type highestRpsDatum struct {
	code   string
	rps_20 int64
	rps_10 int64
	rps_5  int64
}

func NewHighestRpsFactor(calDate string, highest_ratio, volume float64) *HighestRpsFactor {
	return &HighestRpsFactor{
		calDate:       calDate,
		highest_ratio: highest_ratio,
		volume:        volume,
		timestamp:     util.ParseDate(calDate).Unix(),
	}
}

func (f *HighestRpsFactor) Run() error {
	if err := f.execute(); err != nil {
		return err
	}
	return nil
}

func (f *HighestRpsFactor) Clean() error {
	// return models.RemoveVcp(f.timestamp)
	return nil
}

func (f *HighestRpsFactor) execute() error {
	rps, err := models.GetRps(f.timestamp, 120)
	if err != nil || rps == nil {
		return err
	}
	queue, err := queue.NewQueue("highest_rps", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(highestRpsDatum)
		code := datum.code
		priceDay, err := models.GetStockPriceList(models.SearchOption{Code: code, Timestamp: f.timestamp})
		if err != nil || priceDay == nil {
			return nil, errors.New("")
		}
		if volume := priceDay[0].GetVolume(); volume < f.volume {
			return nil, errors.New("")
		}

		isApproached, err := priceDay[0].CheckBreakHighest(code, f.timestamp)
		if err != nil || !isApproached {
			return nil, errors.New("")
		}
		if err := f.valuationFilter(code, 80); err != nil {
			return nil, errors.New("")
		}
		return models.HighestRps{
			RpsBase: models.RpsBase{
				Code:      code,
				Timestamp: f.timestamp,
				Date:      f.calDate,
			},
			Rps_20: datum.rps_20,
			Rps_10: datum.rps_10,
			Rps_5:  datum.rps_5,
		}, nil
	}, func(data []interface{}) error {
		if err := models.InsertHighestRps(data); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, data := range rps {
		queue.Push(highestRpsDatum{
			code:   data.RpsBase.Code,
			rps_20: data.Rps_20,
			rps_10: data.Rps_10,
			rps_5:  data.Rps_5,
		})
	}
	queue.Close()
	return nil
}

func (f *HighestRpsFactor) valuationFilter(code string, marketCap float64) error {
	datas, err := service.InitFundamental(code, f.calDate, 1)
	if err != nil {
		return err
	}
	if len(datas) != 1 {
		return fmt.Errorf("fetch valuation error, code: %v", code)
	}
	val := datas[0].(models.Valuation).MarketCap
	if math.Dim(val, marketCap) < 0.1 {
		return fmt.Errorf("marketCap is less than %v, code: %v", marketCap, code)
	}
	return nil
}
