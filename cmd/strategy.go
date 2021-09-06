/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-06 10:47:54
 */
package cmd

import (
	"fmt"

	"github.cedric1996.com/go-trader/app"
	"github.cedric1996.com/go-trader/app/strategy"
	"github.com/urfave/cli"
)

var (
	CmdStrategy = cli.Command{
		Name:  "strategy",
		Usage: "cal strategy",
		Subcommands: []cli.Command{
			subcmdVcpTr,
			subcmdPriceDaily,
			subcmdPriceClean,
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
