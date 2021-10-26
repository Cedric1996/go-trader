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

type LowestRpsFactor struct {
	calDate   string `bson:"calDate, omitempty"`
	timestamp int64  `bson:"timestamp, omitempty"`
}

type lowestRpsDatum struct {
	code      string
	calDate   string
	timestamp int64
	rps_250   int64
	rps_120   int64
	rps_60    int64
}

func NewLowestRpsFactor(calDate string) *LowestRpsFactor {
	return &LowestRpsFactor{
		calDate:   calDate,
		timestamp: util.ParseDate(calDate).Unix(),
	}
}

func (f *LowestRpsFactor) Run() error {
	if err := f.execute(); err != nil {
		return err
	}
	return nil
}

// func (f *LowestRpsFactor) Clean() error {
// 	return models.RemoveLowestRps(f.timestamp)
// }

func (f *LowestRpsFactor) execute() error {
	rps, err := models.GetRpsByOpt(models.SearchOption{EndAt: f.timestamp})
	if err != nil || rps == nil {
		return err
	}
	queue, err := queue.NewQueue("highest_rps", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(lowestRpsDatum)
		code := datum.code
		timestamp := datum.timestamp
		highest, err := models.GetHighestList(models.SearchOption{Code: code, BeginAt: timestamp + 1}, "highest_120")
		if err != nil || len(highest) < 120 {
			return nil, errors.New("")
		}
		prices, err := models.GetStockPriceList(models.SearchOption{Code: code, Timestamp: timestamp})
		if err != nil || prices == nil {
			return nil, errors.New("")
		}
		highestPrice := highest[0]
		if len(highest) >= 120 {
			highestPrice = highest[len(highest)-120]
		}
		lowest, err := models.GetHighestList(models.SearchOption{Code: code, Timestamp: highestPrice.Timestamp}, "lowest_120")
		if err != nil || lowest == nil {
			return nil, errors.New("")
		}
		price := prices[0].Price.Close
		return models.LowestRps{
			RpsBase: models.RpsBase{
				Code:      code,
				Timestamp: timestamp,
				Date:      datum.calDate,
			},
			Highest:  highestPrice.Price,
			Lowest:   lowest[0].Price,
			Price:    price,
			Net:      (highestPrice.Price - price) / price,
			DrawBack: (lowest[0].Price - price) / price,
		}, nil
	}, func(data []interface{}) error {
		if err := models.InsertLowestRps(data); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, data := range rps {
		if data.Rps_250 >= 90 || data.Rps_120 >= 90 || data.Rps_60 >= 90 {
			queue.Push(lowestRpsDatum{
				code:      data.RpsBase.Code,
				calDate:   data.RpsBase.Date,
				timestamp: data.RpsBase.Timestamp,
			})
		}
	}
	queue.Close()
	return nil
}

func (f *LowestRpsFactor) valuationFilter(code string, marketCap float64) error {
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
