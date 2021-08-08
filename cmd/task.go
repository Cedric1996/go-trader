/*
 * @Author: cedric.jia
 * @Date: 2021-08-06 15:42:34
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-06 15:51:24
 */

package cmd

import (
	"github.cedric1996.com/go-trader/app"
	"github.cedric1996.com/go-trader/app/factor"
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
	rps := factor.NewRpsFactor("rps", 120, 85, "2021-08-05")
	if err := rps.Get(); err != nil {
		return err
	}
	return nil
}
