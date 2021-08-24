/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-23 23:56:35
 */
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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
			subcmdHighest,
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
		Action: runGetVcp,
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

	subcmdHighest = cli.Command{
		Name:   "highest",
		Usage:  "fetch daily stock price and update stock table",
		Action: runGetHighest,
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
	if err != nil {
		return fmt.Errorf("execute fetch daily price fail, please check it: %s", err)
	}
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
	if err := models.InitTrueRangeTableIndexes(); err != nil {
		return err
	}
	// if err := models.InitRpsIncreaseTableIndexes(); err != nil {
	// 	return err
	// }
	// if err := models.InitEmaTableIndexes(); err != nil {
	// 	return err
	// }
	// if err := models.InitVcpTableIndexes(); err != nil {
	// 	return err
	// }
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
	codes := make([]string, 0)
	for _, vcp := range vcps {
		parts := strings.Split(vcp.RpsBase.Code, ".")
		prefix := "sh"
		if parts[1] == "XSHE" {
			prefix = "sz"
		}
		codes = append(codes, prefix+parts[0])
	}
	result := make(map[string]interface{})
	result["leek-fund.stocks"] = codes

	data, err := json.Marshal(&result)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(".result/result.json", data, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func runGetHighest(c *cli.Context) error {
	app.Init()
	highest, _ := models.GetHighestList(models.SearchOption{Reversed: true, Limit: 1}, "highest")
	lowest, _ := models.GetHighestList(models.SearchOption{Reversed: true, Limit: 1}, "lowest")
	valuation, _ := models.GetValuation(models.SearchOption{Reversed: true, Limit: 1})
	fmt.Println(util.ToDate(highest[0].Timestamp), util.ToDate(lowest[0].Timestamp), util.ToDate(valuation[0].Timestamp))
	return nil
}
