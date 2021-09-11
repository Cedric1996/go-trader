/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-09 09:30:48
 */
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.cedric1996.com/go-trader/app"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/strategy"
	"github.cedric1996.com/go-trader/app/util"
	"github.com/urfave/cli"
)

var (
	CmdStrategy = cli.Command{
		Name:  "strategy",
		Usage: "cal strategy",
		Subcommands: []cli.Command{
			subcmdVcpTr,
			subcmdHighestRps,
		},
	}

	subcmdVcpTr = cli.Command{
		Name:  "vcp",
		Usage: "vcp tr strategy",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "output,o"},
		},
		Action: runVcpTr,
	}

	subcmdHighestRps = cli.Command{
		Name:  "high",
		Usage: "",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "run,r"},
		},
		Action: runHighestRps,
	}

	subcmdPosHighestRps = cli.Command{
		Name:  "high",
		Usage: "",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "date,d"},
		},
		Action: runHighestRpsPos,
	}
)

func runHighestRpsPos(c *cli.Context) error {
	app.Init()
	date := c.String("date")
	if len(date) == 0 {
		return nil
	}
	v := strategy.NewHighestRpsStrategy("", date)
	datas := []string{}
	pos, err := v.Pos()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 5, 5, 10, ' ', 0)
	fmt.Fprintln(w, "代码\t名称\t买入\t止盈\t止损\t5日强\t10日强\t20日强\t")
	for _, v := range pos {
		datas = append(datas, v.Code)
		fmt.Fprintf(w, "%s\t%s\t%.2f\t%.2f\t%.2f\t%d\t%d\t%d\t\n", v.Code, v.Name, v.DealPrice, v.SellPrice, v.LossPrice, v.RPS_5, v.RPS_10, v.RPS_20)
	}
	w.Flush()
	if err := service.GenerateVcpFile(datas); err != nil {
		return err
	}
	return nil
}

func runVcpTr(c *cli.Context) error {
	app.Init()
	// v := strategy.NewVcpStrategy("vcp_ema_strategy")
	name := "vcp_tr_strategy_02"
	v := strategy.NewVcpStrategy(name)

	run := c.Bool("run")
	if run {
		if err := models.InitStrategyIndexes(name); err != nil {
			return err
		}
		if err := v.Run(); err != nil {
			return fmt.Errorf("execute vcp tr strategy fail, please check it", err)
		}
	}
	nums := []string{"02"}
	
	// nums := []string{"01", "02", "03", "04", "05", "06", "07", "08"}
	res := []*testResult{}
	for _, s := range nums {
		res = append(res, vcpTrTest(s, 1, 10))
		res = append(res, vcpTrTest(s, 2, 10))
		res = append(res, vcpTrTest(s, 3, 10))
		res = append(res, vcpTrTest(s, 4, 10))
		res = append(res, vcpTrTest(s, 5, 10))
		res = append(res, vcpTrTest(s, 6, 10))
		res = append(res, vcpTrTest(s, 7, 10))
		res = append(res, vcpTrTest(s, 8, 10))
		res = append(res, vcpTrTest(s, 9, 10))
	}
	for _, r := range res {
	fmt.Printf("策略序号：%s, 总收益: %3f, 最大月亏损: %3f, 最大持仓: %d, 最大亏损数: %d\n",r.num,  r.net, r.drawBack, r.posMax, r.lossMax)
	}

	return nil
}


func vcpTrTest(num string, posMax, lossMax int) *testResult{
	name := fmt.Sprintf("vcp_tr_strategy_%s", num)

	// v := strategy.NewHighestRpsStrategy(name, 90, 90, 0)
	v := strategy.NewVcpStrategy(name)
	netFunc := func() (float64, float64) {
		v.Net = 1.0
		v.DrawBack = 1.0
		// start := util.ParseDate("2019-04-01") 
		// start := util.ParseDate("2019-04-10")
		// start := util.ParseDate("2019-04-11")

		start := util.ParseDate("2019-04-06")
		// start := util.ParseDate("2019-09-01")
		end := start.AddDate(0, 0, 13)
		for i := 0; i < 29; i++ {
			v.Test(util.ToDate(start.Unix()), util.ToDate(end.Unix()),posMax, lossMax)
			// v.Test(util.ToDate(start.AddDate(0, 0, 7).Unix()), util.ToDate(end.AddDate(0, 0, 7).Unix()))
			v.Test(util.ToDate(start.AddDate(0, 0, 14).Unix()), util.ToDate(end.AddDate(0, 0, 14).Unix()),posMax, lossMax)
			// v.Test(util.ToDate(start.AddDate(0, 0, 21).Unix()), util.ToDate(end.AddDate(0, 0, 21).Unix()))
			start = start.AddDate(0, 1, 0)
			end = end.AddDate(0, 1, 0)
		}
		fmt.Println("完成测试")
		return v.Net, v.DrawBack
	}
	netTotals, drawBackTotals := 0.0, 0.0
	count := 5
	for i := 0; i < count; i++ {
		net, drawBack := netFunc()
		netTotals += net
		drawBackTotals += drawBack
	}
	return &testResult{
		num: num,
		net:  netTotals/float64(count),
		drawBack: drawBackTotals/float64(count),
		posMax: posMax,
		lossMax: lossMax,
	}
}

func runHighestRps(c *cli.Context) error {
	app.Init()
	name := "highest_rps_strategy_10"

	v := strategy.NewHighestRpsStrategy(name, "")
	run := c.Bool("run")
	if run {
		if err := models.InitStrategyIndexes(name); err != nil {
			return err
		}
		if err := v.Run(); err != nil {
			return fmt.Errorf("execute vcp tr strategy fail, please check it", err)
		}
	}

	// taskQueue := queue.NewTaskQueue("highest test", 10, func(data interface{}) error {
	// 	num := data.(string)
	// 	r := highestRpsTest(num)
	// 	fmt.Printf("策略序号：%s, 总收益: %3f, 最大月亏损: %3f\n",r.num,  r.net, r.drawBack)
	// 	return nil
	// }, func(dateChan *chan interface{}) {
	// 	// nums := []string{"09"}
	// 	nums := []string{"01", "02", "03", "04", "05", "06", "07", "08","09"}
	// 	for _, num := range nums {
	// 		*dateChan <- num
	// 	}
	// })
	// if err := taskQueue.Run(); err != nil {
	// 	return err
	// }
	nums := []string{"05"}
	
	// nums := []string{"01", "02", "03", "04", "05", "06", "07", "08"}
	res := []*testResult{}
	for _, s := range nums {
		res = append(res, highestRpsTest(s, 6, 10))
		res = append(res, highestRpsTest(s, 7, 10))
		res = append(res, highestRpsTest(s, 8, 10))
	}
	for _, r := range res {
	fmt.Printf("策略序号：%s, 总收益: %3f, 最大月亏损: %3f, 最大持仓: %d, 最大亏损数: %d\n",r.num,  r.net, r.drawBack, r.posMax, r.lossMax)
	}

	return nil
}

type testResult struct {
	num string
	net float64
	drawBack float64
	posMax int
	lossMax int
}

func highestRpsTest(num string, posMax, lossMax int) *testResult{
	name := fmt.Sprintf("highest_rps_strategy_%s", num)

	// v := strategy.NewHighestRpsStrategy(name, 90, 90, 0)
	v := strategy.NewHighestRpsStrategy(name, "")
	netFunc := func() (float64, float64) {
		v.Net = 1.0
		v.DrawBack = 1.0
		// start := util.ParseDate("2019-04-01") 
		// start := util.ParseDate("2019-04-10")
		// start := util.ParseDate("2019-04-11")

		// start := util.ParseDate("2019-04-06")
		// start := util.ParseDate("2019-09-01")
		start := util.ParseDate("2021-01-03")
		end := start.AddDate(0, 0, 13)
		for i := 0; i < 8; i++ {
			v.Test(util.ToDate(start.Unix()), util.ToDate(end.Unix()),posMax, lossMax)
			// v.Test(util.ToDate(start.AddDate(0, 0, 7).Unix()), util.ToDate(end.AddDate(0, 0, 7).Unix()))
			v.Test(util.ToDate(start.AddDate(0, 0, 15).Unix()), util.ToDate(end.AddDate(0, 0, 15).Unix()),posMax, lossMax)
			// v.Test(util.ToDate(start.AddDate(0, 0, 21).Unix()), util.ToDate(end.AddDate(0, 0, 21).Unix()))
			start = start.AddDate(0, 1, 0)
			end = end.AddDate(0, 1, 0)
		}
		fmt.Println("完成测试")
		return v.Net, v.DrawBack
	}
	netTotals, drawBackTotals := 0.0, 0.0
	count := 10
	for i := 0; i < count; i++ {
		net, drawBack := netFunc()
		netTotals += net
		drawBackTotals += drawBack
	}
	return &testResult{
		num: num,
		net:  netTotals/float64(count),
		drawBack: drawBackTotals/float64(count),
		posMax: posMax,
		lossMax: lossMax,
	}
}