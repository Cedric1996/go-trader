/*
 * @Author: cedric.jia
 * @Date: 2021-08-06 15:42:34
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-21 22:55:14
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
			subCmdTaskVerifyRefDate,
			subCmdTaskTrend,
			subCmdTaskEma,
			subCmdTaskHighLowIndex,
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

	subCmdTaskHighLowIndex = cli.Command{
		Name:   "nh_nl",
		Usage:  "new high new low index",
		Action: runHighLowIndexFactor,
	}
)

func runRpsFactor(c *cli.Context) error {
	app.Init()
	t := util.ParseDate("2020-03-18").Unix()
	tradeDays, err := models.GetTradeDay(true, 200, t)
	if err != nil {
		return err
	}
	taskQueue := queue.NewTaskQueue("rps", 20, func(data interface{}) error {
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
	t := util.ParseDate("2021-08-17").Unix()
	tradeDays, err := models.GetTradeDay(true, 4, t)
	if err != nil {
		return err
	}
	taskQueue := queue.NewTaskQueue("highest", 4, func(data interface{}) error {
		date := data.(string)
		f := factor.NewHighestFactor("highest", date, 120, true)
		if err := f.Run(); err != nil {
			return err
		}
		fmt.Printf("highest task has been done, date: %s\n", date)
		return nil
	}, func(dateChan *chan interface{}) {
		for _, day := range tradeDays {
			*dateChan <- day.Date
		}
		fmt.Printf("highest task count: %d\n", len(tradeDays))
	})
	if err := taskQueue.Run(); err != nil {
		return err
	}
	return nil
}

func runVerifyRefDate(c *cli.Context) error {
	app.Init()
	taskQueue := queue.NewTaskQueue("verify_ref_date", 50, func(data interface{}) error {
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
	taskQueue := queue.NewTaskQueue("trend", 50, func(data interface{}) error {
		date := data.(string)
		f := factor.NewTrendFactor(date, 60, 0.95, 0.75, 2.0, 80)
		if err := f.Run(); err != nil {
			return err
		}
		return nil
	}, func(dateChan *chan interface{}) {
		t := util.ParseDate("2021-08-15").Unix()
		tradeDays, err := models.GetTradeDay(true, 300, t)
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

func runEmaFactor(c *cli.Context) error {
	app.Init()
	taskQueue := queue.NewTaskQueue("ema", 20, func(data interface{}) error {
		date := data.(string)
		f := factor.NewEmaFactor(date, 1)
		if err := f.Run(); err != nil {
			return err
		}
		return nil
	}, func(dateChan *chan interface{}) {
		t := util.ParseDate("2021-08-17").Unix()
		tradeDays, err := models.GetTradeDay(true, 800, t)
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

func runHighLowIndexFactor(c *cli.Context) error {
	app.Init()
	f := factor.NewHighLowIndexFactor("nh_nw", "2021-08-20", true)
	if err := f.Run(); err != nil {
		return err
	}
	return nil
}
