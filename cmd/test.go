/*
 * @Author: cedric.jia
 * @Date: 2021-03-13 14:54:23
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-05 12:40:00
 */
package cmd

import (
	"github.cedric1996.com/go-trader/app"
	"github.com/urfave/cli"
)

// CmdWeb represents the available web sub-command.
var (
	CmdTest = cli.Command{
		Name:        "test",
		Usage:       "Test EzTrade cmd",
		Description: `EzTrade Test cmd helps you test basic feature and APIs`,
		Action:      runTest,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "test, t",
				Value: "",
				Usage: "Exec the whole basic test flow",
			},
		},
	}
	CmdServer = cli.Command{
		Name:        "server",
		Usage:       "run EzTrade server",
		Description: `EzTrade Server cmd helps you run EzTrade server`,
		Action:      runServer,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "server, s",
				Value: "",
				Usage: "run EzTrade server",
			},
		},
	}
)

func runTest(c *cli.Context) error {
	app.Init()

	return nil
}

func runServer(c *cli.Context) error {
	app.RunServer()
	return nil
}
