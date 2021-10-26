/*
 * @Author: cedric.jia
 * @Date: 2021-03-13 14:51:05
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-26 17:25:21
 */

package main

import (
	"log"
	"os"
	"runtime"
	"strings"

	"github.cedric1996.com/go-trader/cmd"
	"github.com/urfave/cli"
)

var (
	Version     = "development"
	Tags        = ""
	MakeVersion = ""
)

func main() {
	app := cli.NewApp()
	app.Name = "EzTrade"
	app.Usage = "A painless self-hosted Quantative service"
	app.Version = Version + formatBuiltWith()
	app.Commands = []cli.Command{
		cmd.CmdTest,
		cmd.CmdServer,
		cmd.CmdFetch,
		cmd.CmdCount,
		cmd.CmdSecurity,
		cmd.CmdIndex,
		cmd.CmdTask,
		cmd.CmdVcp,
		cmd.CmdPosition,
		cmd.CmdStrategy,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Run go-trader fail! %s", err)
	}

}

func formatBuiltWith() string {
	var version = runtime.Version()
	if len(MakeVersion) > 0 {
		version = MakeVersion + ", " + runtime.Version()
	}
	if len(Tags) == 0 {
		return " built with " + version
	}

	return " built with " + version + " : " + strings.ReplaceAll(Tags, " ", ", ")
}
