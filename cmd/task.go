/*
 * @Author: cedric.jia
 * @Date: 2021-08-06 15:42:34
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-08 09:48:03
 */

package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.cedric1996.com/go-trader/app"
	"github.cedric1996.com/go-trader/app/factor"
	"github.cedric1996.com/go-trader/app/models"
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
)

func runRpsFactor(c *cli.Context) error {
	startT := time.Now()
	app.Init()
	t := util.ParseDate("2020-03-24").Unix()
	tradeDays, err := models.GetTradeDay(true, 40, t)
	if err != nil {
		return err
	}

	dateChan := make(chan string, 20)
	task := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		task.Add(1)
		go func() error {
			for date := range dateChan {
				rps := factor.NewRpsFactor("rps", 120, 85, date)
				if err := rps.Get(); err != nil {
					return err
				}
				if err := rps.Run(); err != nil {
					return err
				}
				fmt.Printf("rps increase task has been done, date: %s\n", date)
			}
			task.Done()
			return nil
		}()
	}

	for _, day := range tradeDays {
		dateChan <- day.Date
	}
	close(dateChan)
	task.Wait()
	fmt.Printf("task rps finished successfully, total time: %s", time.Since(startT).String())
	return nil
}

func runGetRps(c *cli.Context) error {
	app.Init()
	// if err := models.DeleteRpsIncrease(util.ParseDate("2021-06-03").Unix()); err != nil {
	// 	return err
	// }
	// if err := models.DeleteRpsIncrease(util.ParseDate("2021-06-14").Unix()); err != nil {
	// 	return err
	// }
	rps := factor.NewRpsFactor("rps", 120, 85, "2021-08-09")
	if err := rps.Calculate(); err != nil {
		return err
	}
	// results, err := models.GetRpsIncrease(models.RpsOption{
	// 	Code: "601952.XSHG",
	// 	// Timestamp: t,
	// })
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(len(results))
	return nil
}
