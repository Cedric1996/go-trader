/*
 * @Author: cedric.jia
 * @Date: 2021-08-06 15:42:34
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-29 00:12:26
 */

package cmd

import (
	"fmt"
	"time"

	"github.cedric1996.com/go-trader/app"
	"github.cedric1996.com/go-trader/app/factor"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
	"github.com/urfave/cli"
)

var (
	CmdTask = cli.Command{
		Name:  "task",
		Usage: "execute calculate task",
		Subcommands: []cli.Command{
			subCmdTaskRps,
			subCmdTaskGetRps,
			subCmdTaskHighest,
			subCmdTaskLowest,
			subCmdTaskVerifyRefDate,
			subCmdTaskTrend,
			subCmdTaskEma,
			subCmdTaskMa,
			subCmdTaskHighLowIndex,
			subCmdTaskTrueRange,
			subCmdTaskModule,
		},
	}

	subCmdTaskRps = cli.Command{
		Name:   "rps",
		Usage:  "rps task",
		Action: runRpsFactor,
	}

	subCmdTaskGetRps = cli.Command{
		Name:   "get_rps",
		Usage:  "get rps task",
		Action: runGetRps,
	}

	subCmdTaskHighest = cli.Command{
		Name:   "highest",
		Usage:  "highest tasks",
		Action: runHighestFactor,
	}

	subCmdTaskLowest = cli.Command{
		Name:   "lowest",
		Usage:  "lowest tasks",
		Action: runLowestFactor,
	}

	subCmdTaskVerifyRefDate = cli.Command{
		Name:   "verify_ref_date",
		Usage:  "verify ref date task",
		Action: runVerifyRefDate,
	}

	subCmdTaskTrend = cli.Command{
		Name:   "trend",
		Usage:  "trend tasks",
		Action: runTrendFactor,
	}

	subCmdTaskEma = cli.Command{
		Name:   "ema",
		Usage:  "moving average tasks",
		Action: runEmaFactor,
	}

	subCmdTaskMa = cli.Command{
		Name:   "ma",
		Usage:  "moving average tasks",
		Action: runMaFactor,
	}

	subCmdTaskHighLowIndex = cli.Command{
		Name:   "nh_nl",
		Usage:  "new high new low index",
		Action: runHighLowIndexFactor,
	}

	subCmdTaskTrueRange = cli.Command{
		Name:   "tr",
		Usage:  "true range and average true range",
		Action: runTrueRangeFactor,
	}

	subCmdTaskModule = cli.Command{
		Name:   "module",
		Usage:  "init module concept",
		Action: runInitModule,
	}
)

func runRpsFactor(c *cli.Context) error {
	app.Init()
	// if err := models.DropCollection("rps"); err != nil {
	// 	return fmt.Errorf("drop collection: %s", err)
	// }
	// if err := models.DropCollection("rps_increase"); err != nil {
	// 	return fmt.Errorf("drop collection: %s", err)
	// }
	// if err := models.InitRpsTableIndexes(); err != nil {
	// 	return err
	// }
	// if err := models.InitRpsIncreaseTableIndexes(); err != nil {
	// 	return err
	// }

	// t := util.TodayUnix()
	t := util.ParseDate("2020-07-29").Unix()

	tradeDays, err := models.GetTradeDay(true, 250, t)
	if err != nil {
		return err
	}
	taskQueue := queue.NewTaskQueue("rps", 50, func(data interface{}) error {
		date := data.(string)
		rps := factor.NewRpsFactor("rps", 120, 85, date)
		if err := rps.Run(); err != nil {
			return err
		}
		fmt.Printf("rps task has been done, date: %s\n", date)
		return nil
	}, func(dateChan *chan interface{}) {
		for _, day := range tradeDays {
			*dateChan <- day.Date
		}
	})
	if err := taskQueue.Run(); err != nil {
		return err
	}
	return nil
}

func runGetRps(c *cli.Context) error {
	startT := time.Now()
	app.Init()
	// if err := models.DeleteRpsIncrease(util.ParseDate("2020-03-20").Unix()); err != nil {
	// 	return err
	// }
	results, err := models.GetRpsIncrease(models.SearchOption{
		Code: "601952.XSHG",
		// Timestamp: t,
	})
	if err != nil {
		return err
	}
	fmt.Printf("get trade day count: %d\n", len(results))
	fmt.Printf("task rps finished successfully, total time: %s", time.Since(startT).String())
	return nil
}

func runHighestFactor(c *cli.Context) error {
	app.Init()
	models.DropCollection("highest_120")
	models.DropCollection("loweset_120")
	models.InitHighestTableIndexes("120")
	taskQueue := queue.NewTaskQueue("highest", 50, func(data interface{}) error {
		code := data.(string)
		f := factor.NewHighestFactor("highest_120", "2021-10-26", 120)
		if err := f.Init(code); err != nil {
			return err
		}
		return nil
	}, func(dateChan *chan interface{}) {
		stocks, _ := models.GetAllSecurities()
		for _, stock := range stocks {
			*dateChan <- stock.Code
		}
	})
	if err := taskQueue.Run(); err != nil {
		return err
	}
	return nil
}

func runLowestFactor(c *cli.Context) error {
	app.Init()
	models.DropCollection("lowest_rps")
	models.InitLowestRpsTableIndexes()
	f := factor.NewLowestRpsFactor("2021-10-26")
	if err := f.Run(); err != nil {
		return err
	}
	// taskQueue := queue.NewTaskQueue("lowest", 50, func(data interface{}) error {
	// 	code := data.(string)
	// 	f := factor.NewLowestRpsFactor("2021-10-26")
	// 	if err := f.Run(); err != nil {
	// 		return err
	// 	}
	// 	return nil
	// }, func(dateChan *chan interface{}) {
	// 	stocks, _ := models.GetAllSecurities()
	// 	for _, stock := range stocks {
	// 		*dateChan <- stock.Code
	// 	}
	// })
	// if err := taskQueue.Run(); err != nil {
	// 	return err
	// }
	return nil
}

func runVerifyRefDate(c *cli.Context) error {
	app.Init()
	taskQueue := queue.NewTaskQueue("verify_ref_date", 30, func(data interface{}) error {
		code := data.(string)
		if err := service.VerifyRefDate(code); err != nil {
			return err
		}
		return nil
	}, func(dateChan *chan interface{}) {
		for code, _ := range service.SecuritySet {
			*dateChan <- code
		}
	})
	if err := taskQueue.Run(); err != nil {
		return err
	}
	return nil
}

func runTrendFactor(c *cli.Context) error {
	app.Init()
	if err := models.DropCollection("highest_approach"); err != nil {
		return fmt.Errorf("drop collection: %s", err)
	}
	if err := models.InittHighestApproachTableIndexes(); err != nil {
		return err
	}
	f := factor.NewTrendFactor("2021-10-26", 0, 0.85, 0, 0)
	if err := f.Run(); err != nil {
		return err
	}
	return nil
}

func runEmaFactor(c *cli.Context) error {
	app.Init()
	taskQueue := queue.NewTaskQueue("ema", 1, func(data interface{}) error {
		date := data.(string)
		f := factor.NewEmaFactor(date, 1000)
		if err := f.Run(); err != nil {
			return err
		}
		return nil
	}, func(dateChan *chan interface{}) {
		t := util.ParseDate("2021-09-02").Unix()
		tradeDays, err := models.GetTradeDay(true, 1, t)
		if err != nil {
			return
		}
		for _, date := range tradeDays {
			*dateChan <- date.Date
		}
	})
	if err := taskQueue.Run(); err != nil {
		return err
	}
	return nil
}

func runMaFactor(c *cli.Context) error {
	app.Init()
	// if err:= models.InitEmaTableIndexes("ma");err!=nil {
	// 	return err
	// }
	f := factor.NewMaFactor("2019-01-01", 700)
	// if err := f.Run(); err != nil {
	// 	return err
	// }
	if err := f.Output(); err != nil {
		return err
	}
	return nil
}

func runHighLowIndexFactor(c *cli.Context) error {
	app.Init()
	f := factor.NewHighLowIndexFactor("nh_nw", "2020-06-01")
	if err := f.Run(); err != nil {
		return err
	}
	return nil
}

func runTrueRangeFactor(c *cli.Context) error {
	app.Init()
	f := factor.NewTrueRangeFactor("2021-08-24", 13)
	if err := f.InitByCode(); err != nil {
		return err
	}
	return nil
}

func runInitModule(c *cli.Context) error {
	app.Init()
	if err := models.DropCollection("stock_module"); err != nil {
		return fmt.Errorf("drop collection: %s", err)
	}
	if err := models.InitStockModuleIndexes(); err != nil {
		return err
	}
	taskQueue := queue.NewTaskQueue("module concept", 20, func(data interface{}) error {
		mod := data.(models.Module)
		stocks, err := service.GetModulesDetail(mod)
		if err != nil {
			return err
		}
		if err := models.InsertStockModule(stocks); err != nil {
			return err
		}
		return nil
	}, func(dateChan *chan interface{}) {
		modules, err := service.GetModuleList("concept", "")
		if err != nil {
			return
		}
		for _, mod := range modules {
			*dateChan <- mod
		}
	})
	if err := taskQueue.Run(); err != nil {
		return err
	}
	return nil
}
