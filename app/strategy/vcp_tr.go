/*
 * @Author: cedric.jia
 * @Date: 2021-09-04 13:58:15
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-06 23:16:14
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
	"github.cedric1996.com/go-trader/app/util"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
)

type vcp struct {
	Name     string
	Net      float64
	DrawBack float64
	dates    []interface{}
	percents []interface{}
	netVal   []interface{}
	periods  []interface{}
}

type TradeUnit struct {
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
		Net:      1.0,
		DrawBack: 1.0,
		dates:    make([]interface{}, 0),
		percents: make([]interface{}, 0),
		netVal:   make([]interface{}, 0),
		periods:  make([]interface{}, 0),
	}
}

func (v *vcp) Run() error {
	queue, err := queue.NewQueue("vcp with ema", "", 100, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(*models.Vcp)
		unit, _ := handleTradeSignal(TradeSignal{
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
		// BeginAt: util.ParseDate("2020-06-02").Unix(),
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

func handleTradeSignal(sig TradeSignal) (unit *TradeUnit, err error) {
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
	const BreakCo = 0.5
	const LossCo = 1
	const ProfitCo = 3
	isDeal := false
	high := 0.0
	for i := 1; i < days; i++ {
		preClose = prices[i-1].Close
		atr = trs[i-1].ATR
		unit.End = prices[i].Timestamp
		if !isDeal {
			if prices[i].High > (preClose+BreakCo*atr) || prices[i].Low < (preClose-BreakCo*atr) {
				dealPrice = prices[i].Close
				isDeal = true
				high = dealPrice
				continue
			} else {
				return nil, errors.New("")
			}
		}
		high = math.Max(high, preClose)
		if prices[i].Low/dealPrice < 0.93 {
			if prices[i].Open/dealPrice > 0.95 {
				sellPrice = dealPrice * 0.94
			} else {
				sellPrice = prices[i].Open * 0.99
			}
			break
		}
		if prices[i].Open > dealPrice && prices[i].Low < high*0.8 {
			unit.End = prices[i].Timestamp
			sellPrice = high * 0.8
			break
		}
		if prices[i].Low < (preClose-LossCo*atr) && (preClose-LossCo*atr)/dealPrice < 0.94 {
			unit.End = prices[i].Timestamp
			sellPrice = preClose - LossCo*atr
			break
		}
		if prices[i].High > (preClose + ProfitCo*atr) {
			sellPrice = preClose + ProfitCo*atr
			break
		}
		sellPrice = prices[i].Close

		// 手动止盈
		if i%20 == 0 {
			net := (sellPrice-dealPrice)/dealPrice + 1
			std := math.Pow(1.247, float64(i/20))
			if net < std {
				break
			}
		}
	}
	if days < 2 || sellPrice == 0 {
		return nil, fmt.Errorf("trade period is too short")
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
		vcpTrs, err := models.GetTradeResult(opt, v.Name)
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

func (v *vcp) Test(start, end string,posMax,lossMax int) {
	// pMap := make(map[int]int)
	queue, _ := queue.NewQueue("test vcpTr with rps", "", 100, 1000, func(data interface{}) (interface{}, error) {
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

func (v *vcp) test(start, end string, posMax,lossMax int) TestResult {
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