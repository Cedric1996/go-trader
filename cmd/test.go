/*
 * @Author: cedric.jia
 * @Date: 2021-03-13 14:54:23
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-03-14 13:09:23
 */
package cmd

import (
	"fmt"

	"github.cedric1996.com/eztrader/app"
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
)

func runTest(c *cli.Context) error {
	fmt.Println("test cmd")
	app.Init()

	return nil
}
