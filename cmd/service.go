/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-04 21:54:16
 */
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.cedric1996.com/go-trader/app"
	"github.cedric1996.com/go-trader/app/service"
	"github.com/urfave/cli"
)

var (
	CmdCount = cli.Command{
		Name:  "count",
		Usage: "fetch spare query count",
		Action: func(c *cli.Context) error {
			app.Init()
			if err := service.GetQueryCount(); err != nil {
				return fmt.Errorf("execute fetch all security cmd fail, please check it: %s", err)
			}
			return nil
		},
	}

	CmdFetch = cli.Command{
		Name:  "price",
		Usage: "fetch price data manually",
		Subcommands: []cli.Command{
			subcmdPriceInit,
			subcmdPriceDaily,
		},
	}
	CmdSecurity = cli.Command{
		Name:   "security",
		Usage:  "fetch all stock securities info and update stock_info table",
		Action: runFetchAllSecurities,
	}

	subcmdPriceInit = cli.Command{
		Name:   "init",
		Usage:  "init stock price and update stock table",
		Action: runStockPriceInit,
	}

	subcmdPriceDaily = cli.Command{
		Name:   "daily",
		Usage:  "fetch daily stock price and update stock table",
		Action: runStockPriceDaily,
	}
)

func runFetchAllSecurities(c *cli.Context) error {
	app.Init()

	if err := service.GetAllSecurities(); err != nil {
		return fmt.Errorf("execute fetch all security cmd fail, please check it: %s", err)
	}
	return nil
}

func runStockPriceDaily(c *cli.Context) error {
	app.Init()
	t := strings.Split(time.Now().Format(time.RFC3339), "T")[0]
	if err := service.FetchStockPriceByDay(t); err != nil {
		return fmt.Errorf("execute fetch daily price fail, please check it: %s", err)
	}
	return nil
}

func runStockPriceInit(c *cli.Context) error {
	app.Init()
	if err := service.InitStockPrice(); err != nil {
		return fmt.Errorf("execute init stock price fail, please check it: %s", err)
	}
	return nil
}
