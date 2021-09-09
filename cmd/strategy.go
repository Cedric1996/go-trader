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
	v := strategy.NewHighestRpsStrategy("highest_rps_strategy_07", date)
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
	v := strategy.NewVcpStrategy("vcp_tr_strategy_02")

	// output := c.Bool("output")
	// if !output {
	if err := v.Run(); err != nil {
		return fmt.Errorf("execute vcp tr strategy fail, please check it", err)
	}
	// } else {
	if err := v.Output(); err != nil {
		return err
	}
	v.Kelly()
	v.Test("2019-06-30", "2021-08-30")
	// }
	return nil
}

func runHighestRps(c *cli.Context) error {
	app.Init()
	// name := "highest_rps_strategy_09"

	// v := strategy.NewHighestRpsStrategy(name, "")
	// run := c.Bool("run")
	// if run {
	// 	if err := models.InitStrategyIndexes(name); err != nil {
	// 		return err
	// 	}
	// 	if err := v.Run(); err != nil {
	// 		return fmt.Errorf("execute vcp tr strategy fail, please check it", err)
	// 	}
	// }

	// nums := []string{"09"}
	nums := []string{"01", "02", "03", "04", "05", "06", "07", "08"}
	res := []*testResult{}
	for _, s := range nums {
		res = append(res, highestRpsTest(s))
	}
	for _, r := range res {
	fmt.Printf("策略序号：%s, 总收益: %3f, 最大月亏损: %3f\n",r.num,  r.net, r.drawBack)
	}

	return nil
}

type testResult struct {
	num string
	net float64
	drawBack float64
}

func highestRpsTest(num string) *testResult{
	name := fmt.Sprintf("highest_rps_strategy_%s", num)

	// v := strategy.NewHighestRpsStrategy(name, 90, 90, 0)
	v := strategy.NewHighestRpsStrategy(name, "")
	netFunc := func() (float64, float64) {
		v.Net = 1.0
		v.DrawBack = 1.0
		// start := util.ParseDate("2019-04-01")
		// start := util.ParseDate("2019-09-01")
		start := util.ParseDate("2020-08-26")

		end := start.AddDate(0, 0, 14)
		for i := 0; i < 24; i++ {
			v.Test(util.ToDate(start.Unix()), util.ToDate(end.Unix()))
			start = start.AddDate(0, 0, 15)
			end = end.AddDate(0, 0, 15)
		}
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
	}
}