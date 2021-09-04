/*
 * @Author: cedric.jia
 * @Date: 2021-09-04 13:58:15
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-04 18:59:20
 */

package strategy

import (
	"fmt"
	"time"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
)

type vcp struct {
	Name     string
	dataChan chan interface{}
}

type tradeUnit struct {
	Code   string    `bson:"code"`
	Start  time.Time `bson:"start"`
	End    time.Time `bson:"end"`
	Period int64     `bson:"period"`
	Net    float64   `bson:"net"`
}

func NewVcpStrategy() *vcp {
	return &vcp{}
}

func (v *vcp) Run() error {
	queue, err := queue.NewQueue("vcp with true range", "", 100, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(*models.Vcp)
		unit, err := handleTradeSignal(TradeSignal{
			Code:      datum.RpsBase.Code,
			StartUnix: datum.RpsBase.Timestamp,
		})
		if err != nil {
			return nil, err
		}

		return unit, nil
	}, func(datas []interface{}) error {
		if err := models.InsertMany(datas, "vcp_tr_strategy"); err != nil {
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
		vcps, err := models.GetVcp(opt)
		if err != nil {
			break
		}
		for _, vcp := range vcps {
			queue.Push(vcp)
		}
		if len(vcps) != 1000 {
			break
		}
		opt.Skip += 1000
	}
	queue.Close()
	return nil
}

func handleTradeSignal(sig TradeSignal) (unit *tradeUnit, err error) {
	opt := models.SearchOption{
		Code:     sig.Code,
		BeginAt:  sig.StartUnix,
		Reversed: true,
	}
	unit = &tradeUnit{
		Code:  sig.Code,
		Start: time.Unix(sig.StartUnix, 0),
	}
	prices, err := models.GetStockPriceList(opt)
	if err != nil {
		return unit, err
	}
	trs, err := models.GetTruesRange(opt)
	if err != nil {
		return unit, err
	}
	days := len(prices)
	if len(prices) > len(trs) {
		days = len(trs)
	}
	dealPrice := prices[0].Close
	var preClose, close, sellPrice float64
	const LossCo = 0.8
	const ProfitCo = 2.5
	for i := 1; i < days; i++ {
		close = prices[i].Close
		preClose = prices[i-1].Close
		if close < (preClose - LossCo*trs[i-1].ATR) {
			unit.End = time.Unix(prices[i].Timestamp, 0)
			sellPrice = preClose - LossCo*trs[i-1].ATR
			break
		}
		if close > (preClose + ProfitCo*trs[i-1].ATR) {
			unit.End = time.Unix(prices[i].Timestamp, 0)
			sellPrice = preClose + ProfitCo*trs[i-1].ATR
			break
		}
	}
	if days < 2 || sellPrice == 0 {
		return unit, fmt.Errorf("trade period is too short")
	}
	unit.Period = int64(unit.End.Sub(unit.Start).Hours() / 24)
	unit.Net = (sellPrice - dealPrice) / dealPrice
	return unit, nil
}
