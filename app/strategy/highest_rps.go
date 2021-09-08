/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 17:02:05
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-07 08:24:24
 */

package strategy

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"text/tabwriter"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/util"
)

type highestRps struct {
	Name   string
	RPS_20 int64
	RPS_10 int64
	RPS_5  int64
	Net    float64
}

func NewHighestRpsStrategy(name string, rps_20, rps_10, rps_5 int64) *highestRps {
	if len(name) == 0 {
		name = "highest_rps_strategy"
	}
	return &highestRps{
		Name:   name,
		RPS_20: rps_20,
		RPS_10: rps_10,
		RPS_5:  rps_5,
		Net:    1.0,
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
		// BeginAt: util.ParseDate("2020-02-03").Unix(),
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
	if rps[0].Rps_20 < 90 {
		return nil, fmt.Errorf("")
	}

	prices, err := models.GetStockPriceList(opt)
	if err != nil {
		return nil, err
	}
	// trs, err := models.GetTruesRange(opt)
	// if err != nil {
	// 	return unit, err
	// }
	days := len(prices)
	if len(prices) > len(rps) {
		days = len(rps)
	}

	// dealPrice := prices[0].Close
	// var preClose, sellPrice, atr, dealPrice float64
	var sellPrice, dealPrice float64

	// const LossCo = 1
	// const ProfitCo = 3
	isDeal := false
	for i := 2; i < days; i++ {
		// preClose = prices[i-1].Close
		// atr = trs[i-1].ATR
		unit.End = prices[i].Timestamp
		if !isDeal {
			// if rps[i-1].Rps_20 >= 90 {
			if prices[i-1].Close != prices[i-1].HighLimit {
				dealPrice = prices[i-1].Close
				isDeal = true
			} else {
				return nil, errors.New("nil")
			}
			// continue
			// } else {
			// 	return nil, errors.New("nil")
			// }
		}
		if prices[i].Open/dealPrice < 0.94 {
			sellPrice = prices[i].Open
			break
		} else if prices[i].Low/dealPrice < 0.95 {
			sellPrice = dealPrice * 0.94
			break
		}
		if i > 0 && rps[i-1].Rps_20 < 85 {
			unit.End = prices[i].Timestamp
			sellPrice = prices[i].Open
			break
		}
		// if prices[i].Low < (preClose-LossCo*atr) && (preClose-LossCo*atr)/dealPrice < 0.94 {
		// 	unit.End = prices[i].Timestamp
		// 	sellPrice = preClose - LossCo*atr
		// 	break
		// }
		// if prices[i].High > (preClose + ProfitCo*atr) {
		// 	sellPrice = preClose + ProfitCo*atr
		// 	break
		// }
		sellPrice = prices[i].Close
	}
	if days < 3 || sellPrice == 0 {
		return nil, fmt.Errorf("trade period is too short")
	}
	unit.Period = int64((unit.End - unit.Start) / (24 * 3600))
	unit.Net = (sellPrice - dealPrice) / dealPrice
	return unit, nil
}

func (v *highestRps) Kelly() error {
	nets := []interface{}{}
	periods := []interface{}{}
	rps, err := models.GetTradeResult(models.SearchOption{}, v.Name)
	if err != nil {
		return nil
	}
	for _, data := range rps {
		nets = append(nets, data.Net*100)
		periods = append(periods, data.Period)
	}
	loss, profit, lossCount, netCount := 0.0, 0.0, 0.0, 0.0
	for _, net := range nets {
		net := net.(float64)
		if net < 0 {
			lossCount++
			loss += net
		} else {
			netCount++
			profit += net
		}
	}
	var periodCount int64
	periodCount = 0
	for _, period := range periods {
		period := period.(int64)
		periodCount += period
	}
	winRate := netCount / (lossCount + netCount)
	netRatio := (profit / netCount) / (-loss / lossCount)
	expectation := netRatio*winRate - (1 - winRate)
	riskExpose := ((netRatio+1)*winRate - 1) / netRatio
	tradePeriod := int(periodCount) / len(periods)
	w := tabwriter.NewWriter(os.Stdout, 5, 5, 10, ' ', 0)
	fmt.Fprintln(w, "胜率\t赔率\t最大亏损\t期望\t平均持仓\t")
	fmt.Fprintf(w, "%.3f\t%.3f\t%.3f\t%.3f\t%d\t\n", winRate, netRatio, riskExpose, expectation, tradePeriod)
	w.Flush()
	return nil
}

type TestResult struct {
	hold     float64
	winRate  float64
	netRatio float64
	drawdown float64
	period   float64
}

func (v *highestRps) Test(start, end string) {
	// pMap := make(map[int]int)
	queue, _ := queue.NewQueue("test highest with rps", "", 100, 1000, func(data interface{}) (interface{}, error) {
		testResult := v.test(start, end)
		return testResult, nil
	}, func(datas []interface{}) error {
		// w := tabwriter.NewWriter(os.Stdout, 5, 5, 10, ' ', 0)
		// fmt.Fprintf(w, "回测区间: %s  -  %s\n", start, end)
		// fmt.Fprintln(w, "收益率\t胜率\t赔率\t最大回撤\t平均持仓\t")
		hold, winRate, netRatio, drawdown, period := 0.0, 0.0, 0.0, 0.0, 0.0
		for _, data := range datas {
			h := data.(TestResult)
			hold += h.hold
			winRate += h.winRate
			netRatio += h.netRatio
			drawdown += h.drawdown
			period += h.period
		}
		length := float64(len(datas))
		// fmt.Fprintf(w, "%.3f\t%.3f\t%.3f\t%.3f\t%.2f\t\n", hold/length, winRate/length, netRatio/length, drawdown/length, period/length)
		// w.Flush()
		v.Net *= (hold / length / 100.0)
		return nil
	})
	for i := 0; i < 10; i++ {
		queue.Push("")
	}
	queue.Close()
}

func (v *highestRps) test(start, end string) TestResult {
	dates, _ := models.GetTradeDays(models.SearchOption{
		BeginAt:  util.ToTimeStamp(start),
		EndAt:    util.ToTimeStamp(end),
		Reversed: true,
	})
	testResult := TestResult{
		hold: 100.0,
	}
	periods := []interface{}{}
	positions := []interface{}{}
	portfolio := make(map[string]position)

	hold, maxHold, drawdown := 100.0, 0.0, 1000.0
	spare := 100.0
	posCount := 0
	posMax := 5
	loss, profit, lossCount, netCount := 0.0, 0.0, 0.0, 0.0
	for i, date := range dates {
		for k, pos := range portfolio {
			if date.Timestamp == pos.End {
				spare += pos.Hold * (1 + pos.Net)
				hold += pos.Hold * pos.Net
				positions = append(positions, pos)
				posCount -= 1
				periods = append(periods, (pos.End-pos.Start)/(3600*24))
				if pos.Net < 0 {
					lossCount++
					loss += pos.Net
				} else {
					netCount++
					profit += pos.Net
				}
				delete(portfolio, k)
			}
		}
		// do not open new pos in last day
		if i == len(dates)-1 {
			break
		}
		vcps, _ := models.GetTradeResultByDay(date.Timestamp, v.Name)
		for i := 0; i < 5; i++ {
			if len(vcps) < 2 {
				break
			}
			ran := rand.Intn(len(vcps) - 1)
			vcp := vcps[ran]
			_, ok := portfolio[vcp.Code]
			if ok {
				continue
			}
			if posCount < posMax && spare > 2 {
				posHold := spare / float64(posMax-posCount)
				spare -= posHold
				posCount += 1
				portfolio[vcp.Code] = position{
					Hold:  posHold,
					Code:  vcp.Code,
					Start: vcp.Start,
					End:   vcp.End,
					Net:   vcp.Net,
				}
			}
		}
		tmp := 0.0
		for _, pos := range portfolio {
			tmp += pos.Hold
		}
		tmp = tmp + spare
		maxHold = math.Max(tmp, maxHold)
		drawdown = math.Min(tmp/maxHold, drawdown)
	}
	for _, pos := range portfolio {
		hold += pos.Hold * pos.Net
		if pos.Net < 0 {
			lossCount++
			loss += pos.Net
		} else {
			netCount++
			profit += pos.Net
		}
	}
	var periodCount int64
	periodCount = 0
	for _, period := range periods {
		period := period.(int64)
		periodCount += period
	}
	testResult.hold = hold
	testResult.winRate = netCount / (lossCount + netCount)
	testResult.netRatio = (profit / netCount) / (-loss / lossCount)
	testResult.period = float64(periodCount) / float64(len(periods))
	testResult.drawdown = drawdown
	return testResult
}
