/*
 * @Author: cedric.jia
 * @Date: 2021-09-06 17:02:05
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-16 12:46:49
 */

package strategy

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"text/tabwriter"

	chart "github.cedric1996.com/go-trader/app/charts"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/util"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type highestRps struct {
	Name     string
	Date     string
	Net      float64
	DrawBack float64
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
	// 05 86-90
	// if rps[0].Rps_20 > 90 || rps[0].Rps_20 < 86 {
	// 09 86-90
	// if rps[0].Rps_20 > 90 || rps[0].Rps_20 < 86 {
	// 10 86-90 , maxClose 6% 止损
	// 11 86-90 , maxClose 5% 止损
	// if rps[0].Rps_20 > 90 || rps[0].Rps_20 < 86 {
	// 12 86-90, maxClose 7% 止损，2.5%止盈
	// 13 86-90, maxClose 8% 止损，不止盈
	// 14 86-90, maxClose 8% 止损，2 止盈
	// 15 86-90, maxClose 8% 止损，2.5 止盈
	// 16 86-90, maxClose 6% 止损，2 止盈

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

	// const LossCo = 1.5
	const ProfitCo = 2
	maxClose := 0.0
	isDeal := false
	for i := 1; i < days; i++ {
		preClose = prices[i-1].Close
		maxClose = math.Max(preClose, maxClose)
		atr = trs[i-1].ATR
		unit.End = prices[i].Timestamp
		if !isDeal {
			if prices[i-1].Close != prices[i-1].HighLimit {
				dealPrice = preClose
				isDeal = true
			} else {
				return nil, errors.New("nil")
			}
		}
		if prices[i].Open/maxClose < 0.94 {
			sellPrice = prices[i].Open
			break
		} else if prices[i].Low/maxClose < 0.94 {
			sellPrice = maxClose * 0.94
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
	queue, _ := queue.NewQueue("test highest with rps", "", 100, 1000, func(data interface{}) (interface{}, error) {
		testResult := v.test(start, end, posMax,lossMax )
		return testResult, nil
	}, func(datas []interface{}) error {
		w := tabwriter.NewWriter(os.Stdout, 5, 5, 10, ' ', 0)
		fmt.Fprintf(w, "回测区间: %s  -  %s\n", start, end)
		fmt.Fprintln(w, "收益率\t胜率\t赔率\t最大回撤\t平均持仓\t")
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
		fmt.Fprintf(w, "%.3f\t%.3f\t%.3f\t%.3f\t%.2f\t\n", hold/length, winRate/length, netRatio/length, drawdown/length, period/length)
		w.Flush()
		v.DrawBack = math.Min(v.DrawBack, hold/length/100.0)
		v.Net *= (hold / length/100.0)
		v.dates = append(v.dates,start)
		v.index = append(v.index,opts.LineData{Value: hold/length - 80.0})
		// fmt.Printf("累计收益率：%.3f\n", v.Net)
		return nil
	})
	for i := 0; i < 10; i++ {
		queue.Push("")
	}
	queue.Close()
}

func (v *highestRps) Chart() *charts.Line{
	return chart.LineChart(v.dates, v.index)
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
			pos.Net -= 0.0013
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
			} 
		}
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
		pos.Net = (prices[len-1].Close - prices[0].Close) / prices[len-1].Close - 0.0013
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
		sellPrice = preClose + 2.5*trs[0].ATR
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