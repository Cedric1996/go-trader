/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-04 18:43:45
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
		Name:   "vcp",
		Usage:  "vcp tr strategy",
		Action: runVcpTr,
	}
)

func runVcpTr(c *cli.Context) error {
	app.Init()
	v := strategy.NewVcpStrategy()
	if err := v.Run(); err != nil {
		return fmt.Errorf("execute vcp tr strategy fail, please check it", err)
	}
	return nil
}
