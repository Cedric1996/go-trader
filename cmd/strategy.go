/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-06 23:17:17
 */
package cmd

import (
	"fmt"

	"github.cedric1996.com/go-trader/app"
	"github.cedric1996.com/go-trader/app/models"
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
			cli.BoolFlag{Name: "output,o"},
		},
		Action: runHighestRps,
	}
)

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
	// name := "highest_rps_strategy_02"
	// name := "highest_rps_strategy_02"
	name := "highest_rps_strategy_02"

	if err := models.InitStrategyIndexes(name); err != nil {
		return err
	}
	// v := strategy.NewHighestRpsStrategy(name, 90, 90, 0)
	v := strategy.NewHighestRpsStrategy(name, 90, 90, 90)

	// // output := c.Bool("output")
	// // if !output {
	if err := v.Run(); err != nil {
		return fmt.Errorf("execute vcp tr strategy fail, please check it", err)
	}
	// // } else {
	// // if err := v.Output(); err != nil {
	// // 	return err
	// // }
	// v.Kelly()
	netFunc := func() float64 {
		v.Net = 1.0
		start := util.ParseDate("2020-08-30")
		end := start.AddDate(0, 1, -1)
		for i := 0; i < 12; i++ {
			v.Test(util.ToDate(start.Unix()), util.ToDate(end.Unix()))
			start = start.AddDate(0, 1, 0)
			end = end.AddDate(0, 1, 0)
		}
		return v.Net
	}
	// start := util.ParseDate("2020-08-30")
	// end := start.AddDate(0, 1, -1)
	// for i := 0; i < 12; i++ {
	// 	v.Test(util.ToDate(start.Unix()), util.ToDate(end.Unix()))
	// 	start = start.AddDate(0, 1, 0)
	// 	end = end.AddDate(0, 1, 0)
	// }
	netTotals := 0.0
	for i := 0; i < 10; i++ {
		netTotals += netFunc()
	}
	fmt.Printf("总收益: %3f\n", netTotals/10)
	// v.Test("2020-08-30", "2021-08-30")
	// v.Test("2019-08-30", "2021-08-30")
	// }
	return nil
}
