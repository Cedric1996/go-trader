/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 09:56:23
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-28 22:39:42
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
	dates    []interface{}
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
		dates:    make([]interface{}, 0),
	}
}

func (v *vcpEma) Run() error {
	queue, err := queue.NewQueue("highest with rps", "", 100, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(*models.Vcp)
		unit := v.vcpEmaTradeSignal(TradeSignal{
			Code:      datum.RpsBase.Code,
			StartUnix: datum.RpsBase.Timestamp,
			// Data: datum.DealPrice,
			Data: 0,
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
			// if data.Period > 1{
			queue.Push(data)
			// }
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
	// ratio := prices[0].Close / prices[0].PreClose
	// if ratio <= 1.0 {
	// 	return nil
	// }

	// rps, err := models.GetRpsByOpt(opt)
	// if err != nil || len(rps) == 0 {
	// 	return nil
	// }

	// unit.RPS_5 = rps[0].Rps_5
	// unit.RPS_10 = rps[0].Rps_10
	// unit.RPS_20 = rps[0].Rps_20
	// unit.RPS_120 = rps[0].Rps_120

	days := len(prices)
	length := 60
	if length > days {
		length = days
	}
	// if len(prices) != len(rps){
	// 	days = len(rps)
	// }
	var sellPrice, dealPrice float64
	maxClose := 0.0
	// maxVolume :=  prices[0].Volume
	// isDeal := false
	unit.StartDate = util.ToDate(prices[0].Timestamp)
	dealPrice = prices[0].Close
	drawBack := 1.0
	maxClose = dealPrice

	for i := 1; i < length; i++ {
		// preClose := prices[i-1].Close
		// maxClose = math.Max(preClose, maxClose)
		// if prices[i].Volume > maxVolume {
		// 	maxVolume = prices[i].Volume
		// }

		unit.End = prices[i].Timestamp
		unit.EndDate = util.ToDate(prices[i].Timestamp)
		unit.Period = int64(i)

		if prices[i].Close < maxClose {
			drawBack = math.Min(drawBack, prices[i].Close/maxClose)
		} else {
			maxClose = math.Max(maxClose, prices[i].Close)
		}
		// if !isDeal {
		// 	unit.StartDate = util.ToDate(prices[i-1].Timestamp)
		// 	dealPrice = prices[i-1].Close
		// 	isDeal = true
		// 	// ratio := prices[i].Close / prices[i].PreClose
		// 	// if ratio > 0.97 && ratio < 1.03 &&  float64(prices[i].Volume) > 1.3 * float64(prices[i-1].Volume){
		// 	// if  float64(prices[i].Volume) > 1.3 * float64(prices[i-1].Volume) {
		// 	// 	dealPrice = prices[i].Close
		// 	// 	isDeal = true
		// 	// 	continue
		// 	// }  else {
		// 	// 	return nil
		// 	// }
		// }
		// if prices[i].Open/dealPrice < 0.93 {
		// 	sellPrice = prices[i].Open
		// 	break
		// }

		// // ratio := prices[1].Open / prices[1].PreClose
		// // if ratio < 0.97 || ratio > 1.03{
		// // 	sellPrice = prices[1].Open
		// // 	break
		// // }

		// if prices[i].Low/dealPrice < 0.93 {
		// 	sellPrice = dealPrice * 0.93
		// 	break
		// }  else if maxClose > prices[0].Close && prices[i].Close/maxClose < 0.95 {
		// 	sellPrice = prices[i].Close
		// 	break
		// }

		// // if float64(prices[1].Volume) <  1.3 * float64(prices[0].Volume) {
		// // 	sellPrice = prices[1].Close
		// // 	break
		// // }

		// if prices[i].Volume > maxVolume && (prices[i].Close < maxClose || prices[i].Close < prices[i].Open){
		// 	sellPrice = prices[i].Close
		// 	break
		// }

		// if prices[i].Volume > prices[i-1].Volume && prices[i].Open > maxClose && prices[i].Close < prices[i].Open{
		// 	sellPrice = prices[i].Close
		// 	break
		// }

		sellPrice = prices[i].Close
	}
	if days < 2 || sellPrice == 0 {
		return nil
	}
	unit.Max = (maxClose - dealPrice) / dealPrice
	unit.Net = (sellPrice - dealPrice) / dealPrice
	unit.Drawback = drawBack
	return unit
}

func (v *vcpEma) Pos() ([]*Pos, *map[string]int, error) {
	// t := util.ParseDate(v.Date).Unix()
	pos := make([]*Pos, 0)
	opt := models.SearchOption{
		BeginAt: util.ParseDate("2021-08-01").Unix(),
		EndAt:   util.ParseDate("2021-09-31").Unix(),
		// Timestamp: t,
	}
	datas, err := models.GetVcpNew(opt)
	if err != nil {
		return nil, nil, err
	}
	modMap := map[string]int{}
	for _, data := range datas {
		opt.Code = data.RpsBase.Code
		rps, err := models.GetRpsByOpt(opt)
		if err != nil || len(rps) == 0 {
			continue
		}
		// if rps[0].Rps_20 > 0 {
		// 	continue
		// }
		// prices, err := models.GetStockPriceList(models.SearchOption{
		// 	Code:   data.RpsBase.Code,
		// 	BeginAt:  t,
		// 	Reversed: true,
		// })
		// if err != nil || len(prices) == 0 {
		// 	continue
		// }
		// if prices[1].Close / prices[1].PreClose <=1 {
		// 	continue
		// }
		// if float64(prices[1].Volume) * 1.3 > float64(prices[0].Volume) {
		// 	continue
		// }

		// mods, _ := models.GetStockModule(models.SearchOption{
		// 	Code: data.RpsBase.Code,
		// })
		// mod := ""
		// for _, mo := range mods {
		// 	_, ok := modMap[mo.ModuleName]
		// 	if !ok {
		// 		modMap[mo.ModuleName] = 1
		// 	} else {
		// 		modMap[mo.ModuleName] += 1
		// 	}
		// 	mod += mo.ModuleName
		// 	mod+= ","
		// }
		info, err := models.GetSecurityByCode(opt.Code)
		pos = append(pos, &Pos{
			Code:   opt.Code,
			Name:   info.DisplayName,
			Period: data.Period,
			RPS_5:  rps[0].Rps_5,
			RPS_10: rps[0].Rps_10,
			RPS_20: rps[0].Rps_20,
			Mod:    util.ToDate(data.RpsBase.Timestamp),
		})
	}
	return pos, &modMap, nil
}
