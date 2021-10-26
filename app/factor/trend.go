/*
 * @Author: cedric.jia
 * @Date: 2021-08-13 15:35:18
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-28 17:36:33
 */

package factor

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
)

var (
	marketCapMap map[string]float64
	mutex        sync.Mutex
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

type codeDatum struct {
	timestamp int64
	rps       int64
}

func NewTrendFactor(calDate string, period int64, highest_ratio, vcp_ratio, volume float64) *TrendFactor {
	marketCapMap = make(map[string]float64)
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
	if err := f.executeContinue(); err != nil {
		return err
	}
	return nil
}

func (f *TrendFactor) RunByCode(code string) error {
	if err := f.executeByCode(code); err != nil {
		return err
	}
	return nil
}

func (f *TrendFactor) Clean() error {
	return models.RemoveVcpNew(f.timestamp)
}

func (f *TrendFactor) execute() error {
	rps, err := models.GetRps(f.timestamp, 120)
	if err != nil || rps == nil {
		return err
	}
	queue, err := queue.NewQueue("trend", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(trendDatum)
		code := datum.code
		priceDay, err := models.GetStockPriceList(models.SearchOption{Code: code, EndAt: f.timestamp, Limit: 2})
		if err != nil || len(priceDay) != 2 {
			return nil, errors.New("")
		}
		// isApproached, ratio, err := priceDay[0].CheckApproachHighest(code, f.period, f.highest_ratio)
		// if err != nil || !isApproached {
		// 	return nil, errors.New("")
		// }
		isBreak, _, err := priceDay[0].CheckBreakHighest(code, "120", f.timestamp)
		if err != nil || !isBreak {
			return nil, errors.New("")
		}

		tradeDays, _ := models.GetTradeDays(models.SearchOption{EndAt: f.timestamp - 1, Limit: 120})
		if len(tradeDays) < 120 {
			return nil, errors.New("")
		}
		vcp, err := models.GetVcpRanges(code, tradeDays[0].Timestamp, tradeDays[30].Timestamp)
		if err != nil || vcp < f.vcp_ratio {
			return nil, errors.New("")
		}
		vcp_2, err := models.GetVcpRanges(code, tradeDays[40].Timestamp, tradeDays[119].Timestamp)
		if err != nil || vcp_2 > vcp || vcp_2 < f.vcp_ratio {
			return nil, errors.New("")
		}
		// vcp, err := models.GetVcpRanges(code, tradeDays[0].Timestamp, tradeDays[59].Timestamp)
		// if err != nil || vcp < f.vcp_ratio {
		// 	return nil, errors.New("")
		// }
		if err := f.valuationFilter(code, 80); err != nil {
			return nil, errors.New("")
		}
		if err := f.volumeFilter(priceDay); err != nil {
			return nil, errors.New("")
		}
		return models.Vcp{
			RpsBase: models.RpsBase{
				Code:      code,
				Timestamp: f.timestamp,
				Date:      f.calDate,
			},
			Period:       f.period,
			HighestRatio: f.highest_ratio,
			// VcpRatio:     ratio,
			Rps_120: datum.rps,
			// DealPrice: 	  high,
		}, nil
	}, func(data []interface{}) error {
		if err := models.InsertVcpNew(data); err != nil {
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

func (f *TrendFactor) executeByCode(code string) error {
	tradeDays, _ := models.GetTradeDay(true, 60, f.timestamp)
	rps, err := models.GetRpsByOpt(models.SearchOption{
		EndAt:   tradeDays[0].Timestamp,
		BeginAt: tradeDays[59].Timestamp,
		Code:    code,
	})
	if err != nil || rps == nil {
		return err
	}
	queue, err := queue.NewQueue("trend", f.calDate, 1, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(codeDatum)
		t := datum.timestamp
		priceDay, err := models.GetStockPriceList(models.SearchOption{Code: code, EndAt: t, Limit: 2})
		if err != nil || len(priceDay) != 2 {
			return nil, errors.New("")
		}
		isBreak, _, err := priceDay[0].CheckBreakHighest(code, "120", t)
		if err != nil || !isBreak {
			return nil, errors.New("")
		}

		tradeDays, _ := models.GetTradeDays(models.SearchOption{EndAt: t - 1, Limit: 120})
		if len(tradeDays) < 120 {
			return nil, errors.New("")
		}
		vcp, err := models.GetVcpRanges(code, tradeDays[0].Timestamp, tradeDays[30].Timestamp)
		if err != nil || vcp < f.vcp_ratio {
			return nil, errors.New("")
		}
		vcp_2, err := models.GetVcpRanges(code, tradeDays[40].Timestamp, tradeDays[119].Timestamp)
		if err != nil || vcp_2 > vcp || vcp_2 < f.vcp_ratio {
			return nil, errors.New("")
		}
		if err := f.valuationFilter(code, 80); err != nil {
			return nil, errors.New("")
		}
		if err := f.volumeFilter(priceDay); err != nil {
			return nil, errors.New("")
		}
		return models.Vcp{
			RpsBase: models.RpsBase{
				Code:      code,
				Timestamp: t,
				Date:      util.ToDate(t),
			},
			Period:       f.period,
			HighestRatio: f.highest_ratio,
			Rps_120:      datum.rps,
		}, nil
	}, func(data []interface{}) error {
		minVcp := models.Vcp{
			RpsBase: models.RpsBase{
				Timestamp: f.timestamp,
			},
		}
		for _, v := range data {
			vcp := v.(models.Vcp)
			if vcp.RpsBase.Timestamp < minVcp.RpsBase.Timestamp {
				minVcp = vcp
			}
		}
		if err := models.InsertVcpNew([]interface{}{minVcp}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, data := range rps {
		if data.Rps_120 > 0 {
			queue.Push(codeDatum{
				timestamp: data.RpsBase.Timestamp,
				rps:       data.Rps_120,
			})
		}
	}
	queue.Close()
	return nil
}

func (f *TrendFactor) executeContinue() error {
	rps, err := models.GetRpsByOpt(models.SearchOption{EndAt: f.timestamp})
	if err != nil || rps == nil {
		return err
	}
	queue, err := queue.NewQueue("trend", f.calDate, 50, 100, func(data interface{}) (interface{}, error) {
		datum := data.(highestRpsDatum)
		code := datum.code
		priceDay, err := models.GetStockPriceList(models.SearchOption{Code: code, EndAt: datum.timestamp, Limit: 10})
		if err != nil || priceDay == nil {
			return nil, errors.New("")
		}
		isApproached, high, err := priceDay[0].CheckApproachHighest(code, f.period, f.highest_ratio)
		if err != nil || !isApproached {
			return nil, errors.New("")
		}
		max := priceDay[0].Close
		for _, price := range priceDay {
			max = math.Max(max, price.Close)
		}
		if priceDay[0].Close/max < 0.95 || priceDay[0].Close == max || high <= max {
			return nil, errors.New("")
		}
		if err := f.valuationFilter(code, 50.0); err != nil {
			return nil, err
		}
		// lowest, err := models.GetHighestList(models.SearchOption{Code: code, EndAt: datum.timestamp - 24*3600, Limit: 1}, "lowest_120")
		// if err != nil || lowest == nil {
		// 	return nil, errors.New("")
		// }
		return models.HighestApproach{
			RpsBase: models.RpsBase{
				Code:      code,
				Timestamp: datum.timestamp,
				Date:      datum.calDate,
			},
			Highest: high,
			// Lowest:  lowest[0].Price,
			// Range:   (high - lowest[0].Price) / lowest[0].Price,
		}, nil
	}, func(data []interface{}) error {
		if err := models.InsertHighestApproach(data); err != nil {
			return err
		}
		return nil
	})
	for _, data := range rps {
		if data.Rps_250 >= 90 || data.Rps_120 >= 90 || data.Rps_60 >= 90 {
			queue.Push(highestRpsDatum{
				code:      data.RpsBase.Code,
				calDate:   data.RpsBase.Date,
				timestamp: data.RpsBase.Timestamp,
			})
		}
	}
	queue.Close()
	return nil
}

func (f *TrendFactor) valuationFilter(code string, marketCap float64) error {
	var val float64
	market, ok := marketCapMap[code]
	if ok {
		mutex.Lock()
		val = market
		mutex.Unlock()
	} else {
		datas, err := service.InitFundamental(code, f.calDate, 1)
		if err != nil {
			return err
		}
		if len(datas) != 1 {
			return fmt.Errorf("fetch valuation error, code: %v", code)
		}
		mutex.Lock()
		val = datas[0].(models.Valuation).MarketCap
		marketCapMap[code] = val
		mutex.Unlock()
	}
	if math.Dim(val, marketCap) == 0 {
		return fmt.Errorf("marketCap is less than %v, code: %v", marketCap, code)
	}
	return nil
}

func (f *TrendFactor) volumeFilter(prices []*models.StockPriceDay) error {
	if volume := prices[0].GetVolume(); volume < f.volume {
		return errors.New("")
	}
	vol := prices[0].Volume - prices[1].Volume
	close := prices[0].Close - prices[1].Close
	if float64(vol)*close < 0 {
		return errors.New("")
	}
	return nil
}
