/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 17:02:05
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-24 17:49:07
 */

package strategy

import (
	"errors"
	"fmt"
	"math"
	"sort"

	chart "github.cedric1996.com/go-trader/app/charts"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/util"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type highestRps struct {
	Name     string
	Date     string
	Net      float64
	DrawBack float64
	NetAvg   float64
	NetCount float64
	dates  []interface{}
	index  []opts.LineData
}

func NewHighestRpsStrategy(name, date string) *highestRps {
	if len(name) == 0 {
		name = "highest_rps_strategy"
	}
	return &highestRps{
		Name:     name,
		Net:      1.0,
		Date:     date,
		DrawBack: 1.0,
		dates: make([]interface{},0),
		index: make([]opts.LineData, 0),
	}
}

func (v *highestRps) Run() error {
	queue, err := queue.NewQueue("highest with rps", "", 100, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(*models.HighestRps)
		unit, _ := v.highestRpsSignal(TradeSignal{
			Code:      datum.RpsBase.Code,
			StartUnix: datum.RpsBase.Timestamp,
		})
		if unit == nil {
			return nil, errors.New("error")
		}
		return unit, nil
	}, func(datas []interface{}) error {
		if err := models.InsertMany(datas, v.Name); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	opt := models.SearchOption{
		Limit: 1000,
		Skip:  0,
	}
	for {
		datas, err := models.GetHighestRps(opt)
		if err != nil {
			break
		}
		for _, data := range datas {
			queue.Push(data)
		}
		if len(datas) != 1000 {
			break
		}
		opt.Skip += 1000
	}
	queue.Close()
	return nil
}

func (v *highestRps) highestRpsSignal(sig TradeSignal) (unit *TradeUnit, err error) {
	opt := models.SearchOption{
		Code:     sig.Code,
		BeginAt:  sig.StartUnix,
		Reversed: true,
	}
	unit = &TradeUnit{
		Code:  sig.Code,
		Start: sig.StartUnix,
	}
	rps, err := models.GetRpsByOpt(opt)
	if err != nil || len(rps) == 0 {
		return nil, err
	}

	if rps[0].Rps_20 > 90 {
		return nil, err
	}

	prices, err := models.GetStockPriceList(opt)
	if err != nil {
		return nil, err
	}
	// emas, err := models.GetEma(opt)
	// if err != nil {
	// 	return unit, err
	// }
	days := len(prices)
	if len(prices) > len(rps) {
		days = len(rps)
	}

	var preClose, sellPrice, dealPrice float64

	maxClose := 0.0
	isDeal := false
	for i := 1; i < days; i++ {
		preClose = prices[i-1].Close
		maxClose = math.Max(preClose, maxClose)
		unit.Period = int64(i)
		unit.End = prices[i].Timestamp
		if !isDeal {
			// ratio := prices[i-1].Close / prices[i-1].HighLimit
			// if ratio > 0.91 && ratio < 0.94{
			if prices[i-1].Close != prices[i-1].HighLimit {
				dealPrice = prices[i-1].Close
				isDeal = true
			} else {
				return nil, errors.New("nil")
			}
		}
		// if rps[i].Rps_5 < 90  && rps[i].Rps_10 < 90{
		// 	sellPrice = prices[i].Close
		// 	break
		// }

		// if rps[i].Rps_5 > 98 {
		// 	sellPrice = prices[i].Close
		// 	break
		// }
		
		if prices[i].Open/dealPrice < 0.94 {
			sellPrice = prices[i].Open
			break
		} else if prices[i].Low/dealPrice < 0.94 {
			sellPrice = dealPrice * 0.94
			break
		}
		// }  else if prices[i].Close / maxClose < 0.92 {
		// 	sellPrice = prices[i].Close
		// 	break
		// }
		// if prices[i].Low < (preClose - LossCo*atr) {
		// 	sellPrice = preClose - LossCo*atr
		// 	break
		// }
		if rps[i].Rps_5 + rps[i].Rps_10  == 0 {
			sellPrice = prices[i].Close
			break
		}
		// if rps[i-1].Rps_5 > rps[i].Rps_5 &&  rps[i].Rps_10 >=  rps[i-1].Rps_10 {
		sum :=rps[i-1].Rps_5 + rps[i-1].Rps_10
		if  sum >= 196 &&  rps[i].Rps_5 + rps[i].Rps_10 < sum {
			sellPrice = prices[i].Close
			break
		}

		sellPrice = prices[i].Close
	}
	if days < 2 || sellPrice == 0 {
		return nil, fmt.Errorf("trade period is too short")
	}
	unit.Max = (maxClose - dealPrice) / dealPrice
	unit.Net = (sellPrice - dealPrice) / dealPrice
	return unit, nil
}

// func (v *highestRps) Chart() *charts.Line{
// 	return chart.LineChart(v.dates, v.index)
// }

type Pos struct {
	Code      string  `bson:"code"`
	Name      string  `bson:"name"`
	DealPrice float64 `bson:"dealPrice"`
	SellPrice float64 `bson:"sellPrice"`
	LossPrice float64 `bson:"lossPrice"`
	RPS_5 int64 `bson:"rps5"`
	RPS_10 int64 `bson:"rps10"`
	RPS_20 int64 `bson:"rps20"`
}

func (v *highestRps) Pos() ([]*Pos, error) {
	t := util.ParseDate(v.Date).Unix()
	pos := make([]*Pos, 0)
	opt := models.SearchOption{
		Timestamp: t,
	}
	datas, err := models.GetHighestRps(opt)
	if err != nil {
		return nil, err
	}
	for _, data := range datas {
		opt.Code = data.RpsBase.Code
		rps, err := models.GetRpsByOpt(opt)
		if err != nil || len(rps) == 0 {
			continue
		}
		prices, err := models.GetStockPriceList(opt)
		if err != nil {
			continue
		}
		
		var sellPrice, dealPrice, lossPrice float64
		if prices[0].Close != prices[0].HighLimit {
			dealPrice = prices[0].Close
		}
		if dealPrice < 0.1 {
			continue
		}
		lossPrice = dealPrice * 0.94
		info, _ := models.GetSecurityByCode(opt.Code)
		pos = append(pos, &Pos{
			Code:      opt.Code,
			Name:      info.DisplayName,
			DealPrice: dealPrice,
			SellPrice: sellPrice,
			LossPrice: lossPrice,
			RPS_5: rps[0].Rps_5,
			RPS_10: rps[0].Rps_10,
			RPS_20: rps[0].Rps_20,

		})
	}
	return pos, nil
}

func HighestRpsPos(code string, t1,t2 int64) (*Pos, error) {
	opt := models.SearchOption{
		BeginAt: t1,
		EndAt:t2,
		Code: code,
		Reversed: true,
	}
	prices, err := models.GetStockPriceList(opt)
	if err != nil {
		return nil, err
	}
	trs, err := models.GetTruesRange(opt)
	if err != nil {
		return nil, err
	}
	var preClose, sellPrice, lossPrice float64
	for _, price := range prices {
		preClose = price.Close
		lossPrice = math.Max(lossPrice, preClose*0.92)
		sellPrice = preClose + 2*trs[0].ATR
	}
	// lossPrice = dealPrice * 0.94
	info, _ := models.GetSecurityByCode(opt.Code)
	pos :=  &Pos{
		Code:      opt.Code,
		Name:      info.DisplayName,
		SellPrice: sellPrice,
		LossPrice: lossPrice,
	}
	return pos, nil
}

type winRate struct {
	t int64
	ratio int
	index int
}

func (v *highestRps) WinRateByDate(start string, period int64) error{
	winRates := []winRate{}
	taskQueue := queue.NewTaskQueue("highest test winRate statistic", 10, func(data interface{}) error {
		t := data.(int64)
		results, _ := models.GetTradeResultByDay(t, v.Name)
		nh_nl, _ := models.GetHighLowIndex(models.SearchOption{
			Timestamp: t,
		})
		if len(results) == 0 {
			winRates = append(winRates, winRate{
				t : t,
				ratio: 50,
				index: nh_nl[0].Index -200,
			})
			return nil
		}
		netCount := 0
		net := 0.0
		endTimestamp := t + period * 3600 * 24
		for _, r := range results {
			if r.Period < 10 {
				net = r.Net
			} else {
				prices, _ := models.GetStockPriceList(models.SearchOption{
					Code:     r.Code,
					BeginAt:  t,
					EndAt:    endTimestamp,
					Reversed: true,
				})
				len := len(prices)
				net = (prices[len-1].Close - prices[0].Close) / prices[len-1].Close
			}
			if net > 0.0 {
				netCount ++
			}
		}
		
		winRates = append(winRates, winRate{
			t : t,
			ratio: netCount*100 / len(results) - 50,
			index: nh_nl[0].Index -200,
		})
		return nil
	}, func(dateChan *chan interface{}) {
		dates, _ := models.GetTradeDays(models.SearchOption{
			BeginAt: util.ParseDate(start).Unix(),
			Reversed: true,
			Limit: 50,
		})
		for _, date := range dates {
			*dateChan <- date.Timestamp
		}
	})
	if err := taskQueue.Run(); err != nil {
		return nil
	}
	sort.Slice(winRates, func(i, j int) bool {
		return winRates[i].t < winRates[j].t
	})
	indexes := []opts.LineData{}
	for _, winRate := range winRates {
		v.dates = append(v.dates, util.ToDate(winRate.t))
		v.index = append(v.index, opts.LineData{Value: winRate.ratio})
		indexes = append(indexes, opts.LineData{Value: winRate.index})
	}
	ch1 :=  chart.LineChart(v.dates, v.index,indexes)
	// ch2 :=  chart.LineChart(v.dates, indexes)
	ch := chart.NewChart("highest rps strategy")
	ch.BarPage(ch1)
	return nil
}