/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-17 23:03:28
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
		},
	}

	CmdFundamental = cli.Command{
		Name:   "fundamental",
		Usage:  "fetchfundamental data",
		Action: runFundamentalInit,
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
	dates, err := service.FetchStockPriceDayDaily()
	if err != nil {
		return fmt.Errorf("execute fetch daily price fail, please check it: %s", err)
	}
	for _, day := range dates {
		rps := factor.NewRpsFactor("rps", 120, 85, day)
		if err := rps.Run(); err != nil {
			return err
		}
		highest := factor.NewHighestFactor("highest", day, 120, false)
		if err := highest.Run(); err != nil {
			return err
		}
		lowest := factor.NewHighestFactor("lowest", day, 120, true)
		if err := lowest.Run(); err != nil {
			return err
		}
		if err := service.InitFundamental(day); err != nil {
			return err
		}
		trend := factor.NewTrendFactor(day, 60, 0.95, 0.75, 2.0, 80)
		if err := trend.Run(); err != nil {
			return err
		}
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

func runFundamentalInit(c *cli.Context) error {
	app.Init()
	if err := service.InitFundamental("valuation"); err != nil {
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
	// if err := models.InitHighestTableIndexes(); err != nil {
	// 	return err
	// }
	// if err := models.InitStockTableIndexes(); err != nil {
	// 	return err
	// }
	if err := models.InitFundamentalIndexes(); err != nil {
		return err
	}

	if err := models.DeleteFundamental(1611846000); err != nil {
		return err
	}
	// if err := models.InitStockTableIndexes(); err != nil {
	// 	return err
	// }
	return nil
}

func runGetVcp(c *cli.Context) error {
	app.Init()
	tradeDay, err := models.GetTradeDay(true, 1, util.TodayUnix())
	if err != nil || len(tradeDay) == 0 {
		return nil
	}
	vcps, err := models.GetNewVcpByDate(tradeDay[0].Timestamp)
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
	// for _, vcp := range vcps {
	// 	fmt.Println(vcp.RpsBase.Code)
	// }

	return nil
}
