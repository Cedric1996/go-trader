/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-30 11:32:29
 */
package cmd

import (
	"fmt"
	"time"

	"github.cedric1996.com/go-trader/app"
	"github.cedric1996.com/go-trader/app/factor"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
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
			subcmdPriceClean,
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
	CmdVcp = cli.Command{
		Name:   "vcp",
		Usage:  "get vcp",
		Action: runGetWeekVcp,
	}
	CmdPosition = cli.Command{
		Name:  "pos",
		Usage: "calculate portfolio",
		Subcommands: []cli.Command{
			subcmdPositionNew,
			subcmdCalPosition,
		},
	}

	subcmdPriceInit = cli.Command{
		Name:  "init",
		Usage: "init stock price and update stock table",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "date,d"},
		},
		Action: runStockPriceInit,
	}

	subcmdPriceDaily = cli.Command{
		Name:   "daily",
		Usage:  "fetch daily stock price and update stock table",
		Action: runStockPriceDaily,
	}
	subcmdPriceClean = cli.Command{
		Name:  "clean",
		Usage: "clean stock price and related data",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "date,d"},
		},
		Action: runStockPriceClean,
	}

	subcmdPositionNew = cli.Command{
		Name:  "new",
		Usage: "new long or short position",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "file,f"},
		},
		Action: runNewPosition,
	}
	subcmdCalPosition = cli.Command{
		Name:   "cal",
		Usage:  "calculate portfolio",
		Action: runCalPortfolio,
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
	// dates, err := service.FetchStockPriceDayDaily()
	// if err != nil {
	// 	return fmt.Errorf("execute fetch daily price fail, please check it: %s", err)
	// }
	dates := []string{"2021-08-27"}
	for _, day := range dates {
		fmt.Printf("begin init stock price by day: %s\n", day)
		if err := service.InitStockPriceByDay(day); err != nil {
			return err
		}
		t := util.ParseDate(day).Unix()
		dates, err := models.GetTradeDay(true, 1, t)
		if err != nil || len(dates) == 0 || dates[0].Timestamp != t {
			return fmt.Errorf("parse date error: invalid trade date %s", day)
		}
		if err := factor.InitFactorByDate(day); err != nil {
			return err
		}
	}
	// if err := service.VerifyStockPriceDay(); err != nil {
	// 	return err
	// }
	return nil
}

func runStockPriceClean(c *cli.Context) error {
	app.Init()
	d := c.String("date")
	if len(d) == 0 {
		return fmt.Errorf("please specify clean date")
	}
	if err := factor.CleanFactorByDate(d); err != nil {
		return err
	}
	return nil
}

func runStockPriceInit(c *cli.Context) error {
	app.Init()
	d := c.String("date")
	if len(d) == 0 {
		return fmt.Errorf("please specify clean date")
	}
	if err := service.InitStockPriceByDay(d); err != nil {
		return fmt.Errorf("execute init stock price fail, please check it: %s", err)
	}
	if err := factor.InitFactorByDate(d); err != nil {
		return err
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
	if err := models.InitPortfolioIndex(); err != nil {
		return err
	}
	return nil
}

func runGetVcp(c *cli.Context) error {
	app.Init()
	tradeDay, err := models.GetTradeDay(true, 2, util.TodayUnix())
	if err != nil || len(tradeDay) != 2 {
		return nil
	}
	vcps, err := models.GetNewVcpByDate(tradeDay[0].Timestamp, tradeDay[1].Timestamp)
	if err != nil {
		return err
	}
	datas := []string{}
	for _, v := range vcps {
		datas = append(datas, v.RpsBase.Code)
	}
	if err := service.GenerateVcpFile(datas); err != nil {
		return err
	}
	return nil
}

func runGetWeekVcp(c *cli.Context) error {
	app.Init()
	codes, err := service.GetVcpByInterval(util.Today(), 5)
	if err != nil {
		return err
	}
	datas := []string{}
	for k, v := range codes {
		if v >= 4 {
			datas = append(datas, k)
		}
	}
	if err := service.GenerateVcpFile(datas); err != nil {
		return err
	}
	return nil
}

func runCalPortfolio(c *cli.Context) error {
	app.Init()
	if err := service.GetPortfolio("portfolio"); err != nil {
		return err
	}
	return nil
}

func runNewPosition(c *cli.Context) error {
	app.Init()
	d := c.String("file")
	if len(d) == 0 {
		d = "long.json"
	}
	if err := service.NewPosition(d); err != nil {
		return err
	}
	return nil
}

func initPortfolio(c *cli.Context) error {
	app.Init()
	if err := models.InsertPortfolio(&models.Portfolio{
		Risk:      0.0,
		Inventory: 0.0,
		Available: 1200000.0,
		IsCurrent: true,
		Timestamp: time.Now().Unix(),
	}); err != nil {
		return err
	}
	return nil
}
