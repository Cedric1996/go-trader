/*
 * @Author: cedric.jia
 * @Date: 2021-09-04 13:58:15
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-06 09:08:01
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
	"github.cedric1996.com/go-trader/app/util"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
)

type vcp struct {
	Name     string
	dates    []interface{}
	percents []interface{}
	netVal   []interface{}
	periods  []interface{}
}

type tradeUnit struct {
	Code   string  `bson:"code"`
	Start  int64   `bson:"start"`
	End    int64   `bson:"end"`
	Period int64   `bson:"period"`
	Net    float64 `bson:"net"`
}

func NewVcpStrategy(name string) *vcp {
	if len(name) == 0 {
		name = "vcp_tr_strategy"
	}
	return &vcp{
		Name:     name,
		dates:    make([]interface{}, 0),
		percents: make([]interface{}, 0),
		netVal:   make([]interface{}, 0),
		periods:  make([]interface{}, 0),
	}
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
		Start: sig.StartUnix,
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
	// dealPrice := prices[0].Close
	var preClose, sellPrice, atr, dealPrice float64
	const BreakCo = 0.3
	const LossCo = 1.5
	const ProfitCo = 3
	isDeal := false
	for i := 1; i < days; i++ {
		preClose = prices[i-1].Close
		atr = trs[i-1].ATR
		unit.End = prices[i].Timestamp
		if !isDeal {
			if prices[i].High > (preClose+BreakCo*atr) && prices[i].Close > (preClose+BreakCo*atr) {
				dealPrice = prices[i].Close
				isDeal = true
				continue
			} else {
				return unit, errors.New("")
			}
		}
		if prices[i].Low < (preClose - LossCo*atr) {
			unit.End = prices[i].Timestamp
			sellPrice = preClose - LossCo*atr
			break
		}
		if prices[i].High > (preClose + ProfitCo*atr) {
			sellPrice = preClose + ProfitCo*atr
			break
		}
		if prices[i].Close/dealPrice < 0.94 {
			sellPrice = prices[i].Close
			break
		}
		sellPrice = prices[i].Close
	}
	if days < 2 || sellPrice == 0 {
		return unit, fmt.Errorf("trade period is too short")
	}
	unit.Period = int64((unit.End - unit.Start) / (24 * 3600))
	unit.Net = (sellPrice - dealPrice) / dealPrice
	return unit, nil
}

func (v *vcp) Output() error {
	barChart := chart.NewBarChart(v.Name)
	barChart.BarPage(v.highLowIndex(), v.vcpTr(), v.net())
	return nil
}

func (v *vcp) vcpTr() *charts.Bar {
	type stat struct {
		Long  int
		Short int
	}
	vcpMap := make(map[int64]stat)
	nets := []interface{}{}
	periods := []interface{}{}
	opt := models.SearchOption{Limit: 1000, Skip: 0}
	for {
		vcpTrs, err := models.GetVcpTr(opt, v.Name)
		if err != nil {
			return nil
		}
		for _, data := range vcpTrs {
			t := data.Start
			_, ok := vcpMap[t]
			if !ok {
				vcpMap[t] = stat{Long: 0, Short: 0}
			}
			long, short := 0, 0
			if data.Net < 0 {
				short = -1
			} else {
				long = 1
			}
			nets = append(nets, data.Net*100)
			periods = append(periods, data.Period)

			tmp := vcpMap[t]
			tmp.Long += long
			tmp.Short += short
			vcpMap[t] = tmp
		}
		if len(vcpTrs) != 1000 {
			break
		}
		opt.Skip += 1000
	}
	v.netVal = nets
	v.periods = periods
	timestamps := v.dates
	dates := []interface{}{}
	percents := []opts.BarData{}
	percentVal := []interface{}{}
	for _, k := range timestamps {
		k := k.(int64)
		val, ok := vcpMap[k]
		if ok {
			dates = append(dates, util.ToDate(k))
			percent := 100 * val.Long / (val.Long - val.Short)
			percents = append(percents, opts.BarData{Value: percent})
			percentVal = append(percentVal, percent)
		}
	}
	v.percents = percentVal
	bar := chart.BarCharts(dates, percents)
	return bar
}

func (v *vcp) highLowIndex() *charts.Bar {
	nhnls, err := models.GetHighLowIndex(models.SearchOption{
		Reversed: true, BeginAt: util.ParseDate("2019-03-08").Unix()})
	if err != nil {
		return nil
	}
	dates := []interface{}{}
	timestamps := []interface{}{}
	index := []opts.BarData{}
	for i := 1; i < len(nhnls); i++ {
		if nhnls[i].Index > nhnls[i-1].Index {
			v := nhnls[i]
			dates = append(dates, v.Date)
			timestamps = append(timestamps, v.Timestamp)
			index = append(index, opts.BarData{Value: v.Index})
		}
	}
	v.dates = timestamps
	bar := chart.BarCharts(dates, index)
	return bar
}

func (v *vcp) percent() *charts.Bar {
	percent := []interface{}{}
	count := []opts.BarData{}
	pMap := make(map[int]int)
	for _, v := range v.percents {
		v := v.(int)
		_, ok := pMap[v]
		if !ok {
			pMap[v] = 0
		}
		pMap[v] += 1
	}
	keys := make([]int, 0, len(pMap))
	for k := range pMap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		percent = append(percent, k)
		count = append(count, opts.BarData{Value: pMap[k]})
	}
	bar := chart.BarCharts(percent, count)
	return bar
}

func (v *vcp) net() *charts.Bar {
	net := []interface{}{}
	count := []opts.BarData{}
	pMap := make(map[int]int)
	for _, v := range v.netVal {
		f := v.(float64)
		v := int(f)
		_, ok := pMap[v]
		if !ok {
			pMap[v] = 0
		}
		pMap[v] += 1
	}
	keys := make([]int, 0, len(pMap))
	for k := range pMap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		net = append(net, k)
		count = append(count, opts.BarData{Value: pMap[k]})
		// count = append(count, opts.ScatterData{Value: pMap[k]})
	}
	bar := chart.BarCharts(net, count)
	return bar
}

func (v *vcp) Kelly() error {
	nets := v.netVal
	periods := v.periods
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

type position struct {
	Code  string
	Hold  float64
	Net   float64
	Start int64
	End   int64
}

func (v *vcp) Test(start, end string) {
	percent := []interface{}{}
	count := []opts.BarData{}
	pMap := make(map[int]int)
	queue, _ := queue.NewQueue("test vcp with true range", "", 100, 1000, func(data interface{}) (interface{}, error) {
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
		barChart := chart.NewBarChart(v.Name)
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

func (v *vcp) test(start, end string) float64 {
	dates, _ := models.GetTradeDays(models.SearchOption{
		BeginAt:  util.ToTimeStamp(start),
		EndAt:    util.ToTimeStamp(end),
		Reversed: true,
	})
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
				delete(portfolio, k)
			}
		}
		// do not open new pos in last day
		if i == len(dates)-1 {
			break
		}
		vcps, _ := models.GetVcpTrByDay(date.Timestamp, v.Name)
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
	return hold
}
