/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 09:56:23
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-25 15:48:57
 */

package strategy

import (
	"errors"
	"math"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/util"
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

func NewVcpEmaStrategy(name, date string) *vcpEma {
	if len(name) == 0 {
		name = "vcp_ema_strategy"
	}
	return &vcpEma{
		Name:     name,
		Net:      1.0,
		Date:     date,
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
		datas, err := models.GetVcpNew(opt)
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

	rps, err := models.GetRpsByOpt(opt)
	if err != nil || len(rps) == 0 {
		return nil
	}

	if rps[0].Rps_20 > 0 {
		return nil
	}

	unit.RPS_5 = rps[0].Rps_5
	unit.RPS_10 = rps[0].Rps_10
	unit.RPS_20 = rps[0].Rps_20
	unit.RPS_120 = rps[0].Rps_120

	days := len(prices)
	if len(prices) != len(rps){
		days = len(rps)
	}
	var sellPrice, dealPrice float64
	maxClose := 0.0
	var maxVolume int64
	isDeal := false
	for i := 1; i < days; i++ {
		preClose := prices[i-1].Close
		maxClose = math.Max(preClose, maxClose)
		if prices[i].Volume > maxVolume {
			maxVolume = prices[i].Volume
		}

		unit.End = prices[i].Timestamp
		unit.EndDate = util.ToDate(prices[i].Timestamp)
		unit.Period = int64(i)
	
		if !isDeal {
			ratio := prices[i].Open / prices[i].PreClose
			if ratio > 0.97 && ratio < 1.03 &&  float64(prices[i].Volume) > 1.3 * float64(prices[i-1].Volume){
				unit.StartDate = util.ToDate(prices[i].Timestamp)
				// dealPrice = prices[i].Open * 1.01 
				dealPrice = prices[i].Open
				isDeal = true
				continue
			} else {
				return nil
			}
		}
		if prices[i].Open/dealPrice < 0.93 {
			sellPrice = prices[i].Open
			break
		} else if prices[i].Low/dealPrice < 0.93 {
			sellPrice = dealPrice * 0.93
			break
		}  else if prices[i].Close/maxClose < 0.95 {
			sellPrice = prices[i].Close
			break
		}  

		if prices[i].Volume >= maxVolume && (prices[i].Close < maxClose || prices[i].Close < prices[i].Open){
			sellPrice = prices[i].Close
			break
		}
		
		if prices[i].Volume >= prices[i-1].Volume && prices[i].Open > maxClose && prices[i].Close < prices[i].Open{
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

func (v *vcpEma) Pos() ([]*Pos, error) {
	t := util.ParseDate(v.Date).Unix()
	pos := make([]*Pos, 0)
	opt := models.SearchOption{
		Timestamp: t,
	}
	datas, err := models.GetVcpNew(opt)
	if err != nil {
		return nil, err
	}
	for _, data := range datas {
		opt.Code = data.RpsBase.Code
		rps, err := models.GetRpsByOpt(opt)
		if err != nil || len(rps) == 0 {
			continue
		}
		
		if rps[0].Rps_20 > 0 || rps[0].Rps_5 == 0{
			continue
		}

		prices, err := models.GetStockPriceList(opt)
		if err != nil {
			continue
		}
		
		var lossPrice, dealPrice float64
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
			LossPrice: lossPrice,
			RPS_5: rps[0].Rps_5,
			RPS_10: rps[0].Rps_10,
			RPS_20: rps[0].Rps_20,
		})
	}
	return pos, nil
}