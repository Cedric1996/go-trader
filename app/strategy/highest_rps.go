/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 17:02:05
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-09 09:42:16
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
	Name     string
	Date     string
	Net      float64
	DrawBack float64
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
	// 01 88-90
	// if rps[0].Rps_20 < 88 || rps[0].Rps_20> 90{
	// 02 87-90
	// if rps[0].Rps_20 < 87 || rps[0].Rps_20> 90{
	// 03 85-87
	// if rps[0].Rps_20 < 85 || rps[0].Rps_20> 87{
	// 04 88-92
	// if rps[0].Rps_20 < 88 || rps[0].Rps_20> 92{
	// 05 86-90
	// if rps[0].Rps_20 > 90 || rps[0].Rps_20 < 86 {
	// 06 86-89
	// 08 90-94
	// if rps[0].Rps_20 > 89 || rps[0].Rps_20 < 86 {
	// 09 86-90
	// if rps[0].Rps_20 > 90 || rps[0].Rps_20 < 86 {
	// 10 86-90 , t1 止损
	if rps[0].Rps_20 > 90 || rps[0].Rps_20 < 86 {
		return nil, fmt.Errorf("")
	}

	prices, err := models.GetStockPriceList(opt)
	if err != nil {
		return nil, err
	}
	trs, err := models.GetTruesRange(opt)
	if err != nil {
		return unit, err
	}
	days := len(prices)
	if len(prices) > len(trs) {
		days = len(trs)
	}

	// dealPrice := prices[0].Close
	var preClose, sellPrice, atr, dealPrice float64
	// var sellPrice, dealPrice float64

	const LossCo = 1
	const ProfitCo = 3
	isDeal := false
	for i := 1; i < days; i++ {
		preClose = prices[i-1].Close
		atr = trs[i-1].ATR
		unit.End = prices[i].Timestamp
		if !isDeal {
			if prices[i-1].Close != prices[i-1].HighLimit {
				dealPrice = prices[i-1].Close
				isDeal = true
			} else {
				return nil, errors.New("nil")
			}
		}
		if prices[i].Open/dealPrice < 0.94 {
			sellPrice = prices[i].Open
			break
		} else if prices[i].Low/dealPrice < 0.94 {
			sellPrice = dealPrice * 0.94
			break
		} else if prices[i].Close < dealPrice  {
			sellPrice = prices[i].Close
			break
		}
		if prices[i].Low < (preClose-LossCo*atr) && (preClose-LossCo*atr)/dealPrice < 0.94 {
			sellPrice = dealPrice * 0.94
			break
		}
		if prices[i].High > (preClose + ProfitCo*atr) {
			sellPrice = preClose + ProfitCo*atr
			break
		}
		sellPrice = prices[i].Close
	}
	if days < 2 || sellPrice == 0 {
		return nil, fmt.Errorf("trade period is too short")
	}
	unit.Period = int64((unit.End - unit.Start) / (24 * 3600))
	unit.Net = (sellPrice - dealPrice) / dealPrice
	return unit, nil
}

type TestResult struct {
	hold     float64
	winRate  float64
	netRatio float64
	drawdown float64
	period   float64
}

func (v *highestRps) Test(start, end string,posMax,lossMax int) {
	// pMap := make(map[int]int)
	queue, _ := queue.NewQueue("test highest with rps", "", 100, 1000, func(data interface{}) (interface{}, error) {
		testResult := v.test(start, end, posMax,lossMax )
		return testResult, nil
	}, func(datas []interface{}) error {
		w := tabwriter.NewWriter(os.Stdout, 5, 5, 10, ' ', 0)
		fmt.Fprintf(w, "回测区间: %s  -  %s\n", start, end)
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
		w.Flush()
		v.DrawBack = math.Min(v.DrawBack, hold/length/100.0)
		v.Net *= (hold / length/100.0)
		// fmt.Printf("累计收益率：%.3f\n", v.Net)
		return nil
	})
	for i := 0; i < 10; i++ {
		queue.Push("")
	}
	queue.Close()
}

func (v *highestRps) test(start, end string, posMax,lossMax int) TestResult {
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
	// posMax := 1
	// lossMax:= 5
	loss, profit, lossCount, netCount := 0.0, 0.0, 0.0, 0.0
	for i, date := range dates {
		// posHold := 0.0
		for k, pos := range portfolio {
			if date.Timestamp == pos.End {
				spare += pos.Hold * (1 + pos.Net)
				hold += pos.Hold * pos.Net
				positions = append(positions, pos)
				posCount -= 1
				periods = append(periods, (pos.End-pos.Start)/(3600*24))
				if pos.Net < 0 {
					lossMax -= 1
					lossCount++
					loss += pos.Net
				} else {
					netCount++
					profit += pos.Net
				}
				delete(portfolio, k)
			} else {
				// prices, _ := models.GetStockPriceList(models.SearchOption{
				// 	Code:     pos.Code,
				// 	BeginAt:  util.ToTimeStamp(start),
				// 	EndAt:    date.Timestamp,
				// 	Reversed: true,
				// })
				// if len(prices) != 0 {
				// 	len := len(prices)
				// 	pos.Net = (prices[len-1].Close - prices[0].Close) / prices[len-1].Close
				// 	posHold += pos.Hold * (1 + pos.Net)
				// }
			}
		}
		// posHold += spare
		// if (posHold / 100.0) <= 0.90 {
		// 	testResult.hold = 90.0
		// 	return testResult
		// }
		// do not open new pos in last day
		if i == len(dates)-1 {
			break
		}
		vcps, _ := models.GetTradeResultByDay(date.Timestamp, v.Name)
		for i := 0; i < 5; i++ {
			if len(vcps) < 1 {
				break
			}
			ran := rand.Intn(len(vcps))
			vcp := vcps[ran]
			_, ok := portfolio[vcp.Code]
			if ok {
				continue
			}
			if posCount < posMax && spare > 2 && lossMax >0{
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
		prices, _ := models.GetStockPriceList(models.SearchOption{
			Code:     pos.Code,
			BeginAt:  util.ToTimeStamp(start),
			EndAt:    util.ToTimeStamp(end),
			Reversed: true,
		})
		len := len(prices)
		pos.Net = (prices[len-1].Close - prices[0].Close) / prices[len-1].Close
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
		// if rps[0].Rps_20 < 90 || rps[0].Rps_20 > 96 {
		// if rps[0].Rps_20 > 94 || rps[0].Rps_20 < 90 {
		if rps[0].Rps_20 < 86 || rps[0].Rps_20> 90{
			continue
		}
		prices, err := models.GetStockPriceList(opt)
		if err != nil {
			continue
		}
		trs, err := models.GetTruesRange(opt)
		if err != nil {
			continue
		}
		var preClose, sellPrice, dealPrice, lossPrice float64
		preClose = prices[0].Close
		if prices[0].Close != prices[0].HighLimit {
			dealPrice = prices[0].Close
		}
		if dealPrice < 0.1 {
			continue
		}
		lossPrice = dealPrice * 0.94
		sellPrice = preClose + 3*trs[0].ATR
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
