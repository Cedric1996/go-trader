/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 09:56:23
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-24 14:47:59
 */

package strategy

import (
	"errors"
	"math"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
)

type vcpEma struct {
	Name     string
	Date     string
	Net      float64
	DrawBack float64
	NetAvg   float64
	NetCount float64
	dates  []interface{}
}

func NewVcpEmaStrategy(name string) *vcpEma {
	if len(name) == 0 {
		name = "vcp_ema_strategy"
	}
	return &vcpEma{
		Name:     name,
		Net:      1.0,
		DrawBack: 1.0,
		dates: make([]interface{},0),
	}
}

func (v *vcpEma) Run() error {
	queue, err := queue.NewQueue("highest with rps", "", 100, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(*models.Vcp)
		unit := v.vcpEmaTradeSignal(TradeSignal{
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
		datas, err := models.GetVcp(opt)
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


func (v *vcpEma) vcpEmaTradeSignal(sig TradeSignal) (unit *TradeUnit) {
	opt := models.SearchOption{
		Code:     sig.Code,
		BeginAt:  sig.StartUnix,
		Reversed: true,
	}
	unit = &TradeUnit{
		Code:  sig.Code,
		Start: sig.StartUnix,
	}
	prices, err := models.GetStockPriceList(opt)
	if err != nil {
		return nil
	}
	// emas, err := models.GetEma(opt)
	// if err != nil {
	// 	return nil
	// }

	rps, err := models.GetRpsByOpt(opt)
	if err != nil || len(rps) == 0 {
		return nil
	}
	// 08 90+
	if rps[0].Rps_120 > rps[0].Rps_20  {
		return nil
	}

	// if rps[0].Rps_20 < 90 || rps[0].Rps_20 > 95{
	// 	return nil
	// }

	days := len(prices)
	if len(prices) > len(rps) {
		days = len(rps)
	}
	var sellPrice, dealPrice float64
	maxClose := 0.0
	isDeal := false
	for i := 1; i < days; i++ {
		preClose := prices[i-1].Close
		maxClose = math.Max(preClose, maxClose)
		unit.End = prices[i].Timestamp
		unit.Period = int64(i)
	
		if !isDeal {
			// 06 ma_12 buy
			// ratio := prices[i-1].Close / prices[i-1].HighLimit
			// if ratio > 0.91 && ratio < 0.94{
			if prices[i-1].Close != prices[i-1].HighLimit {
				dealPrice = prices[i-1].Close
				isDeal = true
				continue
			} else {
				return nil
			}
		}
		if prices[i].Open/dealPrice < 0.92 {
			sellPrice = prices[i].Open
			break
		} else if prices[i].Low/dealPrice < 0.92 {
			sellPrice = dealPrice * 0.92
			break
		}  else if prices[i].Close / maxClose < 0.95 {
			sellPrice = maxClose * 0.95
			break
		}
		
		if rps[i].Rps_5 < rps[i-1].Rps_5 && rps[i].Rps_10 > rps[i-1].Rps_10{
			sellPrice = prices[i].Close
			break
		}
		sellPrice = prices[i].Close
	}
	if days < 2 || sellPrice == 0 {
		return nil
	}
	unit.Max = (maxClose - dealPrice) / dealPrice
	unit.Net = (sellPrice - dealPrice) / dealPrice
	return unit
}
