/*
 * @Author: cedric.jia
 * @Date: 2021-08-06 15:42:34
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-08 09:48:03
 */

package cmd

import (
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
		},
	}

	subCmdTaskRps = cli.Command{
		Name:   "rps",
		Usage:  "rps task",
		Action: runRpsFactor,
	}
)

func runRpsFactor(c *cli.Context) error {
	app.Init()
	t := util.ParseDate("2021-08-06").Unix()
	tradeDays, err := models.GetTradeDay(false, 100, t)
	if err != nil {
		return err
	}
	for _, day := range tradeDays {
		rps := factor.NewRpsFactor("rps", 120, 85, day.Date)
		if err := rps.Get(); err != nil {
			return err
		}
		if err := rps.Run(); err != nil {
			return err
		}
	}
	return nil
}
