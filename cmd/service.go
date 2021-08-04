/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-07-27 23:21:09
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
	CmdFetch = cli.Command{
		Name:  "fetch",
		Usage: "fetch data manually",
		Subcommands: []cli.Command{
			subcmdAllSecurities,
			subcmdPriceDaily,
		},
	}

	subcmdAllSecurities = cli.Command{
		Name:   "security",
		Usage:  "fetch all stock securities info and update stock_info table",
		Action: runFetchAllSecurities,
	}

	subcmdPriceDaily = cli.Command{
		Name:   "price",
		Usage:  "fetch daily stock price and update stock table",
		Action: runFetchStockPriceDay,
	}
)

func runFetchAllSecurities(c *cli.Context) error {
	app.Init()

	if err := service.GetAllSecurities(); err != nil {
		return fmt.Errorf("execute fetch all security cmd fail, please check it: %s", err)
	}
	return nil
}

func runFetchStockPriceDay(c *cli.Context) error {
	app.Init()
	t := strings.Split(time.Now().Format(time.RFC3339), "T")[0]
	if err := service.FetchStockPriceByDay(t); err != nil {
		return fmt.Errorf("execute fetch daily price fail, please check it: %s", err)
	}
	return nil
}
