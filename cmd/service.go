/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-26 17:25:19
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
			subcmdPriceHour,
			subcmdPriceWeek,
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
		Name:  "trend",
		Usage: "get vcp",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "date,d"},
		},
		Action: runGetNewHighestRps,
	}
	CmdPosition = cli.Command{
		Name:  "pos",
		Usage: "calculate portfolio",
		Subcommands: []cli.Command{
			subcmdPositionNew,
			subcmdCalPosition,
			subcmdHoldPosition,
			subcmdPosHighestRps,
			subcmdPosVcp,
		},
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

	subcmdPriceHour = cli.Command{
		Name:   "hour",
		Usage:  "fetch hour stock price and update stock table",
		Action: runStockPriceHour,
	}

	subcmdPriceWeek = cli.Command{
		Name:   "week",
		Usage:  "fetch week stock price and update stock table",
		Action: runStockPriceWeek,
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
			cli.StringFlag{Name: "type,t"},
		},
		Action: runNewPosition,
	}
	subcmdCalPosition = cli.Command{
		Name:  "cal",
		Usage: "calculate portfolio",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "file,f"},
			cli.StringFlag{Name: "date,d"},
		},
		Action: runCalPortfolio,
	}
	subcmdHoldPosition = cli.Command{
		Name:   "hold",
		Usage:  "output holding positions",
		Action: runHoldPosition,
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
	dates, err := service.FetchStockPriceDayDaily()
	if err != nil || len(dates) == 0 {
		return fmt.Errorf("execute fetch daily price fail, please check it: %s", err)
	}
	if err := service.InitStockPriceByDay(dates); err != nil {
		return err
	}
	// t := util.ParseDate(day).Unix()
	// dates, err := models.GetTradeDay(true, 1, t)
	// if err != nil || len(dates) == 0 || dates[0].Timestamp != t {
	// 	return fmt.Errorf("parse date error: invalid trade date %s", day)
	// }
	// if err := factor.InitFactorByDate(day); err != nil {
	// 	return err
	// }
	// if err := service.VerifyStockPriceDay(); err != nil {
	// 	return err
	// }
	return nil
}

func runStockPriceWeek(c *cli.Context) error {
	app.Init()
	prices, err := service.GetPricesByWeek("300725.XSHE", "2021-10-08", 10)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(prices[0])
	return nil
}

func runStockPriceHour(c *cli.Context) error {
	app.Init()
	prices, err := service.GetPricesByHour("300725.XSHE", "2021-09-24", 8)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(prices[0])
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
	if err := service.InitStockSecurity(); err != nil {
		return fmt.Errorf("execute init stock security fail, please check it: %s", err)
	}
	if err := models.PruneStockSecurity(); err != nil {
		return fmt.Errorf("execute init stock security fail, please check it: %s", err)
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
	// if err := models.InitStrategyIndexes("highest_rps_strategy"); err != nil {
	// 	return err
	// }
	if err := models.InitStockTableIndexes(); err != nil {
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

func runGetNewHighestRps(c *cli.Context) error {
	app.Init()
	// date := c.String("date")
	// var unix int64
	// if len(date) == 0 {
	// 	unix = time.Now().AddDate(0, 0, -1).Unix()
	// } else {
	// 	unix = util.ParseDate(date).Unix()
	// }
	// m, err := service.GetNewRps(unix)
	// if err != nil {
	// 	return err
	// }
	// datas := []string{}
	// for k, _ := range *m {
	// 	datas = append(datas, k)
	// }
	// if err := service.GenerateVcpFile(datas); err != nil {
	// 	return err
	// }
	if err := service.ExportHighestApproach(); err != nil {
		return err
	}
	return nil
}

func runCalPortfolio(c *cli.Context) error {
	app.Init()
	filename := c.String("file")
	date := c.String("date")
	var unix int64
	if len(filename) == 0 {
		filename = "portfolio.json"
	}
	if len(date) == 0 {
		unix = time.Now().AddDate(0, 0, -1).Unix()
	} else {
		unix = util.ParseDate(date).Unix()
	}
	if err := service.GetPortfolio(filename, unix); err != nil {
		return err
	}
	return nil
}

func runHoldPosition(c *cli.Context) error {
	app.Init()
	if err := service.GetPositionSignal(); err != nil {
		return err
	}
	return nil
}

func runNewPosition(c *cli.Context) error {
	app.Init()
	d := c.String("file")
	flag := c.Bool("type")
	if len(d) == 0 {
		d = "long.json"
	}
	if err := service.NewPosition(d, flag); err != nil {
		return err
	}
	return nil
}

func initPortfolio(c *cli.Context) error {
	app.Init()
	if err := models.InsertPortfolio(&models.Portfolio{
		Risk:      0.0,
		Inventory: 0.0,
		Available: 851.80,
		IsCurrent: true,
		Timestamp: time.Now().Unix(),
	}); err != nil {
		return err
	}
	return nil
}

func initPosition(c *cli.Context) error {
	app.Init()
	if err := models.InsertPortfolio(&models.Portfolio{
		Risk:      0.0,
		Inventory: 0.0,
		Available: 851.80,
		IsCurrent: true,
		Timestamp: time.Now().Unix(),
	}); err != nil {
		return err
	}

	positions := []interface{}{
		models.Position{
			Code:      "600760.XSHG",
			Volume:    5976,
			BeginAt:   util.ParseDate("2021-08-23").Unix(),
			EndAt:     util.MaxInt(),
			DealPrice: 72.980,
			LossPrice: 71.0,
		},
		models.Position{
			Code:      "600096.XSHG",
			Volume:    10600,
			EndAt:     util.MaxInt(),
			BeginAt:   util.ParseDate("2021-08-25").Unix(),
			DealPrice: 19.241,
			LossPrice: 18.5,
		},
		models.Position{
			Code:      "300316.XSHE",
			Volume:    2900,
			EndAt:     util.MaxInt(),
			BeginAt:   util.ParseDate("2021-08-26").Unix(),
			DealPrice: 69.543,
			LossPrice: 65.1,
		},
		models.Position{
			Code:      "600623.XSHG",
			Volume:    12400,
			BeginAt:   util.ParseDate("2021-08-30").Unix(),
			EndAt:     util.MaxInt(),
			DealPrice: 12.627,
			LossPrice: 12.16,
		},
		models.Position{
			Code:      "300587.XSHE",
			Volume:    7200,
			BeginAt:   util.ParseDate("2021-08-30").Unix(),
			EndAt:     util.MaxInt(),
			DealPrice: 21.860,
			LossPrice: 20.0,
		},
	}
	if err := models.InsertPositions(positions); err != nil {
		return err
	}
	return nil
}
