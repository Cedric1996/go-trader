/*
 * @Author: cedric.jia
 * @Date: 2021-08-13 15:35:18
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-01 10:58:06
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

type TrendFactor struct {
	calDate       string  `bson:"calDate, omitempty"`
	timestamp     int64   `bson:"timestamp, omitempty"`
	period        int64   `bson:"period, omitempty"`
	highest_ratio float64 `bson:"highest_ratio, omitempty"`
	vcp_ratio     float64 `bson:"vcp_ratio, omitempty"`
	volume        float64 `bson:"volume, omitempty"`
}

type trendDatum struct {
	code string
	rps  int64
}

func NewTrendFactor(calDate string, period int64, highest_ratio, vcp_ratio, volume float64) *TrendFactor {
	return &TrendFactor{
		calDate:       calDate,
		period:        period,
		highest_ratio: highest_ratio,
		vcp_ratio:     vcp_ratio,
		volume:        volume,
		timestamp:     util.ParseDate(calDate).Unix(),
	}
}

func (f *TrendFactor) Run() error {
	if err := f.execute(); err != nil {
		return err
	}
	return nil
}

func (f *TrendFactor) Clean() error {
	return models.RemoveVcp(f.timestamp)
}

func (f *TrendFactor) execute() error {
	rps, err := models.GetRps(f.timestamp, 120)
	if err != nil || rps == nil {
		return err
	}
	queue, err := queue.NewQueue("trend", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(trendDatum)
		code := datum.code
		priceDay, err := models.GetStockPriceList(models.SearchOption{Code: code, Timestamp: f.timestamp})
		if err != nil || priceDay == nil {
			return nil, err
		}
		if volume := priceDay[0].GetVolume(); volume < f.volume {
			return nil, err
		}

		isApproached, err := priceDay[0].CheckApproachHighest(code, f.timestamp, f.highest_ratio)
		if err != nil || !isApproached {
			return nil, err
		}
		vcp, err := models.GetVcpRange(code, f.timestamp, f.period)
		if err != nil || vcp > f.vcp_ratio {
			return nil, err
		}
		if err := f.valuationFilter(code, 80); err != nil {
			return nil, err
		}
		return models.Vcp{
			RpsBase: models.RpsBase{
				Code:      code,
				Timestamp: f.timestamp,
				Date:      f.calDate,
			},
			Period:       f.period,
			HighestRatio: f.highest_ratio,
			VcpRatio:     vcp,
			Rps_120:      datum.rps,
		}, nil
	}, func(data []interface{}) error {
		if err := models.InsertVcp(data); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, data := range rps {
		queue.Push(trendDatum{
			code: data.RpsBase.Code,
			rps:  data.Rps_120,
		})
	}
	queue.Close()
	return nil
}

func (f *TrendFactor) valuationFilter(code string, marketCap float64) error {
	datas, err := service.InitFundamental(code, f.calDate, 1)
	if err != nil {
		return err
	}
	if len(datas) != 1 {
		return fmt.Errorf("fetch valuation error, code: %v", code)
	}
	val := datas[0].(models.Valuation).MarketCap
	if math.Dim(val, marketCap) == 0 {
		return fmt.Errorf("marketCap is less than %v, code: %v", marketCap, code)
	}
	return nil
}
