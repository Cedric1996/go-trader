/*
 * @Author: cedric.jia
 * @Date: 2021-07-27 23:13:32
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-07-27 23:21:09
 */
package cmd

import (
	"fmt"

	"github.cedric1996.com/go-trader/app"
	"github.cedric1996.com/go-trader/app/service"
	"github.com/urfave/cli"
)

var (
	CmdFetch = cli.Command{
		Name: "fetch",
		Usage: "fetch data manually",
		Subcommands: []cli.Command{
			subcmdAllSecurities,
		},
	}

	subcmdAllSecurities = cli.Command{
		Name: "security",
		Usage: "fetch all stock securities info and update stock_info table",
		Action: runFetchAllSecurities,
	}
)

func runFetchAllSecurities(c *cli.Context) error {
	app.Init()

	if err := service.GetAllSecurities(); err != nil {
		return fmt.Errorf("execute fetch all security cmd fail, please check it")
	}
	return nil
}