/*
 * @Author: cedric.jia
 * @Date: 2021-09-04 13:58:15
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-05 15:37:16
 */

package strategy

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

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
}

type tradeUnit struct {
	Code   string    `bson:"code"`
	Start  time.Time `bson:"start"`
	End    time.Time `bson:"end"`
	Period int64     `bson:"period"`
	Net    float64   `bson:"net"`
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
	}
}

func (v *vcp) Run() error {
	queue, err := queue.NewQueue("vcp with true range", "", 1000, 1000, func(data interface{}) (interface{}, error) {
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
	// dealPrice := prices[0].Close
	var preClose, sellPrice, atr, dealPrice float64
	const BreakCo = 0.3
	const LossCo = 0.3
	const ProfitCo = 2.5
	isDeal := false
	for i := 1; i < days; i++ {
		preClose = prices[i-1].Close
		atr = trs[i-1].ATR
		if !isDeal {
			if prices[i].High > (preClose + BreakCo*atr) {
				dealPrice = preClose + BreakCo*atr
				isDeal = true
				continue
			} else {
				return unit, errors.New("")
			}
		}
		if prices[i].Low < (preClose - LossCo*atr) {
			unit.End = time.Unix(prices[i].Timestamp, 0)
			sellPrice = preClose - LossCo*atr
			break
		}
		if prices[i].High > (preClose + ProfitCo*atr) {
			unit.End = time.Unix(prices[i].Timestamp, 0)
			sellPrice = preClose + ProfitCo*atr
			break
		}
		sellPrice = prices[i].Close
	}
	if days < 2 || sellPrice == 0 {
		return unit, fmt.Errorf("trade period is too short")
	}
	unit.Period = int64(unit.End.Sub(unit.Start).Hours() / 24)
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
	opt := models.SearchOption{Limit: 1000, Skip: 0}
	for {
		vcpTrs, err := models.GetVcpTr(opt, v.Name)
		if err != nil {
			return nil
		}
		for _, data := range vcpTrs {
			t := data.Start.Unix()
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
	timestamps := v.dates
	dates := []interface{}{}
	percents := []opts.BarData{}
	percentVal := []interface{}{}

	// keys := make([]int, 0, len(vcpMap))
	// for k := range vcpMap {
	// 	keys = append(keys, int(k))
	// }
	// sort.Ints(keys)
	for _, k := range timestamps {
		k := k.(int64)
		val, ok := vcpMap[k]
		if ok {
			dates = append(dates, util.ToDate(k))
			percent := 100 * val.Long / (val.Long - val.Short)
			percents = append(percents, opts.BarData{Value: percent})
			percentVal = append(percentVal, percent)
			// longs = append(longs, opts.BarData{Value: vcpMap[k].Long})
			// shorts = append(shorts, opts.BarData{Value: vcpMap[k].Short})
		}
	}
	v.percents = percentVal
	// bar := chart.BarCharts(dates, longs, shorts)
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
	winRate := netCount / (lossCount + netCount)
	netRatio := (profit / netCount) / (-loss / lossCount)
	expectation := netRatio*winRate - (1 - netRatio)
	riskExpose := ((netRatio+1)*winRate - 1) / netRatio

	w := tabwriter.NewWriter(os.Stdout, 5, 5, 10, ' ', 0)
	fmt.Fprintln(w, "胜率\t赔率\t最大亏损\t期望\t")
	fmt.Fprintf(w, "%.3f\t%.3f\t%.3f\t%.3f\t\n", winRate, netRatio, riskExpose, expectation)
	w.Flush()
	return nil
}
