/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 16:16:50
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-28 14:25:16
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
	code      string
	calDate   string
	timestamp int64
	rps_250   int64
	rps_120   int64
	rps_60    int64
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
	return models.RemoveHighestRps(f.timestamp)
}

func (f *HighestRpsFactor) execute() error {
	rps, err := models.GetRpsByOpt(models.SearchOption{EndAt: f.timestamp})
	if err != nil || rps == nil {
		return err
	}
	queue, err := queue.NewQueue("highest_rps", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(highestRpsDatum)
		code := datum.code
		timestamp := datum.timestamp
		rpsIncrease, err := models.GetRpsIncrease(models.SearchOption{Code: code, Timestamp: timestamp})
		if err != nil || rpsIncrease == nil {
			return nil, errors.New("")
		}
		highest, err := models.GetHighestList(models.SearchOption{Code: code, BeginAt: timestamp}, "highest_120")
		if err != nil || len(highest) < 120 {
			return nil, errors.New("")
		}
		prices, err := models.GetStockPriceList(models.SearchOption{Code: code, Timestamp: timestamp})
		if err != nil || prices == nil {
			return nil, errors.New("")
		}
		highestPrice := highest[len(highest)-120]
		return models.HighestRps{
			RpsBase: models.RpsBase{
				Code:      code,
				Timestamp: timestamp,
				Date:      datum.calDate,
			},
			Rps_250:         datum.rps_250,
			Rps_120:         datum.rps_120,
			Rps_60:          datum.rps_60,
			RpsIncrease_250: rpsIncrease[0].Increase_250,
			RpsIncrease_120: rpsIncrease[0].Increase_120,
			RpsIncrease_60:  rpsIncrease[0].Increase_60,
			Net:             (highestPrice.Price - prices[0].Price.Close) / prices[0].Price.Close,
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
		if data.Rps_250 >= 90 || data.Rps_120 >= 90 || data.Rps_60 >= 90 {
			queue.Push(highestRpsDatum{
				code:      data.RpsBase.Code,
				calDate:   data.RpsBase.Date,
				timestamp: data.RpsBase.Timestamp,
				rps_250:   data.Rps_250,
				rps_120:   data.Rps_120,
				rps_60:    data.Rps_60,
			})
		}
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
