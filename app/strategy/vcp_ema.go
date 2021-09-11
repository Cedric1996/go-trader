/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 09:56:23
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-06 11:43:40
 */

package strategy

import (
	"github.cedric1996.com/go-trader/app/models"
)

func VcpEmaTradeSignal(sig TradeSignal) (unit *TradeUnit) {
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
	emas, err := models.GetEma(opt)
	if err != nil {
		return nil
	}
	days := len(prices)
	if len(prices) > len(emas) {
		days = len(emas)
	}
	// dealPrice := prices[0].Close
	var sellPrice, dealPrice float64
	isDeal := false
	for i := 1; i < days; i++ {
		// preClose = prices[i-1].Close
		close := prices[i].Close
		ema := emas[i]
		if ema.MA_6 < 0.1 || ema.MA_12 < 0.1 || ema.MA_26 < 0.1 {
			return nil
		}
		unit.End = prices[i].Timestamp
		if !isDeal {
			if close <= ema.MA_6 && close >= ema.MA_12 {
				dealPrice = close
				isDeal = true
				continue
			} else {
				return nil
			}
		}
		if prices[i].Low < ema.MA_12 {
			unit.End = prices[i].Timestamp
			sellPrice = ema.MA_12
			break
		}
		if ema.MA_12 > ema.MA_6 && ema.MA_6 > prices[i].Low && ema.MA_6 < prices[i].High {
			sellPrice = ema.MA_6
			break
		}
		if close/dealPrice < 0.95 {
			sellPrice = close
			break
		}
		sellPrice = close
	}
	if days < 2 || sellPrice == 0 {
		return nil
	}
	unit.Period = int64((unit.End - unit.Start) / (24 * 3600))
	unit.Net = (sellPrice - dealPrice) / dealPrice
	return unit
}
