/*
 * @Author: cedric.jia
 * @Date: 2021-09-05 22:14:02
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-27 17:59:02
 */

package strategy

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"text/tabwriter"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/util"
)

type Tester struct {
	Name     string
	Net      float64
	DrawBack float64
	NetAvg   float64
	NetCount float64
	dates  []interface{}
}

func NewTester(name string) *Tester {
	return &Tester{
		Name: name,
	}
}

func (v *Tester)Test(start, end string,posMax,lossMax int) {
	queue, _ := queue.NewQueue("test highest with rps", "", 100, 1000, func(data interface{}) (interface{}, error) {
		testResult :=v.test(start, end, posMax,lossMax)
		return testResult, nil
	}, func(datas []interface{}) error {
		hold, winRate, netRatio, netTotal, netCount, drawdown, period := 0.0, 0.0 ,0.0, 0.0, 0.0, 0.0, 0.0
		for _, data := range datas {
			h := data.(TestResult)
			hold += h.hold
			netTotal += h.netTotal
			netCount += h.netCount
			winRate += h.winRate
			netRatio += h.netRatio
			drawdown += h.drawdown
			period += h.period
		}
		length := float64(len(datas))
		w := tabwriter.NewWriter(os.Stdout, 5, 5, 10, ' ', 0)
		fmt.Fprintf(w, "回测区间: %s  -  %s\n", start, end)
		fmt.Fprintln(w, "收益率\t胜率\t赔率\t最大回撤\t平均持仓\t平均盈利\t")
		fmt.Fprintf(w, "%.3f\t%.3f\t%.3f\t%.3f\t%.2f\t%.2f\t\n", hold/length, winRate/length, netRatio/length, drawdown/length, period/length, netTotal/netCount)
		w.Flush()
		v.DrawBack = math.Min(v.DrawBack, hold/length/100.0)
		v.Net *= (hold / length/100.0)
		if netCount > 0.1 {
			v.NetAvg += netTotal / netCount
			v.NetCount += 1
		}
		// v.dates = append(v.dates,start)
		// v.index = append(v.index,opts.LineData{Value: hold/length - 80.0})
		fmt.Printf("累计收益率：%.3f\n", v.Net)
		return nil
	})
	for i := 0; i < 10; i++ {
		queue.Push("")
	}
	queue.Close()
}

func (v *Tester) test(start, end string, posMax,lossMax int) TestResult {
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
	loss, profit, lossCount, netCount := 0.0, 0.0, 0.0, 0.0
	for i, date := range dates {
		// posHold := 0.0
		for k, pos := range portfolio {
			if date.Timestamp == pos.End {
				pos.Net -= 0.0013
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
		// vcp, _ := models.GetVcp(models.SearchOption{
		// 	Timestamp: date.Timestamp,
		// })
		// if len(vcp)< 10 {
		// 	continue
		// }
		vcps, _ := models.GetTradeResultByDay(date.Timestamp, v.Name)
		for i := 0; i < posMax; i++ {
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
			BeginAt:  pos.Start,
			EndAt:    util.ToTimeStamp(end),
			Reversed: true,
		})
		len := len(prices)
		pos.Net = (prices[len-1].Close - prices[0].Close) / prices[0].Close - 0.0013
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
	testResult.netCount = netCount
	testResult.netTotal = profit
	testResult.period = float64(periodCount) / float64(len(periods))
	testResult.drawdown = drawdown
	return testResult
}