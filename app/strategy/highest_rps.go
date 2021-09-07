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
	"math/rand"
	"os"
	"sort"
	"text/tabwriter"

	chart "github.cedric1996.com/go-trader/app/charts"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/util"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type highestRps struct {
	Name     string
	dates    []interface{}
	percents []interface{}
	netVal   []interface{}
	periods  []interface{}
}

func NewHighestRpsStrategy(name string) *highestRps {
	if len(name) == 0 {
		name = "highest_rps_strategy"
	}
	return &highestRps{
		Name:     name,
		dates:    make([]interface{}, 0),
		percents: make([]interface{}, 0),
		netVal:   make([]interface{}, 0),
		periods:  make([]interface{}, 0),
	}
}

func (v *highestRps) Run() error {
	queue, err := queue.NewQueue("highest with rps", "", 100, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(*models.HighestRps)
		unit, _ := highestRpsSignal(TradeSignal{
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

func highestRpsSignal(sig TradeSignal) (unit *TradeUnit, err error) {
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
		return nil, err
	}
	rps, err := models.GetRpsByOpt(opt)
	if err != nil {
		return nil, err
	}
	days := len(prices)
	if len(prices) > len(rps) {
		days = len(rps)
	}
	// dealPrice := prices[0].Close
	var sellPrice, dealPrice float64
	// const LossCo = 1
	isDeal := false
	for i := 0; i < days; i++ {
		unit.End = prices[i].Timestamp
		if !isDeal {
			if rps[i].Rps_10 > 0 && rps[i].Rps_5 > 0 {
				dealPrice = prices[i].Close
				isDeal = true
				continue
			} else {
				return nil, errors.New("nil")
			}
		}
		if prices[i].Low/dealPrice < 0.93 {
			if prices[i].Open/dealPrice > 0.95 {
				sellPrice = dealPrice * 0.94
			} else {
				sellPrice = prices[i].Open * 0.99
			}
			break
		}
		if i > 0 && rps[i-1].Rps_20 < 85 {
			unit.End = prices[i].Timestamp
			sellPrice = prices[i].Open
			break
		}
		sellPrice = prices[i].Close
		// 手动止盈
		// if i%20 == 0 {
		// 	net := (sellPrice-dealPrice)/dealPrice + 1
		// 	std := math.Pow(1.247, float64(i/20))
		// 	if net < std {
		// 		break
		// 	}
		// }
	}
	if days < 2 || sellPrice == 0 {
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
	// nets := v.netVal
	// periods := v.periods
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
	percent := []interface{}{}
	count := []opts.BarData{}
	pMap := make(map[int]int)
	queue, _ := queue.NewQueue("test highest with rps", "", 100, 1000, func(data interface{}) (interface{}, error) {
		hold := v.test(start, end)
		return hold, nil
	}, func(datas []interface{}) error {
		for _, data := range datas {
			h := data.(float64)
			hold := int(h)
			_, ok := pMap[hold]
			if !ok {
				pMap[hold] = 0
			}
			pMap[hold] += 1
		}
		keys := make([]int, 0, len(pMap))
		for k := range pMap {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			percent = append(percent, k)
			count = append(count, opts.BarData{Value: pMap[k]})
		}
		bar := chart.BarCharts(percent, count)
		barChart := chart.NewBarChart(v.Name + "_test")
		barChart.BarPage(bar)
		return nil
	})
	for i := 0; i < 1000; i++ {
		queue.Push("")
	}
	queue.Close()
	// w := tabwriter.NewWriter(os.Stdout, 5, 5, 10, ' ', 0)
	// fmt.Fprintln(w, "净值\t交易次数\t")
	// fmt.Fprintf(w, "%.3f\t%d\t\n", hold, len(positions))
	// w.Flush()
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
	nets := []interface{}{}
	periods := []interface{}{}
	positions := []interface{}{}
	portfolio := make(map[string]position)
	hold := 100.0
	spare := 100.0
	posCount := 0
	posMax := 5
	for i, date := range dates {
		for k, pos := range portfolio {
			if date.Timestamp == pos.End {
				spare += pos.Hold * (1 + pos.Net)
				hold += pos.Hold * pos.Net
				positions = append(positions, pos)
				posCount -= 1
				nets = append(nets, pos.Net*100)
				periods = append(periods, (pos.End-pos.Start)/(3600*24))
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
	}
	for _, pos := range portfolio {
		hold += pos.Hold * pos.Net
	}
	return testResult
}
