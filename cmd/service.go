/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-06 14:29:56
 */
package cmd

import (
	"fmt"

	"github.cedric1996.com/go-trader/app"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/service"
	"github.com/urfave/cli"
)

var (
	CmdCount = cli.Command{
		Name:   "count",
		Usage:  "fetch spare query count",
		Action: runCount,
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
	CmdIndex = cli.Command{
		Name:   "index",
		Usage:  "init mongodb table indexes",
		Action: runInitIndex,
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
	if err := service.FetchStockPriceDayDaily(); err != nil {
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

func runCount(c *cli.Context) error {
	app.Init()
	if err := service.GetQueryCount(); err != nil {
		return fmt.Errorf("execute fetch all security cmd fail, please check it: %s", err)
	}
	return nil
}

func runInitIndex(c *cli.Context) error {
	app.Init()
	if err := models.InitRpsTableIndexes(); err != nil {
		return err
	}
	// if err := models.InitStockTableIndexes(); err != nil {
	// 	return err
	// }
	return nil
}
